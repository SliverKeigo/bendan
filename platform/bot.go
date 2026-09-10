package platform

import "context"

// Bot contains the operations that command handlers may perform.
// Unsupported operations must return ErrUnsupported rather than silently fail.
type Bot interface {
	Identity() User
	SendText(context.Context, Chat, string) (*Message, error)
	ReplyText(context.Context, *Message, string) (*Message, error)
	DeleteMessage(context.Context, *Message) error
	PinMessage(context.Context, *Message) error
	UnpinMessage(context.Context, Chat, string) error
}

// Capabilities records platform support for features that cannot be expressed
// consistently across Telegram and OneBot.
type Capabilities struct {
	CanDeleteMessage bool
	CanEditMessage   bool
	CanPinMessage    bool
	HasInlineQuery   bool
}

// CapabilityProvider is optional because a command should not need platform
// details unless it uses a non-portable feature.
type CapabilityProvider interface {
	Capabilities() Capabilities
}
