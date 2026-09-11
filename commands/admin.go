package commands

import (
	"github.com/sxyazi/bendan/platform"
	"github.com/sxyazi/bendan/utils"
)

func isAdministrator(message *platform.Message) bool {
	administratorQQ := utils.Config("administrator_qq")
	return message != nil && administratorQQ != "" && message.Sender.ID == administratorQQ
}
