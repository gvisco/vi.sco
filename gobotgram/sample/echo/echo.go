package echo

import (
	"log"

	"github.com/gvisco/vi.sco/gobotgram"
)

type EchoBotFactory struct {
}

type EchoBot struct {
	workspace string
}

func NewFactory() *EchoBotFactory {
	return &EchoBotFactory{}
}

func (*EchoBotFactory) CreateBot(workspace string) (gobotgram.GobotgramBot, error) {
	log.Printf("[New EchoBot created] Workspace {%s}", workspace)
	return EchoBot{workspace: workspace}, nil
}

func (EchoBot) OnUpdate(userId string, userName string, message string) string {
	return message
}
