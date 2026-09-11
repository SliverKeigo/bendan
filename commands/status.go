package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sxyazi/bendan/platform"
)

func Status(ctx context.Context, message *platform.Message) bool {
	if message.Text != "//status" {
		return false
	}
	if !isAdministrator(message) {
		return true
	}

	connection := "未知"
	if provider, ok := Bot.(platform.RuntimeStatusProvider); ok {
		if provider.RuntimeStatus().Connected {
			connection = "已连接"
		} else {
			connection = "未连接"
		}
	}
	_, _, zh, latin := actionLexiconStatus()
	_, hushed := hushUntil(message.Chat.ID)
	replyText(ctx, message, fmt.Sprintf("Bendan 状态\nOneBot：%s\n运行时间：%s\n动作词表：中文 %d，英文 %d\n当前会话：%s", connection, formatDuration(time.Since(startedAt)), zh, latin, hushStatus(hushed)))
	return true
}

func HushStatus(ctx context.Context, message *platform.Message) bool {
	if message.Text != "//hush status" {
		return false
	}
	if !isAdministrator(message) {
		return true
	}

	until, hushed := hushUntil(message.Chat.ID)
	if !hushed {
		replyText(ctx, message, "当前会话未静默")
		return true
	}
	replyText(ctx, message, fmt.Sprintf("当前会话静默中，剩余 %s（至 %s）", formatDuration(time.Until(until)), until.Format("15:04:05")))
	return true
}

func hushStatus(hushed bool) string {
	if hushed {
		return "静默中"
	}
	return "正常"
}

func formatDuration(duration time.Duration) string {
	if duration < 0 {
		duration = 0
	}
	duration = duration.Round(time.Second)
	hours := int(duration / time.Hour)
	minutes := int(duration % time.Hour / time.Minute)
	seconds := int(duration % time.Minute / time.Second)
	parts := make([]string, 0, 3)
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d小时", hours))
	}
	if minutes > 0 || hours > 0 {
		parts = append(parts, fmt.Sprintf("%d分", minutes))
	}
	parts = append(parts, fmt.Sprintf("%d秒", seconds))
	return strings.Join(parts, "")
}
