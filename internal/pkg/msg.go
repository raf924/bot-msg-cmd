package pkg

import (
	"fmt"
	"github.com/raf924/bot/pkg/bot/command"
	"github.com/raf924/bot/pkg/storage"
	messages "github.com/raf924/connector-api/pkg/gen"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"strings"
	"time"
)

type Recipient struct {
	nick string
	id   string
}

func (r Recipient) UnmarshalText([]byte) error {
	return nil
}

func (r Recipient) MarshalText() (text []byte, err error) {
	return []byte(fmt.Sprintf("%s#%s", r.nick, r.id)), nil
}

func (r *Recipient) UnmarshalJSON(bytes []byte) error {
	text := strings.TrimLeft(string(bytes), "\"")
	text = strings.TrimRight(text, "\"")
	parts := strings.Split(text, "#")
	*r = Recipient{
		nick: parts[0],
		id:   parts[1],
	}
	return nil
}

type MessageCommand struct {
	command.NoOpCommand
	userMessages   map[Recipient][]*messages.MessagePacket
	bot            command.Executor
	messageStorage storage.Storage
}

func (m *MessageCommand) Init(bot command.Executor) error {
	m.bot = bot
	m.userMessages = map[Recipient][]*messages.MessagePacket{}
	messageStorageLocation := bot.ApiKeys()["messageStorageLocation"]
	messageStorage, err := storage.NewFileStorage(messageStorageLocation)
	if err != nil {
		messageStorage = storage.NewNoOpStorage()
	}
	m.messageStorage = messageStorage
	m.load()
	return nil
}

func (m *MessageCommand) Name() string {
	return "message"
}

func (m *MessageCommand) Aliases() []string {
	return []string{"msg", "m"}
}

func (m *MessageCommand) load() {
	err := m.messageStorage.Load(&m.userMessages)
	if err != nil {
		log.Println("error reading message file:", err.Error())
		return
	}
}

func (m *MessageCommand) save() {
	m.messageStorage.Save(m.userMessages)
}

func (m *MessageCommand) Execute(command *messages.CommandPacket) ([]*messages.BotPacket, error) {
	defer m.save()
	ms, err := m.OnChat(&messages.MessagePacket{
		Timestamp: command.GetTimestamp(),
		Message:   "",
		User:      command.GetUser(),
		Private:   command.GetPrivate(),
	})
	if err != nil {
		println("error:", err.Error())
	}
	if len(command.GetArgs()) == 0 {
		return nil, fmt.Errorf("can't send message to no one")
	}
	message := strings.TrimSpace(strings.TrimPrefix(command.GetArgString(), command.GetArgs()[0]))
	if len(message) == 0 {
		return ms, nil
	}
	to := strings.TrimLeft(command.GetArgs()[0], "@")
	user := strings.Split(to, "#")
	var id string
	nick := user[0]
	if len(user) > 1 {
		id = user[1]
	}
	recipient := Recipient{}
	toUser, ok := m.bot.OnlineUsers()[nick]
	if !ok {
		toUser = messages.User{
			Nick:  nick,
			Id:    id,
			Mod:   false,
			Admin: false,
		}
	}
	if len(toUser.GetId()) == 0 {
		recipient.nick = toUser.Nick
	} else {
		recipient.id = toUser.GetId()
	}
	if _, ok = m.userMessages[recipient]; !ok {
		m.userMessages[recipient] = make([]*messages.MessagePacket, 0)
	}
	m.userMessages[recipient] = append(m.userMessages[recipient], &messages.MessagePacket{
		Timestamp: command.GetTimestamp(),
		Message:   message,
		User:      command.GetUser(),
		Private:   command.GetPrivate(),
	})
	ms = append(ms, &messages.BotPacket{
		Timestamp: timestamppb.Now(),
		Message:   fmt.Sprintf("@%s will receive your message once they're back", to),
		Recipient: command.GetUser(),
		Private:   command.GetPrivate(),
	})
	return ms, nil
}

func (m *MessageCommand) OnChat(message *messages.MessagePacket) ([]*messages.BotPacket, error) {
	defer m.save()
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
		if userMessage.GetPrivate() || message.GetPrivate() {
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
