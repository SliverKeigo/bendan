package commands

import (
	"context"
	"log"

	"github.com/sxyazi/bendan/platform"
)

var Bot platform.Bot

var viaMessage = []func(context.Context, *platform.Message) bool{
	Hush,
	Whoami,
	Eval,
	Me,
	Dontworry,
	Call,
	Purify,
	Mark,
	YesRight,
	YesIs,
	YesCan,
	YesLook,
}

// Handle dispatches a platform-neutral message to the first matching command.
func Handle(ctx context.Context, message *platform.Message) {
	if message == nil || message.IsBot {
		return
	}
	for _, handler := range viaMessage {
		if handler(ctx, message) {
			return
		}
	}
}

func sendText(ctx context.Context, chat platform.Chat, text string) *platform.Message {
	message, err := Bot.SendText(ctx, chat, text)
	if err != nil {
		log.Printf("send text: %v", err)
		return nil
	}
	return message
}

func replyText(ctx context.Context, message *platform.Message, text string) *platform.Message {
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
