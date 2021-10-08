package pkg

import (
	"fmt"
	"github.com/raf924/bot/pkg/bot/command"
	"github.com/raf924/bot/pkg/domain"
	"strings"
)

var _ command.Command = (*TellCommand)(nil)

func NewTellCommand(private bool) command.Command {
	return &TellCommand{private: private}
}

type TellCommand struct {
	command.NoOpInterceptor
	bot     command.Executor
	private bool
}

func (t *TellCommand) Init(bot command.Executor) error {
	t.bot = bot
	return nil
}

func (t *TellCommand) Name() string {
	if t.private {
		return "whisper"
	}
	return "tell"
}

func (t *TellCommand) Aliases() []string {
	return nil
}

func (t *TellCommand) Execute(command *domain.CommandMessage) ([]*domain.ClientMessage, error) {
	to := strings.TrimSpace(command.Args()[0])
	argString := command.ArgString()
	text := strings.TrimSpace(strings.TrimPrefix(argString, to))
	actualTo := strings.TrimLeft(to, "@")
	recipient := t.bot.OnlineUsers().Find(actualTo)
	if recipient == nil {
		return []*domain.ClientMessage{
			domain.NewClientMessage(fmt.Sprintf("%s isn't online", actualTo), command.Sender(), command.Private()),
		}, nil
	}
	return []*domain.ClientMessage{
		domain.NewClientMessage(text, recipient, t.private),
	}, nil
}
