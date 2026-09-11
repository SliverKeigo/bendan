package commands

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/sxyazi/bendan/platform"
)

var Bot platform.Bot

const automaticReplyCooldown = 5 * time.Second
const messageDeduplicationWindow = 10 * time.Minute

type automaticReplyContextKey struct{}
type automaticReplyMessageContextKey struct{}

var messageGuards = struct {
	sync.Mutex
	automaticReplies map[string]time.Time
	messages         map[string]time.Time
}{
	automaticReplies: make(map[string]time.Time),
	messages:         make(map[string]time.Time),
}

var directHandlers = []func(context.Context, *platform.Message) bool{
	Hush,
	Whoami,
	Eval,
	Me,
	Dontworry,
	Call,
}

var automaticHandlers = []func(context.Context, *platform.Message) bool{
	Mark,
	YesRight,
	YesIs,
	YesCan,
	YesLook,
}

// Handle dispatches a platform-neutral message to the first matching command.
func Handle(ctx context.Context, message *platform.Message) {
	if message == nil || message.IsBot || !acceptMessage(message) {
		return
	}
	for _, handler := range directHandlers {
		if handler(ctx, message) {
			return
		}
	}

	ctx = context.WithValue(ctx, automaticReplyContextKey{}, true)
	ctx = context.WithValue(ctx, automaticReplyMessageContextKey{}, message)
	for _, handler := range automaticHandlers {
		if handler(ctx, message) {
			return
		}
	}
}

func acceptMessage(message *platform.Message) bool {
	if message.ID == "" {
		return true
	}

	now := time.Now()
	key := message.Chat.ID + ":" + message.ID
	messageGuards.Lock()
	defer messageGuards.Unlock()
	for key, seenAt := range messageGuards.messages {
		if now.Sub(seenAt) >= messageDeduplicationWindow {
			delete(messageGuards.messages, key)
		}
	}
	if _, seen := messageGuards.messages[key]; seen {
		return false
	}
	messageGuards.messages[key] = now
	return true
}

func canSend(ctx context.Context, chat platform.Chat, sender platform.User) bool {
	if ctx.Value(automaticReplyContextKey{}) == nil {
		return true
	}

	now := time.Now()
	key := chat.ID + ":" + sender.ID
	messageGuards.Lock()
	defer messageGuards.Unlock()
	lastReply, replied := messageGuards.automaticReplies[key]
	if replied && now.Sub(lastReply) < automaticReplyCooldown {
		return false
	}
	messageGuards.automaticReplies[key] = now
	return true
}

func resetMessageGuards(t interface{ Cleanup(func()) }) {
	messageGuards.Lock()
	previousAutomaticReplies := messageGuards.automaticReplies
	previousMessages := messageGuards.messages
	messageGuards.automaticReplies = make(map[string]time.Time)
	messageGuards.messages = make(map[string]time.Time)
	messageGuards.Unlock()
	t.Cleanup(func() {
		messageGuards.Lock()
		messageGuards.automaticReplies = previousAutomaticReplies
		messageGuards.messages = previousMessages
		messageGuards.Unlock()
	})
}

func sendText(ctx context.Context, chat platform.Chat, text string) *platform.Message {
	if message, automatic := ctx.Value(automaticReplyMessageContextKey{}).(*platform.Message); automatic && !canSend(ctx, chat, message.Sender) {
		return nil
	}
	message, err := Bot.SendText(ctx, chat, text)
	if err != nil {
		log.Printf("send text: %v", err)
		return nil
	}
	return message
}

func replyText(ctx context.Context, message *platform.Message, text string) *platform.Message {
	chat := message.Chat
	sender := message.Sender
	if automaticMessage, automatic := ctx.Value(automaticReplyMessageContextKey{}).(*platform.Message); automatic {
		chat = automaticMessage.Chat
		sender = automaticMessage.Sender
	}
	if !canSend(ctx, chat, sender) {
		return nil
	}
	sent, err := Bot.ReplyText(ctx, message, text)
	if err != nil {
		log.Printf("reply text: %v", err)
		return nil
	}
	return sent
}

func deleteMessage(ctx context.Context, message *platform.Message) bool {
	if err := Bot.DeleteMessage(ctx, message); err != nil {
		log.Printf("delete message: %v", err)
		return false
	}
	return true
}
