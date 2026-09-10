package commands

import (
	"context"
	"math/rand"
	"regexp"

	"github.com/sxyazi/bendan/commands/yes"
	"github.com/sxyazi/bendan/platform"
)

var reMark = regexp.MustCompile(`^[?？¿‽]+$`)

func Mark(ctx context.Context, message *platform.Message) bool {
	if !reMark.MatchString(message.Text) {
		return false
	}

	text := ""
	if rand.Float64() > .9 {
		text = []string{"啊？", "嗯？"}[rand.Intn(2)]
	} else {
		text = yesSel([2][]string{{"?", "？", "¿"}, {message.Text[:rand.Intn(len(message.Text))+1]}}, &yes.Token{Sub: message.Text})
	}
	replyText(ctx, message, text)
	return true
}
