// Package eventstore persists automatic reply events asynchronously.
package eventstore

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sxyazi/bendan/commands"
)

const defaultQueueSize = 1024
const defaultBatchSize = 50
const defaultFlushInterval = time.Second

// Store implements commands.AutomaticReplyRecorder with a bounded, non-blocking
// queue. Database failures never delay or disable bot replies.
type Store struct {
	pool  *pgxpool.Pool
	queue chan commands.AutomaticReplyEvent
}

// New connects to PostgreSQL and creates the event table and indexes.
func New(ctx context.Context, databaseURL string) (*Store, error) {
	if databaseURL == "" {
		return nil, errors.New("database URL is empty")
	}
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}
	config.MaxConns = 3
	config.MinConns = 0
	config.MaxConnIdleTime = 5 * time.Minute
	config.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	setupCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(setupCtx); err != nil {
		pool.Close()
		return nil, err
	}
	if _, err := pool.Exec(setupCtx, schemaSQL); err != nil {
		pool.Close()
		return nil, err
	}
	return &Store{pool: pool, queue: make(chan commands.AutomaticReplyEvent, defaultQueueSize)}, nil
}

// Record enqueues an event without blocking the message handling goroutine.
func (s *Store) Record(event commands.AutomaticReplyEvent) {
	if s == nil {
		return
	}
	select {
	case s.queue <- event:
	default:
		log.Printf("automatic reply event dropped handler=%s reason=queue_full", event.Handler)
	}
}

// Run writes batches until ctx is cancelled, then performs one final flush.
func (s *Store) Run(ctx context.Context) {
	if s == nil {
		return
	}
	ticker := time.NewTicker(defaultFlushInterval)
	defer ticker.Stop()

	batch := make([]commands.AutomaticReplyEvent, 0, defaultBatchSize)
	for {
		select {
		case event := <-s.queue:
			batch = append(batch, event)
			if len(batch) >= defaultBatchSize {
				s.flush(ctx, batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			batch = s.drain(batch)
			s.flush(ctx, batch)
			batch = batch[:0]
		case <-ctx.Done():
			batch = s.drain(batch)
			flushCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			s.flush(flushCtx, batch)
			cancel()
			return
		}
	}
}

func (s *Store) drain(batch []commands.AutomaticReplyEvent) []commands.AutomaticReplyEvent {
	for len(batch) < defaultBatchSize {
		select {
		case event := <-s.queue:
			batch = append(batch, event)
		default:
			return batch
		}
	}
	return batch
}

func (s *Store) flush(parent context.Context, batch []commands.AutomaticReplyEvent) {
	if len(batch) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(parent, 3*time.Second)
	defer cancel()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Printf("automatic reply event batch failed count=%d error=%v", len(batch), err)
		return
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	for _, event := range batch {
		metadata, _ := json.Marshal(map[string]string{"delivery": event.Delivery})
		_, err = tx.Exec(ctx, `
			INSERT INTO bendan_auto_reply_events
			(handler, chat_type, chat_hash, sender_hash, input_text, reply_text, metadata)
			VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), $5, $6, $7::jsonb)`,
			event.Handler, event.ChatKind, event.ChatHash, event.SenderHash,
			event.InputText, event.ReplyText, metadata,
		)
		if err != nil {
			log.Printf("automatic reply event batch failed count=%d error=%v", len(batch), err)
			return
		}
	}
	if err := tx.Commit(ctx); err != nil {
		log.Printf("automatic reply event batch failed count=%d error=%v", len(batch), err)
		return
	}
	log.Printf("automatic reply events stored count=%d", len(batch))
}

// Close releases database connections after Run has stopped.
func (s *Store) Close() {
	if s != nil {
		s.pool.Close()
	}
}

const schemaSQL = `
CREATE TABLE IF NOT EXISTS bendan_auto_reply_events (
    id          BIGSERIAL PRIMARY KEY,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    handler     TEXT NOT NULL,
    chat_type   TEXT NOT NULL DEFAULT '',
    chat_hash   TEXT,
    sender_hash TEXT,
    input_text  TEXT NOT NULL,
    reply_text  TEXT NOT NULL,
    metadata    JSONB NOT NULL DEFAULT '{}'::jsonb
);
CREATE INDEX IF NOT EXISTS idx_bendan_auto_reply_events_time
    ON bendan_auto_reply_events (occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_bendan_auto_reply_events_handler_time
    ON bendan_auto_reply_events (handler, occurred_at DESC);
`
