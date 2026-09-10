package commands

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/sxyazi/bendan/platform"
)

var seDontworry = []string{
	"$0，努力过了，很厉害了",
	"$0，努力假装是人类，很厉害了",
	"$0，相信自己的感受，你的感觉是对的，很厉害了",
	"$0，鼓起勇气去做，一小步也很厉害了",
	"$0，没进步也没关系，很厉害了",
	"$0，还活着已经很厉害了",
	"$0，一分钟也很厉害了",
	"$0，能发泄出来就很厉害了",
	"$0，已经很努力了，很厉害了",
	"$0，保持清醒也很厉害了",
	"$0，有勇气say hi已经很厉害了",
	"$0，写不出来但还在坚持写，很厉害了",
	"$0，写下TODO已经很厉害了",
	"$0，建好文件夹已经很厉害了",
	"$0，感到沮丧也没关系，很厉害了",
	"$0，用简单模式通关也很厉害了",
	"$0，好好休息吧，已经很厉害了",
	"$0，能和同事沟通已经很厉害了",
	"$0，能打开VSCode已经很厉害了",
	"$0，去面过了，很厉害了",
	"$0，疯狂焦虑也没有放弃，很厉害了",
	"$0，迈出了第一步，很厉害了",
	"$0，都会过去的",
	"$0，不想坚持下去了也没关系，辛苦了",
	"$0，辛苦了",
	"$0，不热爱生活也可以的，你已经很厉害了",
	"$0，逃避虽可耻但有用",
	"$0，就算没结果，但努力过了，很厉害了",
	"$0，和花花草草说说话，已经很厉害了",
	"$0，能知道自己喜欢什么，很厉害了",
	"$0，能说出想说的，很厉害了",
	"$0，下定决心躺平已经很厉害了",
	"$0，没成功也尝试过了，很厉害了",
	"$0，不努力也可以的，很厉害了",
	"$0，做不到还在坚持，已经很厉害了",
	"$0，有时间能玩游戏，很厉害了",
	"$0，不要因为无法改变的事惩罚自己",
	"$0，工作真的很累，真的辛苦了",
	"$0，带着面具活下去也已经很厉害了",
	"$0，明天再学吧，很厉害了",
	"$0，学不动了也很厉害了",
	"$0，忘了就忘了，重要的东西一定会再次出现",
	"$0，会写hello world已经很厉害了",
	"$0，坚持到现在了，很厉害了",
	"$0，心情很容易被影响不是你的错，很厉害了",
	"$0，还在对这个世界有所期待，很厉害了",
	"$0，认识到自己也有做不到的事情，很厉害了",
	"$0，做了比没做好，很厉害了",
	"$0，学习真的很难，真的辛苦了",
	"$0，意识到自己累了，很厉害了",
	"$0，你努力的样子真的很酷",
	"紧紧握住彼此的手吧，$0，都很厉害了",
	"$0，对人性仍抱有期待，已经很厉害了",
}

func Dontworry(ctx context.Context, message *platform.Message) bool {
	var reply string
	switch message.Text {
	case "/没关系", "/没事的":
		reply = strings.TrimPrefix(message.Text, "/")
	default:
		return false
	}

	selectText := func(userID string) string {
		_, minute, _ := time.Now().Clock()
		sum := sha1.Sum([]byte(fmt.Sprintf("%s %d", userID, minute)))
		index := int(binary.BigEndian.Uint16(sum[:])) % len(seDontworry)
		return strings.Replace(seDontworry[index], "$0", reply, 1)
	}

	if message.ReplyTo == nil {
		sendText(ctx, message.Chat, selectText(message.Sender.ID))
	} else {
		replyText(ctx, message.ReplyTo, selectText(message.ReplyTo.Sender.ID))
	}
	return true
}
