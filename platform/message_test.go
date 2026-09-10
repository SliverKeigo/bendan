package platform

import "testing"

func TestMessageContent(t *testing.T) {
	tests := []struct {
		name string
		msg  Message
		want string
	}{
		{name: "text", msg: Message{Text: "hello"}, want: "hello"},
		{name: "caption", msg: Message{Caption: "hello"}, want: "hello"},
		{name: "both", msg: Message{Text: "hello", Caption: "world"}, want: "hello\nworld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.msg.Content(); got != tt.want {
				t.Fatalf("Content() = %q, want %q", got, tt.want)
			}
		})
	}
}
