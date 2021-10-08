package pkg

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestMessagesToJson(t *testing.T) {
	fullRecipient := User{
		nick: "nick",
		id:   "id",
	}
	nickRecipient := User{
		nick: "nick",
	}
	idRecipient := User{
		id: "id",
	}
	messages := map[User][]Message{
		fullRecipient: {
			{
				Timestamp: time.Now().UnixMilli(),
				Message:   "fullRecipient",
				Private:   false,
			},
		},
		nickRecipient: {
			{
				Timestamp: time.Now().UnixMilli(),
				Message:   "nickRecipient",
				Private:   false,
			},
		},
		idRecipient: {
			{
				Timestamp: time.Now().UnixMilli(),
				Message:   "idRecipient",
				Private:   false,
			},
		},
	}
	buf, err := json.Marshal(messages)
	if err != nil {
		t.Errorf(err.Error())
	}
	t.Log(string(buf))
	var decodedMessages map[User][]Message
	err = json.Unmarshal(buf, &decodedMessages)
	if err != nil {
		t.Errorf(err.Error())
	}
	if !reflect.DeepEqual(messages, decodedMessages) {
		t.Errorf("expected\n%v\n, got\n%v", messages, decodedMessages)
	}
}
