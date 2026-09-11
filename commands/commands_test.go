package commands

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/sxyazi/bendan/platform"
)

const testAdministratorQQ = "1226355793"

func withTestAdministrator(t *testing.T) {
	t.Setenv("ADMINISTRATOR_QQ", testAdministratorQQ)
}

func TestMarkReplyIsValidUTF8(t *testing.T) {
	for _, input := range []string{"？", "¿", "‽", "？？？", "?？¿‽"} {
		t.Run(input, func(t *testing.T) {
			bot := &recordingBot{identity: platform.User{ID: "99", DisplayName: "Bendan"}}
			withTestBot(t, bot)

			message := &platform.Message{
				Chat:   platform.Chat{ID: "123", Kind: "group"},
				Sender: platform.User{ID: "1", DisplayName: "Keigo"},
				Text:   input,
			}
			for i := 0; i < 100; i++ {
				bot.mu.Lock()
				bot.replied = nil
				bot.mu.Unlock()

				if !Mark(context.Background(), message) {
					t.Fatalf("Mark returned false for %q", input)
				}

				bot.mu.Lock()
				if len(bot.replied) != 1 {
					bot.mu.Unlock()
					t.Fatalf("replied = %#v, want exactly one reply", bot.replied)
				}
				reply := bot.replied[0]
				bot.mu.Unlock()
				if !utf8.ValidString(reply) {
					t.Fatalf("reply %q contains invalid UTF-8 bytes: % x", reply, []byte(reply))
				}
			}
		})
	}
}

func TestStatusCommandsRequireAdministrator(t *testing.T) {
	withTestAdministrator(t)
	bot := &recordingBot{identity: platform.User{ID: "99", DisplayName: "Bendan"}}
	withTestBot(t, bot)

	for _, test := range []struct {
		name   string
		sender string
		text   string
		want   string
	}{
		{name: "non administrator status is ignored", sender: "2", text: "//status", want: ""},
		{name: "administrator sees status", sender: testAdministratorQQ, text: "//status", want: "Bendan 状态"},
		{name: "administrator sees unhushed status", sender: testAdministratorQQ, text: "//hush status", want: "当前会话未静默"},
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

func TestHushStatusReportsRemainingTime(t *testing.T) {
	withTestAdministrator(t)
	bot := &recordingBot{identity: platform.User{ID: "99", DisplayName: "Bendan"}}
	withTestBot(t, bot)

	previousDir := hushDir
	hushDir = t.TempDir()
	t.Cleanup(func() { hushDir = previousDir })
	if err := os.WriteFile(hushDir+"/123", nil, 0o600); err != nil {
		t.Fatal(err)
	}

	Handle(context.Background(), &platform.Message{
		Chat:   platform.Chat{ID: "123", Kind: "group"},
		Sender: platform.User{ID: testAdministratorQQ, DisplayName: "Keigo"},
		Text:   "//hush status",
	})
	if len(bot.replied) != 1 || !strings.Contains(bot.replied[0], "当前会话静默中") {
		t.Fatalf("replied = %#v, want active hush status", bot.replied)
	}
}

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
	withTestAdministrator(t)
	bot := &recordingBot{identity: platform.User{ID: "99", DisplayName: "Bendan"}}
	withTestBot(t, bot)

	for _, test := range []struct {
		name   string
		sender string
		text   string
		want   string
	}{
		{name: "non administrator is ignored", sender: "2", text: "//actions", want: ""},
		{name: "administrator sees status", sender: testAdministratorQQ, text: "//actions", want: "动作词表"},
		{name: "administrator lists actions", sender: testAdministratorQQ, text: "//actions list", want: "中文："},
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
	withTestAdministrator(t)
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

func TestDefaultActionLexiconMappings(t *testing.T) {
	if err := LoadActionLexicon("does-not-exist.json"); err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"喝":         "喝了{target}",
		"摸":         "摸了摸{target}",
		"摸摸":        "摸了摸{target}",
		"抱":         "抱了抱{target}",
		"抱抱":        "抱了抱{target}",
		"拍":         "拍了拍{target}",
		"拍拍":        "拍了拍{target}",
		"拍肩":        "拍了拍{target}的肩",
		"戳":         "戳了戳{target}",
		"戳戳":        "戳了戳{target}",
		"亲":         "亲了亲{target}",
		"亲亲":        "亲了亲{target}",
		"揉":         "揉了揉{target}",
		"揉揉":        "揉了揉{target}",
		"揉头":        "揉了揉{target}的头",
		"捏":         "捏了捏{target}",
		"捏捏":        "捏了捏{target}",
		"蹭":         "蹭了蹭{target}",
		"蹭蹭":        "蹭了蹭{target}",
		"贴":         "和{target}贴了贴",
		"贴贴":        "和{target}贴了贴",
		"啵":         "亲了{target}一口",
		"啵啵":        "亲了{target}两口",
		"抓":         "抓住了{target}",
		"挠":         "挠了挠{target}",
		"挠挠":        "挠了挠{target}",
		"挠痒":        "给{target}挠了挠痒",
		"打":         "打了{target}",
		"踢":         "踢了{target}",
		"咬":         "咬了{target}",
		"咬咬":        "咬了咬{target}",
		"舔":         "舔了{target}",
		"舔舔":        "舔了舔{target}",
		"夸":         "夸了{target}",
		"夸夸":        "夸了夸{target}",
		"挥":         "向{target}挥了挥手",
		"拉":         "拉了拉{target}",
		"推":         "推了推{target}",
		"递":         "递给了{target}",
		"喂":         "喂了{target}",
		"塞":         "塞给了{target}",
		"比心":        "给{target}比了个心",
		"击掌":        "和{target}击了个掌",
		"握手":        "和{target}握了握手",
		"举":         "把{target}举了举",
		"举高高":       "把{target}举高高了",
		"rua":       "揉了揉{target}",
		"hug":       "抱了抱{target}",
		"cuddle":    "抱了抱{target}",
		"hold":      "抱住了{target}",
		"kiss":      "亲了亲{target}",
		"pat":       "拍了拍{target}",
		"pet":       "摸了摸{target}",
		"headpat":   "摸了摸{target}的头",
		"poke":      "戳了戳{target}",
		"boop":      "轻轻碰了碰{target}",
		"bonk":      "敲了敲{target}",
		"slap":      "打了{target}",
		"bite":      "咬了{target}",
		"lick":      "舔了{target}",
		"squeeze":   "捏了捏{target}",
		"tickle":    "给{target}挠了挠痒",
		"feed":      "喂了{target}",
		"wave":      "向{target}挥了挥手",
		"fistbump":  "和{target}碰了碰拳",
		"highfive":  "和{target}击了个掌",
		"handshake": "和{target}握了握手",
	}
	for action, display := range want {
		if !isAction(action) {
			t.Errorf("isAction(%q) = false", action)
		}
		if got := actionDisplay(action); got != display {
			t.Errorf("actionDisplay(%q) = %q, want %q", action, got, display)
		}
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
		{name: "slash action only", text: "/摸", want: "Keigo 摸了摸自己！"},
		{name: "unlisted Chinese action", text: "/看看", want: ""},
		{name: "bare action requires a result", text: "摸", want: ""},
		{name: "emoji action only", text: "/🤔", want: "Keigo 🤔 自己！"},
		{name: "completed action with result", text: "/喝了 自己", want: "Keigo 喝了 自己！"},
		{name: "slash action with result", text: "/摸 智智", want: "Keigo 摸了摸智智！"},
		{name: "bare action with result", text: "摸 智智", want: "Keigo 摸了摸智智！"},
		{name: "bare multi-character action with result", text: "抱抱 智智", want: "Keigo 抱了抱智智！"},
		{name: "new Chinese action", text: "握手 智智", want: "Keigo 和智智握了握手！"},
		{name: "directed Chinese action", text: "比心 智智", want: "Keigo 给智智比了个心！"},
		{name: "reciprocal Chinese action", text: "击掌 智智", want: "Keigo 和智智击了个掌！"},
		{name: "whitelisted English action", text: "rua 智智", want: "Keigo 揉了揉智智！"},
		{name: "slash whitelisted English action", text: "/rua 智智", want: "Keigo 揉了揉智智！"},
		{name: "English action ignores case", text: "Hug 智智", want: "Keigo 抱了抱智智！"},
		{name: "new English alias", text: "headpat 智智", want: "Keigo 摸了摸智智的头！"},
		{name: "directed English alias", text: "feed 智智", want: "Keigo 喂了智智！"},
		{name: "English action without slash", text: "highfive 智智", want: "Keigo 和智智击了个掌！"},
		{name: "reciprocal English alias", text: "handshake 智智", want: "Keigo 和智智握了握手！"},
		{name: "unlisted English text", text: "recent 对吗？", want: ""},
		{
			name:    "action replying to another user without result",
			text:    "摸",
			replyTo: &platform.Message{Sender: platform.User{ID: "2", DisplayName: "智智"}},
			want:    "Keigo 摸了摸智智！",
		},
		{
			name:     "action with mention uses reply name when mention omits it",
			text:     "摸",
			mentions: []platform.User{{ID: "2"}},
			replyTo:  &platform.Message{Sender: platform.User{ID: "2", DisplayName: "智智"}},
			want:     "Keigo 摸了摸智智！",
		},
		{
			name:     "action with bot mention target",
			text:     "摸",
			mentions: []platform.User{{ID: "99", DisplayName: "Bendan"}},
			want:     "Keigo 摸了摸Bendan！",
		},
		{
			name:    "action with result replying to bot targets bot",
			text:    "摸 头",
			replyTo: &platform.Message{Sender: platform.User{ID: "99", DisplayName: "Bendan"}},
			want:    "Keigo 摸了摸Bendan的头！",
		},
		{
			name:    "action with result replying to another user",
			text:    "/摸 头",
			replyTo: &platform.Message{Sender: platform.User{ID: "2", DisplayName: "智智"}},
			want:    "Keigo 摸了摸智智的头！",
		},
		{
			name:    "repeated action with result replying to another user",
			text:    "摸摸 头",
			replyTo: &platform.Message{Sender: platform.User{ID: "2", DisplayName: "智智"}},
			want:    "Keigo 摸了摸智智的头！",
		},
		{
			name:    "templated action with result replying to another user",
			text:    "喂 糖",
			replyTo: &platform.Message{Sender: platform.User{ID: "2", DisplayName: "智智"}},
			want:    "Keigo 喂了智智的糖！",
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
