package commands

import (
	"context"
	"testing"

	"github.com/sxyazi/bendan/platform"
)

type collectingReplyRecorder struct {
	events []AutomaticReplyEvent
}

func (r *collectingReplyRecorder) Record(event AutomaticReplyEvent) {
	r.events = append(r.events, event)
}

func TestHandleRecordsSuccessfulAutomaticReply(t *testing.T) {
	resetMessageGuards(t)
	previousBot := Bot
	previousRecorder := automaticReplyRecorder
	previousSalt := automaticReplyHashSalt
	t.Cleanup(func() {
		Bot = previousBot
		SetAutomaticReplyRecorder(previousRecorder, previousSalt)
	})

	bot := &recordingBot{}
	recorder := &collectingReplyRecorder{}
	Bot = bot
	SetAutomaticReplyRecorder(recorder, "test-salt")

	message := &platform.Message{
		ID:     "message-1",
		Chat:   platform.Chat{ID: "group-1", Kind: "group"},
		Sender: platform.User{ID: "user-1"},
		Text:   "猫还是狗",
	}
	Handle(context.Background(), message)

	if len(recorder.events) != 1 {
		t.Fatalf("recorded events = %d, want 1", len(recorder.events))
	}
	event := recorder.events[0]
	if event.Handler != "yes_choice" || event.InputText != "猫还是狗" || event.ReplyText == "" {
		t.Fatalf("unexpected event: %#v", event)
	}
	if event.ChatHash == "" || event.ChatHash == message.Chat.ID {
		t.Fatalf("chat hash was not anonymized: %q", event.ChatHash)
	}
	if event.SenderHash == "" || event.SenderHash == message.Sender.ID {
		t.Fatalf("sender hash was not anonymized: %q", event.SenderHash)
	}
}

func TestDirectCommandIsNotRecordedAsAutomaticReply(t *testing.T) {
	resetMessageGuards(t)
	previousBot := Bot
	previousRecorder := automaticReplyRecorder
	previousSalt := automaticReplyHashSalt
	t.Cleanup(func() {
		Bot = previousBot
		SetAutomaticReplyRecorder(previousRecorder, previousSalt)
	})

	bot := &recordingBot{}
	recorder := &collectingReplyRecorder{}
	Bot = bot
	SetAutomaticReplyRecorder(recorder, "test-salt")

	Handle(context.Background(), &platform.Message{
		ID:     "message-2",
		Chat:   platform.Chat{ID: "group-1", Kind: "group"},
		Sender: platform.User{ID: "user-1"},
		Text:   "/status",
	})

	if len(recorder.events) != 0 {
		t.Fatalf("recorded events = %d, want 0", len(recorder.events))
	}
}
