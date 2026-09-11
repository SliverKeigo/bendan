package yes

import "testing"

func TestChoiceTokenize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// 群聊中常见的不带标点短句。
		{input: "猫还是狗", want: "sub=, obj=猫, ind=狗"},
		{input: "奶茶 还是 咖啡", want: "sub=, obj=奶茶, ind=咖啡"},
		{input: "去北京还是去上海呀", want: "sub=, obj=去北京, ind=去上海"},
		{input: "吃饭还是睡觉", want: "sub=, obj=吃饭, ind=睡觉"},
		{input: "坐公交还是坐地铁", want: "sub=, obj=坐公交, ind=坐地铁"},

		// 明确引导词和长句上下文。
		{input: "是猫还是狗？", want: "sub=, obj=猫, ind=狗"},
		{input: "你是猫还是狗？", want: "sub=你, obj=猫, ind=狗"},
		{input: "我好困，是现在睡呢还是待会再睡", want: "sub=我好困, obj=现在睡, ind=待会再睡"},
		{input: "最近天气一直不太稳定，周末计划出去玩的话，是去交通方便但人很多的市中心逛街呢还是去稍微远一点但比较安静的郊外公园散步", want: "sub=最近天气一直不太稳定，周末计划出去玩的话, obj=去交通方便但人很多的市中心逛街, ind=去稍微远一点但比较安静的郊外公园散步"},
		{input: "我明天早上八点有个非常重要的会议，但今晚还有很多工作没有做完，是现在先睡几个小时明早起来做呢还是今晚熬夜全部做完再睡", want: "sub=我明天早上八点有个非常重要的会议，但今晚还有很多工作没有做完, obj=现在先睡几个小时明早起来做, ind=今晚熬夜全部做完再睡"},
		{input: "你觉得在完全没有准备的情况下直接参加考试比较好，还是先延期一周认真复习之后再参加比较好", want: "sub=你觉得, obj=在完全没有准备的情况下直接参加考试比较好, ind=先延期一周认真复习之后再参加比较好"},
		{input: "到底继续维护旧系统还是重写新系统", want: "sub=到底, obj=继续维护旧系统, ind=重写新系统"},

		// 缺失选项和多选项不处理。
		{input: "还是狗？", want: ""},
		{input: "猫还是？", want: ""},
		{input: "早餐吃包子还是面条还是米粉", want: ""},

		// “还是”表示“仍然”或属于陈述性关联结构时不处理。
		{input: "无论是现在开始还是明天开始，只要最后做完就行", want: ""},
		{input: "不管你选择坐公交还是坐地铁，我都会在终点等你", want: ""},
		{input: "不论晴天还是雨天都要训练", want: ""},
		{input: "与其纠结喝咖啡还是喝茶，不如直接喝水", want: ""},
		{input: "他还是决定明天再说", want: ""},
		{input: "我还是觉得不太行", want: ""},
		{input: "你还是先休息吧", want: ""},
		{input: "最后还是选择了放弃", want: ""},
		{input: "这个功能还是挺好用的", want: ""},
		{input: "但是AA还是个BB", want: ""},
	}

	for _, test := range tests {
		if got := ChoiceTokenize(test.input); got.String() != test.want {
			t.Errorf("ChoiceTokenize(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}
