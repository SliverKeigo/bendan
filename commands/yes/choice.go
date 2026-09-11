package yes

import (
	"fmt"
	"regexp"
	"strings"
)

var reChoice = regexp.MustCompile(fmt.Sprintf(`^\s*(.+?)\s*(?:%s)*还是\s*(.+?)\s*(?:%s)*$`, marks, marks))

// ChoiceTokenize parses a two-option question such as "猫还是狗？".
func ChoiceTokenize(s string) *Token {
	if token := IsTokenize(s); token != nil && token.Ind != "" {
		return &Token{Typ: TypChoice, Sub: token.Sub, Obj: token.Obj, Ind: token.Ind, Word: "还是"}
	}
	trimmed := strings.TrimSpace(s)
	if strings.HasPrefix(trimmed, "但是") || strings.HasSuffix(trimmed, "吧") || strings.HasSuffix(trimmed, "罢") {
		return nil
	}

	ms := reChoice.FindStringSubmatch(s)
	if ms == nil {
		return nil
	}

	left := strings.Trim(strings.TrimSpace(ms[1]), marks)
	right := strings.Trim(strings.TrimSpace(ms[2]), marks)
	if left == "" || right == "" || reDeterminer.MatchString(left) || reDeterminer.MatchString(right) {
		return nil
	}

	return &Token{Typ: TypChoice, Obj: left, Ind: right, Word: "还是"}
}
