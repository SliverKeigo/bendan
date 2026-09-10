package commands

import (
	"bytes"
	"context"
	"regexp"
	"strings"

	"github.com/sxyazi/bendan/commands/eval"
	"github.com/sxyazi/bendan/platform"
)

var reEval = regexp.MustCompile(`(?mi)^//\s*(go|golang|js|javascript|node|nodejs)[\s\n]+([\s\S]+)`)

// Eval executes Go or JavaScript after a double-slash command.
func Eval(ctx context.Context, message *platform.Message) bool {
	matches := reEval.FindStringSubmatch(message.Text)
	if len(matches) < 3 {
		return false
	}

	result := make(chan []string, 1)
	go func() {
		switch strings.ToLower(matches[1]) {
		case "go", "golang":
			result <- eval.NewGo().Eval(matches[2])
		case "js", "javascript", "node", "nodejs":
			result <- eval.NewNode().Eval(matches[2])
		default:
			result <- []string{"Unknown language"}
		}
	}()

	var output bytes.Buffer
	for _, line := range <-result {
		if line != "" {
			output.WriteString(line)
		}
	}
	if output.Len() == 0 {
		output.WriteString("No output")
	}
	replyText(ctx, message, output.String())
	return true
}
