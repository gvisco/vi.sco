package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	toml "github.com/pelletier/go-toml"
)

type Config struct {
	Bot struct {
		Token string
	}
	Permissions struct {
		Owner int
	}
}

func initConfig() (*Config, error) {
	file, err := os.Open("config.toml")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	dec := toml.NewDecoder(file)
	if err := dec.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}

func initBot(config *Config) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(config.Bot.Token)
	if err != nil {
		return nil, err
	}

	return bot, nil
}

func main() {
	config, err := initConfig()
	if err != nil {
		log.Panicf("Cannot read the configuration - %s", err)
		os.Exit(1)
	}

	bot, err := initBot(config)
	if err != nil {
		log.Panicf("Cannot initialize the bot - %s", err)
		os.Exit(1)
	}

	// bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panicf("Cannot initialize the updates channel - %s", err)
		os.Exit(1)
	}

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		} else if update.Message.From.ID != config.Permissions.Owner {
			log.Printf("Ignoring message from user %s. Text: %s", update.Message.From.UserName, update.Message.Text)
			continue
		}

		log.Printf("[%s] %s", update.Message.From, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		// msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
	}
}
