package gotto

import (
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	toml "github.com/pelletier/go-toml"
)

type Gotto struct {
	tgbot         *tgbotapi.BotAPI
	config        *Config
	conversations map[int64]*Conversation
	factories     []GottoBotFactory
}

type GottoBotFactory interface {
	CreateBot(workspace string) (GottoBot, error)
}

type GottoBot interface {
	OnUpdate(userId string, userName string, message string) string
}

type Config struct {
	Bot struct {
		Token string
	}
	Permissions struct {
		Allowed []int
	}
}

type Conversation struct {
	channel   chan *tgbotapi.Update
	chatId    int64
	config    *Config
	workspace string
	bots      []GottoBot
}

func (engine *Gotto) newConversation(chatId int64) (*Conversation, error) {
	cc := &Conversation{}
	cc.channel = make(chan *tgbotapi.Update)
	cc.chatId = chatId
	cc.config = engine.config
	cc.bots = []GottoBot{}
	// create the workspace
	workspace := "./workspace/" + fmt.Sprint(chatId)
	err := os.MkdirAll(workspace, os.ModePerm)
	if err != nil {
		log.Printf("[ERROR Cannot create workspace] ChatId {%d} Workspace {%s}", chatId, workspace)
		return nil, fmt.Errorf("Cannot create workspace dir - %s", err)
	}
	cc.workspace = workspace
	// initialize individual bots
	for _, f := range engine.factories {
		bot, err := f.CreateBot(cc.workspace)
		if err != nil {
			log.Printf("[ERROR Cannot initialize bot] BotFactory {%+v} ChatId {%d} Workspace {%s}", f, cc.chatId, cc.workspace)
			continue
		}
		cc.bots = append(cc.bots, bot)
	}
	// start message dispatching
	go func(conversation *Conversation) {
		for {
			upd := <-conversation.channel
			for _, bot := range conversation.bots {
				reply := bot.OnUpdate(fmt.Sprint(upd.Message.From.ID), fmt.Sprint(upd.Message.From), upd.Message.Text)
				if reply != "" {
					msg := tgbotapi.NewMessage(conversation.chatId, reply)
					engine.tgbot.Send(msg)
				}
			}
		}
	}(cc)

	return cc, nil
}

func (engine *Gotto) getConversation(chatId int64) (*Conversation, error) {
	conversation := engine.conversations[chatId]
	if conversation == nil {
		log.Printf("[Init chat channel] Chat {%d}", chatId)
		var err error
		conversation, err = engine.newConversation(chatId)
		if err != nil {
			return nil, err
		}
		engine.conversations[chatId] = conversation
	}
	return conversation, nil
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

func (config *Config) isAllowed(id int) bool {
	for _, a := range config.Permissions.Allowed {
		if a == id {
			return true
		}
	}
	return false
}

func initBot(config *Config) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(config.Bot.Token)
	if err != nil {
		return nil, err
	}

	return bot, nil
}

func NewGotto() (*Gotto, error) {
	config, err := initConfig()
	if err != nil {
		log.Printf("Cannot read the configuration - %s", err)
		return nil, err
	}

	bot, err := initBot(config)
	if err != nil {
		log.Printf("Cannot initialize the bot - %s", err)
		return nil, err
	}

	// bot.Debug = true

	log.Printf("Initialized bot on account %s", bot.Self.UserName)

	return &Gotto{
		tgbot:         bot,
		config:        config,
		conversations: make(map[int64]*Conversation),
		factories:     []GottoBotFactory{},
	}, nil
}

func (engine *Gotto) RegisterBot(factory GottoBotFactory) {
	engine.factories = append(engine.factories, factory)
}

func (engine *Gotto) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := engine.tgbot.GetUpdatesChan(u)
	if err != nil {
		log.Printf("Cannot initialize the updates channel - %s", err)
		os.Exit(1)
	}

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		} else if engine.config.isAllowed(update.Message.From.ID) {
			msg := update.Message
			log.Printf("[Processing] User {%s} Text {%s} Chat {%d}", msg.From, msg.Text, msg.Chat.ID)
			conversation, err := engine.getConversation(msg.Chat.ID)
			if err != nil {
				log.Printf("[ERROR Cannot get Conversation] Chat {%d}", msg.Chat.ID)
				continue
			}
			// dispatch the update to the right conversation
			conversation.channel <- &update
		} else {
			log.Printf("[Ignoring] User {%s} UserId {%d} Text {%s} Chat {%d}", update.Message.From,
				update.Message.From.ID, update.Message.Text, update.Message.Chat.ID)
		}
	}
}
