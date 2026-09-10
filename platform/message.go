// Package platform defines the messaging contract used by command handlers.
package platform

// Message contains the platform-neutral fields needed by Bendan commands.
// IDs are strings because QQ and Telegram use different identifier formats.
type Message struct {
	ID                     string
	Chat                   Chat
	Sender                 User
	Text                   string
	Caption                string
	ReplyTo                *Message
	IsBot                  bool
	IsForwardedChannelPost bool
}

// Content returns all text that may contain a command or URL.
func (m *Message) Content() string {
	if m.Caption == "" {
		return m.Text
	}
	if m.Text == "" {
		return m.Caption
	}
	return m.Text + "\n" + m.Caption
}

// Chat identifies a conversation.
type Chat struct {
	ID   string
	Name string
	Kind string
}

// User identifies the author of a message.
type User struct {
	ID          string
	DisplayName string
	Username    string
	IsBot       bool
}
