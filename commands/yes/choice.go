package yes

import "strings"

const choiceTrimCutset = " \t\r\n啊阿呀吗嘛呢捏,.?!;，。？！；"

var choiceStatementPrefixes = []string{
	"无论", "不论", "不管", "与其", "即使", "哪怕", "但是", "可是", "然而",
}

var choicePromptPrefixes = []string{
	"大家觉得", "你们觉得", "你觉得", "大家认为", "你们认为", "你认为",
	"你们说", "你说", "你们看", "你看", "到底", "究竟", "选", "选择",
}

var choiceAdverbRights = []string{
	"觉得", "认为", "决定", "选择了", "选了", "还是", "会", "要", "得",
	"应该", "先", "继续", "挺", "很", "比较", "不太", "仍然", "仍旧",
	"依然", "最好", "终于", "又",
}

// ChoiceTokenize parses a two-option expression around the conjunction "还是".
// It does not require punctuation because casual chat questions commonly omit it.
func ChoiceTokenize(s string) *Token {
	text := strings.TrimSpace(s)
	if strings.Count(text, "还是") != 1 {
		return nil
	}
	for _, prefix := range choiceStatementPrefixes {
		if strings.HasPrefix(text, prefix) {
			return nil
		}
	}

	parts := strings.SplitN(text, "还是", 2)
	leftRaw := strings.TrimSpace(parts[0])
	right := trimChoiceOption(parts[1])
	if leftRaw == "" || right == "" {
		return nil
	}

	var context, left string
	if marker := strings.LastIndex(leftRaw, "是"); marker >= 0 {
		context = trimChoiceContext(leftRaw[:marker])
		left = trimChoiceOption(leftRaw[marker+len("是"):])
	} else if prompt, rest := splitChoicePrompt(leftRaw); prompt != "" {
		context = prompt
		left = trimChoiceOption(rest)
	} else {
		left = trimChoiceOption(leftRaw)
		if likelyAdverbialStill(left, right) {
			return nil
		}
	}

	if left == "" || right == "" || reDeterminer.MatchString(left) || reDeterminer.MatchString(right) {
		return nil
	}
	return &Token{Typ: TypChoice, Sub: context, Obj: left, Ind: right, Word: "还是"}
}

func splitChoicePrompt(s string) (string, string) {
	for _, prefix := range choicePromptPrefixes {
		if strings.HasPrefix(s, prefix) {
			return prefix, strings.TrimSpace(s[len(prefix):])
		}
	}
	return "", s
}

func trimChoiceOption(s string) string {
	return strings.Trim(strings.TrimSpace(s), choiceTrimCutset)
}

func trimChoiceContext(s string) string {
	return strings.Trim(strings.TrimSpace(s), " \t\r\n,.?!;:，。？！；：")
}

func likelyAdverbialStill(left, right string) bool {
	badRight := false
	for _, prefix := range choiceAdverbRights {
		if strings.HasPrefix(right, prefix) {
			badRight = true
			break
		}
	}
	if !badRight {
		// “这/那还是……的”通常是在表达“仍然处于某种状态”，而不是二选一。
		if (left == "这" || left == "那") && strings.HasSuffix(right, "的") {
			return true
		}
		return false
	}

	if left == "我" || left == "你" || left == "他" || left == "她" || left == "它" ||
		left == "我们" || left == "你们" || left == "他们" || left == "她们" || left == "它们" ||
		left == "最后" || left == "最终" || left == "后来" || left == "结果" {
		return true
	}
	if (strings.HasPrefix(left, "这") || strings.HasPrefix(left, "那")) && len([]rune(left)) <= 12 {
		return true
	}
	for _, subject := range []string{"我", "你", "他", "她", "它", "我们", "你们", "他们", "她们", "它们"} {
		if strings.HasPrefix(left, subject) && len([]rune(left)) <= 12 {
			return true
		}
	}
	return false
}
