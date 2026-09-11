package commands

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/sxyazi/bendan/platform"
)

const hushDuration = 30 * time.Minute

var hushDir = filepath.Join(os.TempDir(), "bendan", "hush")
var reHush = regexp.MustCompile("别说话|不要说话|别讲话|不要讲话|闭嘴|住嘴|安静|别吵|消停")
var reUnHush = regexp.MustCompile("说话")

func init() {
	if err := os.MkdirAll(hushDir, 0755); err != nil {
		log.Printf("create hush directory: %v", err)
	}
}

func hushUntil(chatID string) (time.Time, bool) {
	info, err := os.Lstat(filepath.Join(hushDir, chatID))
	if err != nil {
		return time.Time{}, false
	}
	until := info.ModTime().Add(hushDuration)
	return until, time.Now().Before(until)
}

func targetsBot(message *platform.Message) bool {
	if message == nil {
		return false
	}
	botID := Bot.Identity().ID
	if message.ReplyTo != nil && message.ReplyTo.Sender.ID == botID {
		return true
	}
	for _, mention := range message.Mentions {
		if mention.ID == botID {
			return true
		}
	}
	return false
}

func Hush(ctx context.Context, message *platform.Message) bool {
	path := filepath.Join(hushDir, message.Chat.ID)
	targetsBot := targetsBot(message)

	if targetsBot && reHush.MatchString(message.Text) {
		if err := os.WriteFile(path, nil, 0644); err == nil {
			replyText(ctx, message, "😭")
		} else {
			log.Printf("hush: %v", err)
			replyText(ctx, message, "不要！")
		}
		return true
	}

	if targetsBot && reUnHush.MatchString(message.Text) {
		if err := os.Remove(path); err == nil || errors.Is(err, os.ErrNotExist) {
			replyText(ctx, message, "好耶！")
		} else {
			log.Printf("unhush: %v", err)
			replyText(ctx, message, "不要！")
		}
		return true
	}

	_, hushed := hushUntil(message.Chat.ID)
	return hushed
}
