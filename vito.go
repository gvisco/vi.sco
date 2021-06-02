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
		Owner   int
		Allowed []int
	}
}

type ChatChannel struct {
	channel chan *tgbotapi.Update
	chatId  int64
	config  *Config
}

func NewChatChannel(chatId int64, config *Config) *ChatChannel {
	cc := &ChatChannel{}
	cc.channel = make(chan *tgbotapi.Update)
	cc.chatId = chatId
	cc.config = config
	return cc
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

func isAllowed(id int, config *Config) bool {
	for _, a := range config.Permissions.Allowed {
		if a == id {
			return true
		}
	}
	return false
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

	channels := make(map[int64]*ChatChannel)
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		} else if isAllowed(update.Message.From.ID, config) {
			msg := update.Message
			log.Printf("[Processing] User {%s} Text {%s} Chat {%d}", msg.From, msg.Text, msg.Chat.ID)
			chatChannel := channels[msg.Chat.ID]
			if chatChannel == nil {
				log.Printf("[Init chat channel] Chat {%d}", msg.Chat.ID)
				chatChannel = NewChatChannel(msg.Chat.ID, config)
				channels[msg.Chat.ID] = chatChannel
				go func(chatChannel *ChatChannel) {
					for {
						upd := <-chatChannel.channel
						msg := tgbotapi.NewMessage(upd.Message.Chat.ID, upd.Message.Text)
						bot.Send(msg)
					}
				}(chatChannel)
			}
			chatChannel.channel <- &update

		} else {
			log.Printf("[Ignoring] User {%s} UserId {%d} Text {%s} Chat {%d}", update.Message.From,
				update.Message.From.ID, update.Message.Text, update.Message.Chat.ID)
		}
	}
}
