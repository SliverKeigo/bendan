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
	if message.ReplyTo == nil || message.ReplyTo.Sender.ID == message.Sender.ID {
		return "自己"
	}
	return senderName(message.ReplyTo)
}

func mentionedTarget(message *platform.Message) string {
	if len(message.Mentions) != 1 {
		return ""
	}
	mention := message.Mentions[0]
	if mention.DisplayName != "" {
		return mention.DisplayName
	}
	if message.ReplyTo != nil && mention.ID == message.ReplyTo.Sender.ID {
		return senderName(message.ReplyTo)
	}
	if mention.ID == Bot.Identity().ID {
		return senderName(&platform.Message{Sender: Bot.Identity()})
	}
	return "@" + mention.ID
}

func actionTarget(message *platform.Message) string {
	if target := mentionedTarget(message); target != "" {
		return target
	}
	return targetOfInteraction(message)
}

func formatActionSentence(sender, display, target string) string {
	if strings.Contains(display, "{target}") {
		return fmt.Sprintf("%s %s！", sender, strings.ReplaceAll(display, "{target}", target))
	}
	return fmt.Sprintf("%s %s %s！", sender, display, target)
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
	if len(params) == 0 || (hasSlash && len(params) > 2) || (!hasSlash && len(params) != 2 && !(len(params) == 1 && (message.ReplyTo != nil || mentionedTarget(message) != ""))) {
		return false
	}
	if (hasSlash && !isSlashAction(params[0])) || (!hasSlash && !isAction(params[0])) {
		return false
	}

	var response string
	switch len(params) {
	case 1:
		response = formatActionSentence(senderName(message), actionDisplay(params[0]), actionTarget(message))
	case 2:
		if message.ReplyTo == nil || message.ReplyTo.Sender.ID == message.Sender.ID {
			response = formatActionSentence(senderName(message), actionDisplay(params[0]), params[1])
		} else {
			target := targetOfInteraction(message)
			display := actionDisplay(params[0])
			if strings.Contains(display, "{target}") {
				response = formatActionSentence(senderName(message), display, target+"的"+params[1])
			} else {
				response = fmt.Sprintf("%s %s %s的%s！", senderName(message), display, target, params[1])
			}
		}
	default:
		return false
	}
	sendText(ctx, message.Chat, response)
	return true
}
