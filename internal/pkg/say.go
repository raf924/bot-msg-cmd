package pkg

import (
	"github.com/raf924/connector-sdk/command"
	"github.com/raf924/connector-sdk/domain"
)

type SayCommand struct {
	command.NoOpInterceptor
	executor command.Executor
}

func (s *SayCommand) Init(bot command.Executor) error {
	s.executor = bot
	return nil
}

func (s *SayCommand) Name() string {
	return "say"
}

func (s *SayCommand) Aliases() []string {
	return nil
}

func (s *SayCommand) Execute(command *domain.CommandMessage) ([]*domain.ClientMessage, error) {
	argString := command.ArgString()
	return []*domain.ClientMessage{
		domain.NewClientMessage(argString, nil, false),
	}, nil
}

var _ command.Command = (*SayCommand)(nil)
