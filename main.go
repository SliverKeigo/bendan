package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sxyazi/bendan/commands"
	"github.com/sxyazi/bendan/platform/onebot"
	"github.com/sxyazi/bendan/utils"
)

func main() {
	endpoint := utils.Config("onebot_ws_url")
	if endpoint == "" {
		log.Fatal("onebot_ws_url is required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	bot := onebot.NewClient(endpoint, utils.Config("onebot_access_token"))
	commands.Bot = bot
	if err := bot.Run(ctx, commands.Handle); err != nil {
		log.Fatal(err)
	}
}
