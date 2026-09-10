package commands

import (
	"context"
	"sync"

	"github.com/sxyazi/bendan/platform"
)

type recordingBot struct {
	mu       sync.Mutex
	identity platform.User
	sent     []string
	replied  []string
}

func (b *recordingBot) Identity() platform.User { return b.identity }

func (b *recordingBot) SendText(_ context.Context, _ platform.Chat, text string) (*platform.Message, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.sent = append(b.sent, text)
	return &platform.Message{ID: "1"}, nil
}

func (b *recordingBot) ReplyText(_ context.Context, _ *platform.Message, text string) (*platform.Message, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.replied = append(b.replied, text)
	return &platform.Message{ID: "2"}, nil
}

func (*recordingBot) DeleteMessage(context.Context, *platform.Message) error { return nil }
func (*recordingBot) PinMessage(context.Context, *platform.Message) error {
	return platform.ErrUnsupported
}
func (*recordingBot) UnpinMessage(context.Context, platform.Chat, string) error {
	return platform.ErrUnsupported
}

func withTestBot(t interface{ Cleanup(func()) }, bot platform.Bot) {
	previous := Bot
	Bot = bot
	t.Cleanup(func() { Bot = previous })
}
