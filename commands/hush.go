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
var reHush = regexp.MustCompile("别说话|闭嘴|安静")
var reUnHush = regexp.MustCompile("说话")

func init() {
	if err := os.MkdirAll(hushDir, 0755); err != nil {
		log.Printf("create hush directory: %v", err)
	}
}

func Hush(ctx context.Context, message *platform.Message) bool {
	path := filepath.Join(hushDir, message.Chat.ID)
	repliesToBot := message.ReplyTo != nil && message.ReplyTo.Sender.ID == Bot.Identity().ID

	if repliesToBot && reHush.MatchString(message.Text) {
		if err := os.WriteFile(path, nil, 0644); err == nil {
			replyText(ctx, message, "😭")
		} else {
			log.Printf("hush: %v", err)
			replyText(ctx, message, "不要！")
		}
		return true
	}

	if repliesToBot && reUnHush.MatchString(message.Text) {
		if err := os.Remove(path); err == nil || errors.Is(err, os.ErrNotExist) {
			replyText(ctx, message, "好耶！")
		} else {
			log.Printf("unhush: %v", err)
			replyText(ctx, message, "不要！")
		}
		return true
	}

	info, err := os.Lstat(path)
	return err == nil && time.Now().Before(info.ModTime().Add(hushDuration))
}
