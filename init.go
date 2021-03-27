package bot_msg_cmd

import (
	"github.com/raf924/bot-msg-cmd/internal/pkg"
	"github.com/raf924/bot/pkg/bot"
)

func init() {
	bot.HandleCommand(&pkg.MessageCommand{})
	bot.HandleCommand(pkg.NewTellCommand(true))
	bot.HandleCommand(pkg.NewTellCommand(false))
}
