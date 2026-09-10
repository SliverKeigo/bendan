// Package onebot adapts OneBot v11 events and actions to Bendan's platform contract.
package onebot

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sxyazi/bendan/platform"
)

// Event is the subset of OneBot v11 message events needed by Bendan.
type Event struct {
	PostType    string  `json:"post_type"`
	MessageType string  `json:"message_type"`
	SubType     string  `json:"sub_type"`
	Time        int64   `json:"time"`
	SelfID      int64   `json:"self_id"`
	MessageID   int64   `json:"message_id"`
	UserID      int64   `json:"user_id"`
	GroupID     int64   `json:"group_id"`
	Message     Message `json:"message"`
	RawMessage  string  `json:"raw_message"`
	Sender      Sender  `json:"sender"`
	Anonymous   any     `json:"anonymous"`
	Reply       *Event  `json:"reply"`
	ReplyTo     *Event  `json:"reply_to"`
}

// Message preserves text segments while tolerating NapCat's string or array wire format.
type Message struct {
	Text string
}

func (m *Message) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		m.Text = text
		return nil
	}

	var segments []struct {
		Type string `json:"type"`
		Data struct {
			Text string `json:"text"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &segments); err != nil {
		return fmt.Errorf("decode OneBot message: %w", err)
	}

	var textBuilder strings.Builder
	for _, segment := range segments {
		if segment.Type == "text" {
			textBuilder.WriteString(segment.Data.Text)
		}
	}
	m.Text = textBuilder.String()
	return nil
}

// Sender identifies a OneBot event sender.
type Sender struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Card     string `json:"card"`
}

// IsMessage reports whether this is a message event handled by Bendan.
func (e Event) IsMessage() bool {
	return e.PostType == "message" && (e.MessageType == "group" || e.MessageType == "private")
}

// ToPlatformMessage converts the OneBot event into Bendan's domain model.
func (e Event) ToPlatformMessage() *platform.Message {
	if !e.IsMessage() {
		return nil
	}

	chatID := e.UserID
	chatName := e.Sender.Nickname
	chatKind := "private"
	if e.MessageType == "group" {
		chatID = e.GroupID
		chatName = ""
		chatKind = "group"
	}

	text := e.RawMessage
	if text == "" {
		text = e.Message.Text
	}

	return &platform.Message{
		ID:      fmt.Sprintf("%d", e.MessageID),
		Chat:    platform.Chat{ID: fmt.Sprintf("%d", chatID), Name: chatName, Kind: chatKind},
		Sender:  e.Sender.toPlatformUser(e.UserID),
		Text:    text,
		ReplyTo: e.replyMessage(),
		IsBot:   e.UserID == e.SelfID,
	}
}

func (e Event) replyMessage() *platform.Message {
	if e.Reply != nil {
		return e.Reply.ToPlatformMessage()
	}
	if e.ReplyTo != nil {
		return e.ReplyTo.ToPlatformMessage()
	}
	return nil
}

func (s Sender) toPlatformUser(userID int64) platform.User {
	name := s.Card
	if name == "" {
		name = s.Nickname
	}
	return platform.User{ID: fmt.Sprintf("%d", userID), DisplayName: name}
}
