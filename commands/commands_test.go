package commands

import (
	"context"
	"strings"
	"testing"

	"github.com/sxyazi/bendan/platform"
)

func TestHandleMeUsesPlatformBot(t *testing.T) {
	bot := &recordingBot{identity: platform.User{ID: "99", DisplayName: "Bendan"}}
	withTestBot(t, bot)

	Handle(context.Background(), &platform.Message{
		Chat:   platform.Chat{ID: "123", Kind: "group"},
		Sender: platform.User{ID: "1", DisplayName: "Keigo"},
		Text:   "/me 喝茶",
	})

	if len(bot.sent) != 1 || bot.sent[0] != "Keigo 喝茶！" {
		t.Fatalf("sent = %#v, want one /me reply", bot.sent)
	}
}

func TestHandleWhoamiUsesQQIdentifiers(t *testing.T) {
	bot := &recordingBot{identity: platform.User{ID: "99"}}
	withTestBot(t, bot)

	Handle(context.Background(), &platform.Message{
		ID:     "77",
		Chat:   platform.Chat{ID: "123456", Kind: "group"},
		Sender: platform.User{ID: "10001", DisplayName: "Keigo"},
		Text:   "//whoami",
	})

	if len(bot.replied) != 1 || !strings.Contains(bot.replied[0], "QQ：10001") || !strings.Contains(bot.replied[0], "会话：123456") {
		t.Fatalf("replied = %#v, want QQ identifiers", bot.replied)
	}
}
