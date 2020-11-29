package pkg

import (
	"fmt"
	"github.com/raf924/bot/api/messages"
	"github.com/raf924/bot/pkg/bot"
	"github.com/raf924/bot/pkg/bot/command"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"time"
)

func init() {
	bot.HandleCommand(&MessageCommand{})
}

type Recipient struct {
	nick string
	id   string
}

type MessageCommand struct {
	command.NoOpCommand
	userMessages    map[Recipient][]*messages.MessagePacket
	privateMessages map[Recipient][][]*messages.MessagePacket
	bot             command.Executor
}

func (m *MessageCommand) Init(bot command.Executor) error {
	m.bot = bot
	m.userMessages = map[Recipient][]*messages.MessagePacket{}
	return nil
}

func (m *MessageCommand) Name() string {
	return "message"
}

func (m *MessageCommand) Aliases() []string {
	return []string{"msg", "m"}
}

func (m *MessageCommand) Execute(command *messages.CommandPacket) ([]*messages.BotPacket, error) {
	to := strings.Join(strings.Split(command.GetArgs()[0], "@"), "")
	recipient := Recipient{}
	toUser, ok := m.bot.OnlineUsers()[to]
	if !ok {
		toUser = messages.User{
			Nick:  to,
			Id:    "",
			Mod:   false,
			Admin: false,
		}
	}
	if len(toUser.GetId()) == 0 {
		recipient.nick = toUser.Nick
	}
	recipient.id = toUser.GetId()
	if _, ok = m.userMessages[recipient]; !ok {
		m.userMessages[recipient] = make([]*messages.MessagePacket, 0)
	}
	m.userMessages[recipient] = append(m.userMessages[recipient], &messages.MessagePacket{
		Timestamp: command.GetTimestamp(),
		Message:   strings.Join(command.GetArgs()[1:], " "),
		User:      command.GetUser(),
		Private:   command.GetPrivate(),
	})
	ms, err := m.OnChat(&messages.MessagePacket{
		Timestamp: command.GetTimestamp(),
		Message:   "",
		User:      command.GetUser(),
		Private:   command.GetPrivate(),
	})
	ms = append(ms, &messages.BotPacket{
		Timestamp: timestamppb.Now(),
		Message:   fmt.Sprintf("@%s will receive your message once they're back", to),
		Recipient: command.GetUser(),
		Private:   command.GetPrivate(),
	})
	return ms, err
}

func (m *MessageCommand) OnChat(message *messages.MessagePacket) ([]*messages.BotPacket, error) {
	recipient := Recipient{id: message.GetUser().GetId()}
	userMessages, ok := m.userMessages[recipient]
	if !ok {
		recipient = Recipient{nick: message.GetUser().GetNick()}
		userMessages, ok = m.userMessages[recipient]
		if !ok {
			return nil, nil
		}
	}
	publicText := ""
	privateText := ""
	publicCount := 0
	privateCount := 0
	now := time.Now()
	for _, userMessage := range userMessages {
		var count *int
		var text *string
		if userMessage.GetPrivate() {
			count = &privateCount
			text = &privateText
		} else {
			count = &publicCount
			text = &publicText
		}
		*count += 1
		timeAgo := now.Sub(userMessage.GetTimestamp().AsTime()).Round(time.Millisecond)
		for _, duration := range []time.Duration{time.Second, time.Minute, time.Hour} {
			if timeAgo < 10*duration {
				break
			}
			timeAgo = timeAgo.Round(duration)
		}
		*text += fmt.Sprintf("[%s ago] %s: %s\n", now.Sub(userMessage.GetTimestamp().AsTime()).Round(time.Second).String(), userMessage.GetUser().GetNick(), userMessage.GetMessage())
	}
	publicPlural := ""
	if publicCount > 1 {
		publicPlural = "s"
	}
	publicText = fmt.Sprintf("you have %d new message%s\n%s", publicCount, publicPlural, publicText)
	privatePlural := ""
	if privateCount > 1 {
		privatePlural = "s"
	}
	privateText = fmt.Sprintf("you have %d new message%s\n%s", privateCount, privatePlural, privateText)
	var allMessages []*messages.BotPacket
	publicMessage := &messages.BotPacket{
		Timestamp: timestamppb.Now(),
		Message:   publicText,
		Recipient: message.User,
		Private:   false,
	}
	privateMessage := &messages.BotPacket{
		Timestamp: timestamppb.Now(),
		Message:   privateText,
		Recipient: message.User,
		Private:   true,
	}
	if publicCount > 0 {
		allMessages = append(allMessages, publicMessage)
	}
	if privateCount > 0 {
		allMessages = append(allMessages, privateMessage)
	}
	delete(m.userMessages, recipient)
	return allMessages, nil
}

func (m *MessageCommand) IgnoreSelf() bool {
	return true
}
