package commands

import (
	"context"
	"fmt"

	"github.com/sxyazi/bendan/platform"
)

func Whoami(ctx context.Context, message *platform.Message) bool {
	if message.Text != "//whoami" {
		return false
	}
	replyText(ctx, message, fmt.Sprintf("QQ：%s\n会话：%s", message.Sender.ID, message.Chat.ID))
	return true
}
