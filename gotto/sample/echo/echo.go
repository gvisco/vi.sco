package echo

import (
	"log"

	"github.com/gvisco/vi.sco/gotto"
)

type EchoBotFactory struct {
}

type EchoBot struct {
	workspace string
}

func NewFactory() *EchoBotFactory {
	return &EchoBotFactory{}
}

func (*EchoBotFactory) CreateBot(workspace string) (gotto.GottoBot, error) {
	log.Printf("[New EchoBot created] Workspace {%s}", workspace)
	return &EchoBot{workspace: workspace}, nil
}

func (bot *EchoBot) OnUpdate(userId string, userName string, message string) string {
	return message
}
