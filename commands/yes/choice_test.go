package yes

import "testing"

func TestChoiceTokenize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "猫还是狗？", want: "sub=, obj=猫, ind=狗"},
		{input: "奶茶 还是 咖啡", want: "sub=, obj=奶茶, ind=咖啡"},
		{input: "去北京还是去上海呀？", want: "sub=, obj=去北京, ind=去上海"},
		{input: "还是狗？", want: ""},
		{input: "猫还是？", want: ""},
		{input: "是猫还是狗？", want: "sub=, obj=猫, ind=狗"},
		{input: "你是猫还是狗？", want: "sub=你, obj=猫, ind=狗"},
		{input: "但是AA还是个BB", want: ""},
	}

	for _, test := range tests {
		if got := ChoiceTokenize(test.input); got.String() != test.want {
			t.Errorf("ChoiceTokenize(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}
