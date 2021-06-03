package main

import (
	"log"
	"os"

	"github.com/gvisco/vi.sco/gobotgram"
	"github.com/gvisco/vi.sco/gobotgram/sample/echo"
)

func main() {
	bot, err := gobotgram.NewGobotgram()
	if err != nil {
		log.Panicf("Cannot initialize the bot - %s", err)
		os.Exit(1)
	}
	bot.RegisterBot(echo.NewFactory())
	bot.Start()
}
