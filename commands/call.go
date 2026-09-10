package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/sxyazi/bendan/platform"
)

func senderName(message *platform.Message) string {
	if message == nil || message.Sender.DisplayName == "" {
		return "有人"
	}
	return message.Sender.DisplayName
}

func targetOfInteraction(message *platform.Message) string {
	if message.ReplyTo == nil || message.ReplyTo.Sender.ID == message.Sender.ID || message.ReplyTo.Sender.ID == Bot.Identity().ID {
		return "自己"
	}
	return senderName(message.ReplyTo)
}

func Call(ctx context.Context, message *platform.Message) bool {
	if len(message.Text) < 2 || message.Text[0] != '/' || message.Text[1] == '/' {
		return false
	}

	params := strings.Fields(message.Text[1:])
	var text string
	switch len(params) {
	case 1:
		text = fmt.Sprintf("%s %s了 %s！", senderName(message), params[0], targetOfInteraction(message))
	case 2:
		text = fmt.Sprintf("%s %s %s %s！", senderName(message), params[0], targetOfInteraction(message), params[1])
	default:
		return false
	}
	sendText(ctx, message.Chat, text)
	return true
}
