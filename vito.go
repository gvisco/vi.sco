package main

import (
	"log"
	"os"

	"github.com/gvisco/vi.sco/bots/gottolists"
	"github.com/gvisco/vi.sco/gotto"
	// "github.com/gvisco/vi.sco/gotto/sample/echo"
)

func main() {
	bot, err := gotto.NewGotto()
	if err != nil {
		log.Panicf("Cannot initialize the bot - %s", err)
		os.Exit(1)
	}
	// bot.RegisterBot(echo.NewFactory())
	bot.RegisterBot(gottolists.NewFactory())
	bot.Start()
}
