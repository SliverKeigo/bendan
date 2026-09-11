package commands

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sxyazi/bendan/platform"
)

func TestActionsCommandEditsRuntimeLexicon(t *testing.T) {
	path := t.TempDir() + "/actions.json"
	if err := LoadActionLexicon(path); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := LoadActionLexicon("does-not-exist.json"); err != nil {
			t.Fatal(err)
		}
	})

	if err := updateActionLexicon("latin", "wave", "挥了挥"); err != nil {
		t.Fatal(err)
	}
	if !isAction("WAVE") || actionDisplay("wave") != "挥了挥" {
		t.Fatalf("saved lexicon did not add wave action")
	}
	if err := removeActionLexiconEntry("latin", "wave"); err != nil {
		t.Fatal(err)
	}
	if isAction("wave") {
		t.Fatal("saved lexicon did not remove wave action")
	}
}

func TestActionsCommandRequiresAdministrator(t *testing.T) {
	bot := &recordingBot{identity: platform.User{ID: "99", DisplayName: "Bendan"}}
	withTestBot(t, bot)

	for _, test := range []struct {
		name   string
		sender string
		text   string
		want   string
	}{
		{name: "non administrator is ignored", sender: "2", text: "//actions", want: ""},
		{name: "administrator sees status", sender: administratorQQ, text: "//actions", want: "动作词表"},
		{name: "administrator lists actions", sender: administratorQQ, text: "//actions list", want: "中文："},
	} {
		t.Run(test.name, func(t *testing.T) {
			bot.mu.Lock()
			bot.replied = nil
			bot.mu.Unlock()
			Handle(context.Background(), &platform.Message{
				Chat:   platform.Chat{ID: "123", Kind: "group"},
				Sender: platform.User{ID: test.sender, DisplayName: "Keigo"},
				Text:   test.text,
			})
			bot.mu.Lock()
			defer bot.mu.Unlock()
			if test.want == "" {
				if len(bot.replied) != 0 {
					t.Fatalf("replied = %#v, want none", bot.replied)
				}
				return
			}
			if len(bot.replied) != 1 || !strings.Contains(bot.replied[0], test.want) {
				t.Fatalf("replied = %#v, want text containing %q", bot.replied, test.want)
			}
		})
	}
}

func TestEvalRequiresAdministrator(t *testing.T) {
	bot := &recordingBot{identity: platform.User{ID: "99", DisplayName: "Bendan"}}
	withTestBot(t, bot)

	handled := Eval(context.Background(), &platform.Message{
		Chat:   platform.Chat{ID: "123", Kind: "group"},
		Sender: platform.User{ID: "2", DisplayName: "Other"},
		Text:   "//go fmt.Println(1)",
	})
	if !handled || len(bot.replied) != 0 {
		t.Fatalf("handled=%t replied=%#v, want silently handled non-admin eval", handled, bot.replied)
	}
}

func TestWatchActionLexiconReloadsModifiedFile(t *testing.T) {
	path := t.TempDir() + "/actions.json"
	writeLexicon := func(contents string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	writeLexicon(`{"zh":{"挥":"挥了挥"}}`)
	if err := LoadActionLexicon(path); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		WatchActionLexicon(ctx, 5*time.Millisecond)
		close(done)
	}()
	defer func() {
		cancel()
		<-done
		if err := LoadActionLexicon("does-not-exist.json"); err != nil {
			t.Fatal(err)
		}
	}()

	time.Sleep(20 * time.Millisecond)
	writeLexicon(`{"zh":{"摇":"摇了摇"}}`)
	deadline := time.Now().Add(time.Second)
	for !isAction("摇") && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if !isAction("摇") || isAction("挥") {
		t.Fatalf("hot reload did not replace actions: shake=%t wave=%t", isAction("摇"), isAction("挥"))
	}
}

func TestWatchActionLexiconKeepsPreviousLexiconAfterInvalidUpdate(t *testing.T) {
	path := t.TempDir() + "/actions.json"
	if err := os.WriteFile(path, []byte(`{"zh":{"挥":"挥了挥"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := LoadActionLexicon(path); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		WatchActionLexicon(ctx, 5*time.Millisecond)
		close(done)
	}()
	defer func() {
		cancel()
		<-done
		if err := LoadActionLexicon("does-not-exist.json"); err != nil {
			t.Fatal(err)
		}
	}()

	time.Sleep(20 * time.Millisecond)
	if err := os.WriteFile(path, []byte(`{"zh":`), 0o600); err != nil {
		t.Fatal(err)
	}
	time.Sleep(30 * time.Millisecond)
	if !isAction("挥") {
		t.Fatal("invalid hot reload discarded the previous action lexicon")
	}
}

func TestLoadActionLexiconUsesRuntimeFile(t *testing.T) {
	path := t.TempDir() + "/actions.json"
	if err := os.WriteFile(path, []byte(`{
		"zh": {"挥": "挥了挥"},
		"latin": {"wave": "挥了挥"}
	}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := LoadActionLexicon(path); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := LoadActionLexicon("does-not-exist.json"); err != nil {
			t.Fatal(err)
		}
	})

	if !isAction("挥") || !isAction("WAVE") {
		t.Fatal("runtime action lexicon did not load configured actions")
	}
	if actionDisplay("wave") != "挥了挥" {
		t.Fatalf("actionDisplay(wave) = %q, want %q", actionDisplay("wave"), "挥了挥")
	}
	if isAction("摸") {
		t.Fatal("runtime action lexicon retained an action outside the configured file")
	}
}

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
			name:    "action replying to another user without result",
			text:    "摸",
			replyTo: &platform.Message{Sender: platform.User{ID: "2", DisplayName: "智智"}},
			want:    "Keigo 摸了 智智！",
		},
		{
			name:     "action with mention uses reply name when mention omits it",
			text:     "摸",
			mentions: []platform.User{{ID: "2"}},
			replyTo:  &platform.Message{Sender: platform.User{ID: "2", DisplayName: "智智"}},
			want:     "Keigo 摸了 智智！",
		},
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
