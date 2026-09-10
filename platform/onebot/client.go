package onebot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sxyazi/bendan/platform"
)

// Client maintains a OneBot v11 forward WebSocket connection to NapCat.
type Client struct {
	endpoint    string
	accessToken string

	mu      sync.RWMutex
	conn    *websocket.Conn
	self    platform.User
	pending map[string]chan response
	writeMu sync.Mutex
}

type response struct {
	Status  string          `json:"status"`
	RetCode int             `json:"retcode"`
	Data    json.RawMessage `json:"data"`
	Echo    string          `json:"echo"`
}

type action struct {
	Action string `json:"action"`
	Params any    `json:"params"`
	Echo   string `json:"echo"`
}

// NewClient creates a client for NapCat's forward WebSocket endpoint.
func NewClient(endpoint, accessToken string) *Client {
	return &Client{endpoint: endpoint, accessToken: accessToken, pending: make(map[string]chan response)}
}

func (c *Client) Identity() platform.User {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.self
}

func (*Client) Capabilities() platform.Capabilities {
	return platform.Capabilities{CanDeleteMessage: true}
}

// Run reconnects until ctx is cancelled and dispatches OneBot message events.
func (c *Client) Run(ctx context.Context, handle func(context.Context, *platform.Message)) error {
	backoff := time.Second
	for {
		if err := ctx.Err(); err != nil {
			return nil
		}
		conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.endpoint, c.headers())
		if err != nil {
			if !wait(ctx, backoff) {
				return nil
			}
			backoff = minDuration(backoff*2, 30*time.Second)
			continue
		}

		backoff = time.Second
		c.setConnection(conn)
		err = c.read(ctx, conn, handle)
		c.clearConnection(conn)
		_ = conn.Close()
		if ctx.Err() != nil {
			return nil
		}
		if err != nil && !wait(ctx, backoff) {
			return nil
		}
	}
}

func (c *Client) headers() http.Header {
	headers := make(http.Header)
	if c.accessToken != "" {
		headers.Set("Authorization", "Bearer "+c.accessToken)
	}
	return headers
}

func (c *Client) read(ctx context.Context, conn *websocket.Conn, handle func(context.Context, *platform.Message)) error {
	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		var envelope struct {
			Echo string `json:"echo"`
		}
		if err := json.Unmarshal(payload, &envelope); err != nil {
			continue
		}
		if envelope.Echo != "" {
			var result response
			if json.Unmarshal(payload, &result) == nil {
				c.resolve(result)
			}
			continue
		}

		var event Event
		if json.Unmarshal(payload, &event) != nil {
			continue
		}
		c.setSelfID(event.SelfID)
		if message := event.ToPlatformMessage(); message != nil {
			go handle(ctx, message)
		}
	}
}

func (c *Client) SendText(ctx context.Context, chat platform.Chat, text string) (*platform.Message, error) {
	chatID, err := strconv.ParseInt(chat.ID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid OneBot chat ID %q: %w", chat.ID, err)
	}
	params := map[string]any{"message": text}
	if chat.Kind == "private" {
		params["user_id"] = chatID
	} else {
		params["group_id"] = chatID
	}
	return c.send(ctx, "send_msg", params, chat)
}

func (c *Client) ReplyText(ctx context.Context, replyTo *platform.Message, text string) (*platform.Message, error) {
	if replyTo == nil {
		return nil, fmt.Errorf("reply target is required")
	}
	messageID, err := strconv.ParseInt(replyTo.ID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid OneBot message ID %q: %w", replyTo.ID, err)
	}
	return c.SendText(ctx, replyTo.Chat, fmt.Sprintf("[CQ:reply,id=%d]%s", messageID, text))
}

func (c *Client) DeleteMessage(ctx context.Context, message *platform.Message) error {
	if message == nil {
		return fmt.Errorf("message is required")
	}
	messageID, err := strconv.ParseInt(message.ID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid OneBot message ID %q: %w", message.ID, err)
	}
	_, err = c.call(ctx, "delete_msg", map[string]any{"message_id": messageID})
	return err
}

func (*Client) PinMessage(context.Context, *platform.Message) error {
	return platform.ErrUnsupported
}

func (*Client) UnpinMessage(context.Context, platform.Chat, string) error {
	return platform.ErrUnsupported
}

func (c *Client) send(ctx context.Context, name string, params map[string]any, chat platform.Chat) (*platform.Message, error) {
	data, err := c.call(ctx, name, params)
	if err != nil {
		return nil, err
	}
	var result struct {
		MessageID int64 `json:"message_id"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("decode %s response: %w", name, err)
	}
	return &platform.Message{ID: strconv.FormatInt(result.MessageID, 10), Chat: chat, Sender: c.Identity()}, nil
}

func (c *Client) call(ctx context.Context, name string, params any) (json.RawMessage, error) {
	echo := strconv.FormatInt(time.Now().UnixNano(), 10)
	result := make(chan response, 1)
	c.mu.Lock()
	conn := c.conn
	if conn != nil {
		c.pending[echo] = result
	}
	c.mu.Unlock()
	if conn == nil {
		return nil, fmt.Errorf("OneBot is not connected")
	}

	payload, err := json.Marshal(action{Action: name, Params: params, Echo: echo})
	if err != nil {
		c.removePending(echo)
		return nil, err
	}
	c.writeMu.Lock()
	err = conn.WriteMessage(websocket.TextMessage, payload)
	c.writeMu.Unlock()
	if err != nil {
		c.removePending(echo)
		return nil, err
	}

	select {
	case response := <-result:
		if response.Status != "ok" || response.RetCode != 0 {
			return nil, fmt.Errorf("OneBot action %s failed: status=%s retcode=%d", name, response.Status, response.RetCode)
		}
		return response.Data, nil
	case <-ctx.Done():
		c.removePending(echo)
		return nil, ctx.Err()
	}
}

func (c *Client) setConnection(conn *websocket.Conn) {
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
}

func (c *Client) clearConnection(conn *websocket.Conn) {
	c.mu.Lock()
	if c.conn == conn {
		c.conn = nil
	}
	pending := c.pending
	c.pending = make(map[string]chan response)
	c.mu.Unlock()
	for _, result := range pending {
		close(result)
	}
}

func (c *Client) resolve(result response) {
	c.mu.Lock()
	waiter := c.pending[result.Echo]
	delete(c.pending, result.Echo)
	c.mu.Unlock()
	if waiter != nil {
		waiter <- result
	}
}

func (c *Client) removePending(echo string) {
	c.mu.Lock()
	delete(c.pending, echo)
	c.mu.Unlock()
}

func (c *Client) setSelfID(id int64) {
	if id == 0 {
		return
	}
	c.mu.Lock()
	c.self = platform.User{ID: strconv.FormatInt(id, 10), DisplayName: "Bendan"}
	c.mu.Unlock()
}

func wait(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
