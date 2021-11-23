package pkg

import (
	"fmt"
	"github.com/raf924/connector-sdk/command"
	"github.com/raf924/connector-sdk/domain"
	"github.com/raf924/connector-sdk/storage"
	"log"
	"strings"
	"time"
)

var _ command.Command = (*MessageCommand)(nil)

type User struct {
	nick string
	id   string
}

type Message struct {
	Message   string `json:"message"`
	Sender    User   `json:"sender"`
	Timestamp int64  `json:"timestamp"`
	Private   bool   `json:"private"`
}

/*func (receiver Message) To() *domain.ChatMessage {
	return domain.NewChatMessage(receiver.Message, receiver.Sender, nil, false, receiver.Private, time.UnixMilli(receiver.Timestamp), true)
}*/

func (r User) UnmarshalText([]byte) error {
	return nil
}

func (r User) MarshalText() (text []byte, err error) {
	return []byte(fmt.Sprintf("%s#%s", r.nick, r.id)), nil
}

func (r *User) UnmarshalJSON(bytes []byte) error {
	text := strings.TrimLeft(string(bytes), "\"")
	text = strings.TrimRight(text, "\"")
	parts := strings.Split(text, "#")
	*r = User{
		nick: parts[0],
		id:   parts[1],
	}
	return nil
}

type MessageCommand struct {
	command.NoOpCommand
	userMessages   map[User][]Message
	bot            command.Executor
	messageStorage storage.Storage
}

func (m *MessageCommand) Init(bot command.Executor) error {
	m.bot = bot
	m.userMessages = map[User][]Message{}
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

func (m *MessageCommand) Execute(command *domain.CommandMessage) ([]*domain.ClientMessage, error) {
	defer m.save()
	ms, err := m.OnChat(domain.NewChatMessage("", command.Sender(), nil, false, command.Private(), command.Timestamp(), false))
	if err != nil {
		println("error:", err.Error())
	}
	if len(command.Args()) == 0 {
		return nil, fmt.Errorf("can't send message to no one")
	}
	message := strings.TrimSpace(strings.TrimPrefix(command.ArgString(), command.Args()[0]))
	if len(message) == 0 {
		return ms, nil
	}
	to := strings.TrimLeft(command.Args()[0], "@")
	user := strings.Split(to, "#")
	var id string
	nick := user[0]
	if len(user) > 1 {
		id = user[1]
	}
	recipient := User{}
	toUser := m.bot.OnlineUsers().Find(nick)
	if toUser == nil {
		toUser = domain.NewUser(nick, id, domain.RegularUser)
	}
	if len(toUser.Id()) == 0 {
		recipient.nick = toUser.Nick()
	} else {
		recipient.id = toUser.Id()
	}
	if _, ok := m.userMessages[recipient]; !ok {
		m.userMessages[recipient] = make([]Message, 0, 1)
	}
	m.userMessages[recipient] = append(m.userMessages[recipient], Message{
		Message: message,
		Sender: User{
			nick: command.Sender().Nick(),
			id:   command.Sender().Id(),
		},
		Timestamp: command.Timestamp().UnixMilli(),
		Private:   command.Private(),
	})
	ms = append(ms, domain.NewClientMessage(fmt.Sprintf("@%s will receive your message once they're back", to), command.Sender(), command.Private()))
	return ms, nil
}

func (m *MessageCommand) OnChat(message *domain.ChatMessage) ([]*domain.ClientMessage, error) {
	defer m.save()
	recipient := User{id: message.Sender().Id()}
	userMessages, ok := m.userMessages[recipient]
	if !ok {
		recipient = User{nick: message.Sender().Nick()}
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
		if userMessage.Private || message.Private() {
			count = &privateCount
			text = &privateText
		} else {
			count = &publicCount
			text = &publicText
		}
		*count += 1
		timeAgo := now.Sub(time.UnixMilli(userMessage.Timestamp))
		for _, duration := range []time.Duration{time.Second, time.Minute, time.Hour} {
			if timeAgo < 10*duration {
				break
			}
			timeAgo = timeAgo.Round(duration)
		}
		*text += fmt.Sprintf("[%s ago] %s: %s\n", now.Sub(time.UnixMilli(userMessage.Timestamp)).Round(time.Second).String(), userMessage.Sender.nick, userMessage.Message)
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
	allMessages := make([]*domain.ClientMessage, 0, 2)
	publicMessage := domain.NewClientMessage(publicText, message.Sender(), false)
	privateMessage := domain.NewClientMessage(privateText, message.Sender(), true)
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
