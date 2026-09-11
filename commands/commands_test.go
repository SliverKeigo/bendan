package commands

import (
	"context"
	"strings"
	"sync"
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
		name     string
		text     string
		mentions []platform.User
		replyTo  *platform.Message
		want     string
	}{
		{name: "slash action only", text: "/摸", want: "Keigo 摸了 自己！"},
		{name: "unlisted Chinese action", text: "/看看", want: ""},
		{name: "bare action requires a result", text: "摸", want: ""},
		{name: "emoji action only", text: "/🤔", want: "Keigo 🤔 自己！"},
		{name: "completed action with result", text: "/喝了 自己", want: "Keigo 喝了 自己！"},
		{name: "slash action with result", text: "/摸 智智", want: "Keigo 摸了 智智！"},
		{name: "bare action with result", text: "摸 智智", want: "Keigo 摸了 智智！"},
		{name: "bare multi-character action with result", text: "抱抱 智智", want: "Keigo 抱了抱 智智！"},
		{name: "whitelisted English action", text: "rua 智智", want: "Keigo 揉了揉 智智！"},
		{name: "slash whitelisted English action", text: "/rua 智智", want: "Keigo 揉了揉 智智！"},
		{name: "English action ignores case", text: "Hug 智智", want: "Keigo 抱了 智智！"},
		{name: "English action without slash", text: "highfive 智智", want: "Keigo 击了掌 智智！"},
		{name: "unlisted English text", text: "recent 对吗？", want: ""},
		{
			name:     "action with bot mention target",
			text:     "摸",
			mentions: []platform.User{{ID: "99", DisplayName: "Bendan"}},
			want:     "Keigo 摸了 Bendan！",
		},
		{
			name:    "action with result replying to bot targets bot",
			text:    "摸 头",
			replyTo: &platform.Message{Sender: platform.User{ID: "99", DisplayName: "Bendan"}},
			want:    "Keigo 摸了 Bendan的头！",
		},
		{
			name:    "action with result replying to another user",
			text:    "/摸 头",
			replyTo: &platform.Message{Sender: platform.User{ID: "2", DisplayName: "智智"}},
			want:    "Keigo 摸了 智智的头！",
		},
		{
			name:    "repeated action with result replying to another user",
			text:    "摸摸 头",
			replyTo: &platform.Message{Sender: platform.User{ID: "2", DisplayName: "智智"}},
			want:    "Keigo 摸了摸 智智的头！",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := &recordingBot{identity: platform.User{ID: "99", DisplayName: "Bendan"}}
			withTestBot(t, bot)

			Handle(context.Background(), &platform.Message{
				Chat:     platform.Chat{ID: "123", Kind: "group"},
				Sender:   platform.User{ID: "1", DisplayName: "Keigo"},
				Text:     tt.text,
				Mentions: tt.mentions,
				ReplyTo:  tt.replyTo,
			})

			if tt.want == "" {
				if len(bot.sent) != 0 {
					t.Fatalf("sent = %#v, want no response", bot.sent)
				}
				return
			}
			if len(bot.sent) != 1 || bot.sent[0] != tt.want {
				t.Fatalf("sent = %#v, want %q", bot.sent, tt.want)
			}
		})
	}
}

func TestHandlePreservesAutomaticRepliesForUnlistedChineseActions(t *testing.T) {
	bot := &recordingBot{identity: platform.User{ID: "99", DisplayName: "Bendan"}}
	withTestBot(t, bot)
	resetMessageGuards(t)

	Handle(context.Background(), &platform.Message{
		ID:     "1",
		Chat:   platform.Chat{ID: "123", Kind: "group"},
		Sender: platform.User{ID: "1", DisplayName: "Keigo"},
		Text:   "看看 这个",
	})

	if len(bot.sent) != 1 || !strings.Contains(bot.sent[0], "看") {
		t.Fatalf("sent = %#v, want a preserved look automatic reply", bot.sent)
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

func TestHandleRateLimitsRepeatedAutomaticResponsesFromOneUser(t *testing.T) {
	bot := &recordingBot{identity: platform.User{ID: "99", DisplayName: "Bendan"}}
	withTestBot(t, bot)
	resetMessageGuards(t)

	for _, id := range []string{"1", "2"} {
		Handle(context.Background(), &platform.Message{
			ID:     id,
			Chat:   platform.Chat{ID: "123", Kind: "group"},
			Sender: platform.User{ID: "1", DisplayName: "Keigo"},
			Text:   "？",
		})
	}

	if len(bot.replied) != 1 {
		t.Fatalf("replied %d messages, want one user-rate-limited response: %#v", len(bot.replied), bot.replied)
	}
}

func TestHandleDoesNotRateLimitExplicitCommands(t *testing.T) {
	bot := &recordingBot{identity: platform.User{ID: "99", DisplayName: "Bendan"}}
	withTestBot(t, bot)
	resetMessageGuards(t)

	for _, id := range []string{"1", "2"} {
		Handle(context.Background(), &platform.Message{
			ID:     id,
			Chat:   platform.Chat{ID: "123", Kind: "group"},
			Sender: platform.User{ID: "1", DisplayName: "Keigo"},
			Text:   "/me 喝茶",
		})
	}

	if len(bot.sent) != 2 {
		t.Fatalf("sent %d messages, want two explicit command responses: %#v", len(bot.sent), bot.sent)
	}
}

func TestHandleDedupesMessageIDs(t *testing.T) {
	bot := &recordingBot{identity: platform.User{ID: "99", DisplayName: "Bendan"}}
	withTestBot(t, bot)
	resetMessageGuards(t)

	message := &platform.Message{
		ID:     "1",
		Chat:   platform.Chat{ID: "123", Kind: "group"},
		Sender: platform.User{ID: "1", DisplayName: "Keigo"},
		Text:   "/me 喝茶",
	}
	Handle(context.Background(), message)
	Handle(context.Background(), message)

	if len(bot.sent) != 1 {
		t.Fatalf("sent %d messages, want one deduplicated response: %#v", len(bot.sent), bot.sent)
	}
}

func TestHandleDedupesConcurrentMessageIDs(t *testing.T) {
	bot := &recordingBot{identity: platform.User{ID: "99", DisplayName: "Bendan"}}
	withTestBot(t, bot)
	resetMessageGuards(t)

	const workers = 32
	var group sync.WaitGroup
	group.Add(workers)
	for range workers {
		go func() {
			defer group.Done()
			Handle(context.Background(), &platform.Message{
				ID:     "1",
				Chat:   platform.Chat{ID: "123", Kind: "group"},
				Sender: platform.User{ID: "1", DisplayName: "Keigo"},
				Text:   "/me 喝茶",
			})
		}()
	}
	group.Wait()

	if len(bot.sent) != 1 {
		t.Fatalf("sent %d messages, want one concurrent deduplicated response: %#v", len(bot.sent), bot.sent)
	}
}

func TestHandleRateLimitsAutomaticReplyToOriginalSender(t *testing.T) {
	bot := &recordingBot{identity: platform.User{ID: "99", DisplayName: "Bendan"}}
	withTestBot(t, bot)
	resetMessageGuards(t)

	for _, id := range []string{"1", "2"} {
		Handle(context.Background(), &platform.Message{
			ID:      id,
			Chat:    platform.Chat{ID: "123", Kind: "group"},
			Sender:  platform.User{ID: "1", DisplayName: "Keigo"},
			Text:    "看看",
			ReplyTo: &platform.Message{Sender: platform.User{ID: "2", DisplayName: "智智"}},
		})
	}

	if len(bot.replied) != 1 {
		t.Fatalf("replied %d messages, want one response limited by the triggering sender: %#v", len(bot.replied), bot.replied)
	}
}
