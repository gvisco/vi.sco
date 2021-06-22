package main

import (
	"flag"
	"log"
	"os"

	"github.com/gvisco/vi.sco/pkg/bots/gottolists"
	"github.com/gvisco/vi.sco/pkg/gotto"
	// "github.com/gvisco/vi.sco/pkg/gotto/sample/echo"
)

func main() {
	config := flag.String("config", "./config.toml", "the .toml configuration file path")
	flag.Parse()

	bot, err := gotto.NewGotto(config)
	if err != nil {
		log.Panicf("Cannot initialize the bot - %s", err)
		os.Exit(1)
	}
	// bot.RegisterBot(echo.NewFactory())
	bot.RegisterBot(gottolists.NewFactory())
	bot.Start()
}
