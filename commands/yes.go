package commands

import (
	"context"
	"math/rand"
	"strings"

	"github.com/sxyazi/bendan/commands/yes"
	"github.com/sxyazi/bendan/platform"
)

func yesSel(options [2][]string, token *yes.Token) string {
	selected := options[0]
	if rand.Float64() > .9 {
		selected = options[rand.Int()&1]
	} else if yes.StableChoice(token) == 1 {
		selected = options[1]
	}
	return selected[rand.Intn(len(selected))]
}

func YesChoice(ctx context.Context, message *platform.Message) bool {
	token := yes.ChoiceTokenize(message.Text)
	if token == nil {
		return false
	}

	choice := token.Obj
	if yes.StableChoice(token) == 1 {
		choice = token.Ind
	}
	templates := []string{
		"{choice}",
		"{choice}！",
		"我选{choice}",
		"还是{choice}吧",
		"肯定是{choice}",
		"{left}和{right}我都要",
	}
	text := templates[rand.Intn(len(templates))]
	text = strings.NewReplacer(
		"{left}", token.Obj,
		"{right}", token.Ind,
		"{choice}", choice,
	).Replace(text)
	sendText(ctx, message.Chat, text)
	return true
}

func YesRight(ctx context.Context, message *platform.Message) bool {
	token := yes.RightTokenize(message.Text)
	if token == nil {
		return false
	}

	var options [2][]string
	switch []rune(token.Word)[0] {
	case '对':
		options = [2][]string{{"对", "对的", "emm好像挺对的"}, {"no", "不对", "不大对"}}
	case '是':
		options = [2][]string{{"是", "对", "是的", "是啊", "是诶"}, {"no", "不是", "不是啊", "不是8", "不是吧", "应该不是"}}
	case '有':
		options = [2][]string{{"有", "有的", "有啊", "有诶", "有吧"}, {"no", "没有", "没啊", "没8", "没吧", "没有吧", "应该没吧"}}
	case '行':
		options = [2][]string{{"行", "行啊", "肯定行", "我觉得行"}, {"不行", "不太行", "应该不行", "我觉得不行"}}
	default:
		switch {
		case strings.Contains(token.Word, "对"):
			options = [2][]string{{"yy", "yyy", "对的", "挺对的", "没毛病"}, {"不对", "不对啊", "不大对", "明显错了", "肯定不对啊"}}
		case strings.Contains(token.Word, "是"):
			options = [2][]string{{"好像是", "应该是", "还真是", "草，还真是"}, {"不是啊", "并不是", "显然不是", "我倒希望是"}}
		case strings.Contains(token.Word, "有"):
			options = [2][]string{{"好像有", "应该有", "还真有", "草，还真有"}, {"没有啊", "并没有", "然而并没有", "我倒希望有"}}
		default:
			options = [2][]string{{"yy", "yyy", "可以", "肯定行", "我觉得行"}, {"不行", "不太行", "应该不行", "肯定不行", "我觉得不行"}}
		}
	}
	sendText(ctx, message.Chat, yesSel(options, token))
	return true
}

func YesIs(ctx context.Context, message *platform.Message) bool {
	token := yes.IsTokenize(message.Text)
	if token == nil {
		return false
	}
	if token.Ind != "" {
		sendText(ctx, message.Chat, yesSel([2][]string{{token.Ind, token.Ind + "！"}, {token.Obj, token.Obj + "！"}}, token))
		return true
	}

	var options [2][]string
	switch token.Typ {
	case yes.TypIs:
		options = [2][]string{{"是", "是的"}, {"不是", "不是啊", "不是哦"}}
	case yes.TypHave:
		options = [2][]string{{"有", "有的", "有啊"}, {"没", "没吧", "没啊", "没有啊"}}
	case yes.TypIsYesNo:
		options = [2][]string{{"是", "是的", "yyy"}, {"no", "不是", "不是啊"}}
	case yes.TypHaveYesNo:
		options = [2][]string{{"有", "有的", "有啊"}, {"无", "没", "没有", "没啊", "并没有"}}
	case yes.TypShouldYesNo:
		switch token.Word {
		case "要不要":
			options = [2][]string{{"要", "要啊", "当然要"}, {"不要", "还是不要了", "没必要"}}
		case "该不该":
			options = [2][]string{{"该", "应该", "当然该"}, {"不该", "还是算了", "不应该"}}
		case "值不值得":
			options = [2][]string{{"值得", "当然值得"}, {"不值得", "还是算了"}}
		}
	case yes.TypHaveSo:
		options = [2][]string{{"是的", "是的捏"}, {"确实有" + token.Obj, "确实是有" + token.Obj}}
	default:
		return false
	}
	sendText(ctx, message.Chat, yesSel(options, token))
	return true
}

func YesCan(ctx context.Context, message *platform.Message) bool {
	token := yes.CanTokenize(message.Text)
	if token == nil {
		return false
	}

	var text string
	switch token.Word {
	case "能不能", "能吗", "能嘛", "能吧", "能罢":
		text = yesSel([2][]string{{"能", "能！", "能啊"}, {"不能", "不能！", "不可以！", "不，你不能"}}, token)
	case "会不会", "会吗", "会嘛", "会吧", "会罢":
		text = yesSel([2][]string{{"会", "会！", "会的"}, {"不会", "不会啊", "不会的！"}}, token)
	case "可不可以":
		text = yesSel([2][]string{{"可以", "可以啊", "当然可以"}, {"不可以", "还是不可以", "不太可以"}}, token)
	case "行不行":
		text = yesSel([2][]string{{"行", "行啊", "肯定行"}, {"不行", "不太行", "还是不行"}}, token)
	case "好不好":
		text = yesSel([2][]string{{"好", "好啊", "当然好"}, {"不好", "不太好", "还是算了"}}, token)
	}
	sendText(ctx, message.Chat, text)
	return true
}

func YesLook(ctx context.Context, message *platform.Message) bool {
	token := yes.LookTokenize(message.Text)
	if token == nil {
		return false
	}
	if rand.Float64() > .9 {
		options := []string{"你的呢", "看看你的", "can can need"}
		text := options[rand.Intn(len(options))]
		if message.ReplyTo == nil {
			sendText(ctx, message.Chat, text)
		} else {
			replyText(ctx, message.ReplyTo, text)
		}
		return true
	}

	text := yesSel([2][]string{{"看看", "想看"}, {"窝也想看", "想看，gkd"}}, token)
	if message.ReplyTo == nil {
		sendText(ctx, message.Chat, text)
	} else {
		replyText(ctx, message.ReplyTo, text)
	}
	return true
}
