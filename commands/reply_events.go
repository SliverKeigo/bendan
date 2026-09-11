package commands

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/sxyazi/bendan/platform"
)

type automaticReplyHandlerContextKey struct{}

// AutomaticReplyEvent is the privacy-preserving record written after an
// automatic reply has been sent successfully.
type AutomaticReplyEvent struct {
	Handler    string
	ChatKind   string
	ChatHash   string
	SenderHash string
	InputText  string
	ReplyText  string
	Delivery   string
}

// AutomaticReplyRecorder accepts automatic reply events without blocking the
// message path. Implementations must be safe for concurrent use.
type AutomaticReplyRecorder interface {
	Record(AutomaticReplyEvent)
}

var automaticReplyRecorder AutomaticReplyRecorder
var automaticReplyHashSalt string

// SetAutomaticReplyRecorder configures automatic-reply persistence. Passing
// nil disables it. Raw chat and sender IDs are never handed to the recorder.
func SetAutomaticReplyRecorder(recorder AutomaticReplyRecorder, hashSalt string) {
	automaticReplyRecorder = recorder
	automaticReplyHashSalt = hashSalt
}

func recordAutomaticReply(ctx context.Context, input *platform.Message, replyText, delivery string) {
	recorder := automaticReplyRecorder
	handler, automatic := ctx.Value(automaticReplyHandlerContextKey{}).(string)
	if !automatic || handler == "" || recorder == nil || input == nil {
		return
	}

	event := AutomaticReplyEvent{
		Handler:    handler,
		ChatKind:   input.Chat.Kind,
		ChatHash:   hashIdentifier("chat", input.Chat.ID),
		SenderHash: hashIdentifier("sender", input.Sender.ID),
		InputText:  input.Content(),
		ReplyText:  replyText,
		Delivery:   delivery,
	}
	recorder.Record(event)
}

func hashIdentifier(kind, value string) string {
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(automaticReplyHashSalt + "\x00" + kind + "\x00" + value))
	return hex.EncodeToString(sum[:])
}
