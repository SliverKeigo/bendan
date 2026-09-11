package commands

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

//go:embed actions.json
var actionData []byte

type actionLexicon struct {
	Chinese map[string]string `json:"zh"`
	Latin   map[string]string `json:"latin"`
}

var actions actionLexicon

func init() {
	if err := json.Unmarshal(actionData, &actions); err != nil {
		panic(fmt.Sprintf("load action lexicon: %v", err))
	}
}

func isAction(action string) bool {
	if action == "" {
		return false
	}
	if _, ok := actions.Chinese[action]; ok {
		return true
	}
	if strings.HasSuffix(action, "了") {
		_, ok := actions.Chinese[strings.TrimSuffix(action, "了")]
		return ok
	}
	_, ok := actions.Latin[strings.ToLower(action)]
	return ok
}

func isSlashAction(action string) bool {
	if isAction(action) {
		return true
	}
	runes := []rune(action)
	return len(runes) > 0 && !unicode.IsLetter(runes[0]) && !unicode.IsNumber(runes[0])
}

func actionDisplay(action string) string {
	if display, ok := actions.Chinese[action]; ok {
		return display
	}
	if strings.HasSuffix(action, "了") {
		if _, ok := actions.Chinese[strings.TrimSuffix(action, "了")]; ok {
			return action
		}
	}
	if display, ok := actions.Latin[strings.ToLower(action)]; ok {
		return display
	}
	return action
}
