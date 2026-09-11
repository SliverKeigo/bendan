package commands

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/sxyazi/bendan/platform"
)

var Bot platform.Bot
var startedAt = time.Now()

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

type namedHandler struct {
	name   string
	handle func(context.Context, *platform.Message) bool
}

var directHandlers = []namedHandler{
	{name: "hush_status", handle: HushStatus},
	{name: "hush", handle: Hush},
	{name: "whoami", handle: Whoami},
	{name: "status", handle: Status},
	{name: "actions", handle: Actions},
	{name: "eval", handle: Eval},
	{name: "me", handle: Me},
	{name: "dontworry", handle: Dontworry},
	{name: "call", handle: Call},
}

var automaticHandlers = []namedHandler{
	{name: "mark", handle: Mark},
	{name: "yes_choice", handle: YesChoice},
	{name: "yes_right", handle: YesRight},
	{name: "yes_is", handle: YesIs},
	{name: "yes_can", handle: YesCan},
	{name: "yes_look", handle: YesLook},
}

// Handle dispatches a platform-neutral message to the first matching command.
func Handle(ctx context.Context, message *platform.Message) {
	if message == nil || message.IsBot || !acceptMessage(message) {
		return
	}
	for _, handler := range directHandlers {
		if handler.handle(ctx, message) {
			log.Printf("command handled sender=%q chat=%q handler=%s", message.Sender.ID, message.Chat.ID, handler.name)
			return
		}
	}

	ctx = context.WithValue(ctx, automaticReplyContextKey{}, true)
	ctx = context.WithValue(ctx, automaticReplyMessageContextKey{}, message)
	for _, handler := range automaticHandlers {
		if handler.handle(ctx, message) {
			log.Printf("automatic reply handled sender=%q chat=%q handler=%s", message.Sender.ID, message.Chat.ID, handler.name)
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
