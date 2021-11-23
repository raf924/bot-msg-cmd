package bot_msg_cmd

import (
	"github.com/raf924/bot-msg-cmd/v2/internal/pkg"
	"github.com/raf924/connector-sdk/command"
)

func init() {
	command.HandleCommand(&pkg.MessageCommand{})
	command.HandleCommand(&pkg.SayCommand{})
	command.HandleCommand(pkg.NewTellCommand(true))
	command.HandleCommand(pkg.NewTellCommand(false))
}
