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

func TestHandleCallFormatsActionsWithoutDuplicatingTarget(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		replyTo *platform.Message
		want    string
	}{
		{name: "action only", text: "/摸", want: "Keigo 摸了 自己！"},
		{name: "completed action with result", text: "/喝了 自己", want: "Keigo 喝了 自己！"},
		{name: "action with result", text: "/摸 智智", want: "Keigo 摸了 智智！"},
		{
			name:    "action with result replying to another user",
			text:    "/摸 头",
			replyTo: &platform.Message{Sender: platform.User{ID: "2", DisplayName: "智智"}},
			want:    "Keigo 摸 智智 头！",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := &recordingBot{identity: platform.User{ID: "99", DisplayName: "Bendan"}}
			withTestBot(t, bot)

			Handle(context.Background(), &platform.Message{
				Chat:    platform.Chat{ID: "123", Kind: "group"},
				Sender:  platform.User{ID: "1", DisplayName: "Keigo"},
				Text:    tt.text,
				ReplyTo: tt.replyTo,
			})

			if len(bot.sent) != 1 || bot.sent[0] != tt.want {
				t.Fatalf("sent = %#v, want %q", bot.sent, tt.want)
			}
		})
	}
}

func TestCallDoesNotTreatLatinSlashCommandsAsActions(t *testing.T) {
	bot := &recordingBot{identity: platform.User{ID: "99", DisplayName: "Bendan"}}
	withTestBot(t, bot)

	handled := Call(context.Background(), &platform.Message{
		Chat:   platform.Chat{ID: "123", Kind: "group"},
		Sender: platform.User{ID: "1", DisplayName: "Keigo"},
		Text:   "/recent 对吗？",
	})

	if handled || len(bot.sent) != 0 || len(bot.replied) != 0 {
		t.Fatalf("Call handled a non-action slash command: handled=%t sent=%#v replied=%#v", handled, bot.sent, bot.replied)
	}
}
