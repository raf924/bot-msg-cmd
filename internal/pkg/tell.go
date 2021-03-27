package pkg

import (
	"fmt"
	"github.com/raf924/bot/api/messages"
	"github.com/raf924/bot/pkg/bot/command"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
)

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

func (t *TellCommand) Execute(command *messages.CommandPacket) ([]*messages.BotPacket, error) {
	to := strings.TrimSpace(command.GetArgs()[0])
	argString := command.GetArgString()
	text := strings.TrimSpace(strings.TrimPrefix(argString, to))
	actualTo := strings.TrimLeft(to, "@")
	recipient, exists := t.bot.OnlineUsers()[actualTo]
	if !exists {
		return []*messages.BotPacket{
			{
				Timestamp: timestamppb.Now(),
				Message:   fmt.Sprintf("%s isn't online", actualTo),
				Recipient: command.GetUser(),
				Private:   command.GetPrivate(),
			},
		}, nil
	}
	return []*messages.BotPacket{
		{
			Timestamp: timestamppb.Now(),
			Message:   text,
			Recipient: &recipient,
			Private:   t.private,
		},
	}, nil
}
