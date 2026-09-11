package commands

import "github.com/sxyazi/bendan/platform"

const administratorQQ = "1226355793"

func isAdministrator(message *platform.Message) bool {
	return message != nil && message.Sender.ID == administratorQQ
}
