package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/sxyazi/bendan/platform"
)

func Me(ctx context.Context, message *platform.Message) bool {
	if !strings.HasPrefix(message.Text, "/me") {
		return false
	}

	text := strings.TrimSpace(message.Text[3:])
	if text == "" {
		return false
	}
	sendText(ctx, message.Chat, fmt.Sprintf("%s %s！", senderName(message), text))
	return true
}
