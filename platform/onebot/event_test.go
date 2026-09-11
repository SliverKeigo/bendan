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

func TestEventPreservesAtMentionTarget(t *testing.T) {
	var event Event
	if err := json.Unmarshal([]byte(`{
		"post_type":"message",
		"message_type":"group",
		"self_id":422345383,
		"message_id":67890,
		"user_id":1226355793,
		"group_id":744196849,
		"sender":{"user_id":1226355793,"nickname":"Keigo","card":""},
		"message":[
			{"type":"text","data":{"text":"摸 "}},
			{"type":"at","data":{"qq":"422345383","name":"Bendan"}}
		]
	}`), &event); err != nil {
		t.Fatal(err)
	}

	message := event.ToPlatformMessage()
	if message == nil || len(message.Mentions) != 1 {
		t.Fatalf("mentions = %#v, want one mention", message)
	}
	if message.Text != "摸" || message.Mentions[0].ID != "422345383" || message.Mentions[0].DisplayName != "Bendan" {
		t.Fatalf("unexpected mention message: %#v", message)
	}
}

func TestEventFallsBackToNameFromAtSegmentText(t *testing.T) {
	var event Event
	if err := json.Unmarshal([]byte(`{
		"post_type":"message",
		"message_type":"group",
		"self_id":422345383,
		"message_id":67890,
		"user_id":1226355793,
		"group_id":744196849,
		"sender":{"user_id":1226355793,"nickname":"Keigo"},
		"message":[
			{"type":"text","data":{"text":"摸  "}},
			{"type":"at","data":{"qq":"3889000871","name":"","text":"@战地1小电视"}},
			{"type":"text","data":{"text":"  \n"}}
		]
	}`), &event); err != nil {
		t.Fatal(err)
	}

	message := event.ToPlatformMessage()
	if message == nil || len(message.Mentions) != 1 {
		t.Fatalf("unexpected message: %#v", message)
	}
	if got, want := message.Mentions[0].DisplayName, "战地1小电视"; got != want {
		t.Fatalf("mention display name = %q, want %q", got, want)
	}
}

func TestEventIgnoresAtAllAndPreservesNamedMentions(t *testing.T) {
	var event Event
	if err := json.Unmarshal([]byte(`{
		"post_type":"message",
		"message_type":"group",
		"self_id":422345383,
		"message_id":67890,
		"user_id":1226355793,
		"group_id":744196849,
		"sender":{"user_id":1226355793,"nickname":"Keigo"},
		"message":[
			{"type":"at","data":{"qq":"all"}},
			{"type":"text","data":{"text":" 摸 "}},
			{"type":"at","data":{"qq":"1797580779","name":"WuWa_MT9985.skill"}}
		]
	}`), &event); err != nil {
		t.Fatal(err)
	}

	message := event.ToPlatformMessage()
	if message == nil || message.Text != "摸" || len(message.Mentions) != 1 {
		t.Fatalf("unexpected message: %#v", message)
	}
	if message.Mentions[0].ID != "1797580779" || message.Mentions[0].DisplayName != "WuWa_MT9985.skill" {
		t.Fatalf("unexpected mentions: %#v", message.Mentions)
	}
}

func TestEventResolvesReplySegmentToBotMessage(t *testing.T) {
	var event Event
	if err := json.Unmarshal([]byte(`{
		"post_type":"message",
		"message_type":"group",
		"self_id":422345383,
		"message_id":67890,
		"user_id":1226355793,
		"group_id":891343531,
		"sender":{"user_id":1226355793,"nickname":"Keigo"},
		"message":[
			{"type":"reply","data":{"id":"123"}},
			{"type":"text","data":{"text":"闭嘴"}}
		]
	}`), &event); err != nil {
		t.Fatal(err)
	}

	message := event.ToPlatformMessage()
	if message == nil {
		t.Fatal("expected platform message")
	}
	if message.ReplyTo == nil {
		t.Fatal("reply target is nil, want reply segment target")
	}
	if message.ReplyTo.ID != "123" {
		t.Fatalf("reply target ID = %q, want 123", message.ReplyTo.ID)
	}
}

func TestEventUsesTextSegmentsWhenReplyAddsAnAtMention(t *testing.T) {
	var event Event
	if err := json.Unmarshal([]byte(`{
		"post_type":"message",
		"message_type":"group",
		"self_id":422345383,
		"message_id":67890,
		"user_id":1226355793,
		"group_id":744196849,
		"raw_message":"[CQ:reply,id=123][CQ:at,qq=422345383] /摸",
		"sender":{"user_id":1226355793,"nickname":"Keigo","card":""},
		"message":[
			{"type":"reply","data":{"id":"123"}},
			{"type":"at","data":{"qq":"422345383"}},
			{"type":"text","data":{"text":" /摸"}}
		]
	}`), &event); err != nil {
		t.Fatal(err)
	}

	message := event.ToPlatformMessage()
	if message == nil {
		t.Fatal("expected message")
	}
	if message.Text != "/摸" {
		t.Fatalf("text = %q, want command text without reply and mention segments", message.Text)
	}
}
