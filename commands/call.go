package commands

import (
	"context"
	"fmt"
	"strings"
	"unicode"

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

func actionWithParticle(action string) string {
	if strings.HasSuffix(action, "了") || !unicode.Is(unicode.Han, []rune(action)[0]) {
		return action
	}
	return action + "了"
}

func Call(ctx context.Context, message *platform.Message) bool {
	text := message.Text
	hasSlash := strings.HasPrefix(text, "/")
	if hasSlash {
		if len(text) < 2 || strings.HasPrefix(text, "//") {
			return false
		}
		text = text[1:]
	}

	params := strings.Fields(text)
	if len(params) == 0 || (hasSlash && len(params) > 2) || (!hasSlash && len(params) != 2) {
		return false
	}
	if hasSlash && !unicode.Is(unicode.Han, []rune(params[0])[0]) && len(params) != 1 {
		return false
	}

	var response string
	switch len(params) {
	case 1:
		response = fmt.Sprintf("%s %s %s！", senderName(message), actionWithParticle(params[0]), targetOfInteraction(message))
	case 2:
		if message.ReplyTo == nil || message.ReplyTo.Sender.ID == message.Sender.ID || message.ReplyTo.Sender.ID == Bot.Identity().ID {
			response = fmt.Sprintf("%s %s %s！", senderName(message), actionWithParticle(params[0]), params[1])
		} else {
			response = fmt.Sprintf("%s %s %s %s！", senderName(message), params[0], targetOfInteraction(message), params[1])
		}
	default:
		return false
	}
	sendText(ctx, message.Chat, response)
	return true
}
