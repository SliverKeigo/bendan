package onebot

import (
	"encoding/json"
	"testing"
)

func TestEventToPlatformMessage(t *testing.T) {
	const raw = `{
		"post_type":"message",
		"message_type":"group",
		"self_id":12345,
		"message_id":67890,
		"user_id":10001,
		"group_id":20002,
		"raw_message":"/me 喝茶",
		"sender":{"user_id":10001,"nickname":"Keigo","card":""},
		"message":[{"type":"text","data":{"text":"/me 喝茶"}}]
	}`

	var event Event
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		t.Fatal(err)
	}
	message := event.ToPlatformMessage()
	if message == nil {
		t.Fatal("expected platform message")
	}
	if message.ID != "67890" || message.Chat.ID != "20002" || message.Sender.ID != "10001" {
		t.Fatalf("unexpected message identity: %#v", message)
	}
	if message.Sender.DisplayName != "Keigo" || message.Text != "/me 喝茶" {
		t.Fatalf("unexpected message content: %#v", message)
	}
}

func TestEventToPlatformPrivateMessage(t *testing.T) {
	var event Event
	if err := json.Unmarshal([]byte(`{
		"post_type":"message",
		"message_type":"private",
		"self_id":12345,
		"message_id":1,
		"user_id":2,
		"raw_message":"hi",
		"sender":{"nickname":"Keigo"},
		"message":"hi"
	}`), &event); err != nil {
		t.Fatal(err)
	}

	message := event.ToPlatformMessage()
	if message == nil || message.Chat.ID != "2" || message.Chat.Name != "Keigo" || message.Text != "hi" {
		t.Fatalf("unexpected private message: %#v", message)
	}
}

func TestEventRejectsNonMessageEvents(t *testing.T) {
	event := Event{PostType: "notice", MessageType: "group"}
	if event.ToPlatformMessage() != nil {
		t.Fatal("expected nil for non-message event")
	}
}
