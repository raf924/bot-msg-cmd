package pkg

import (
	"encoding/json"
	messages2 "github.com/raf924/connector-api/pkg/gen"
	"google.golang.org/protobuf/types/known/timestamppb"
	"reflect"
	"testing"
)

func TestMessagesToJson(t *testing.T) {
	fullRecipient := Recipient{
		nick: "nick",
		id:   "id",
	}
	nickRecipient := Recipient{
		nick: "nick",
	}
	idRecipient := Recipient{
		id: "id",
	}
	messages := map[Recipient][]*messages2.MessagePacket{
		fullRecipient: {
			{
				Timestamp: timestamppb.Now(),
				Message:   "fullRecipient",
				Private:   false,
			},
		},
		nickRecipient: {
			{
				Timestamp: timestamppb.Now(),
				Message:   "nickRecipient",
				Private:   false,
			},
		},
		idRecipient: {
			{
				Timestamp: timestamppb.Now(),
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
	var decodedMessages map[Recipient][]*messages2.MessagePacket
	err = json.Unmarshal(buf, &decodedMessages)
	if err != nil {
		t.Errorf(err.Error())
	}
	if !reflect.DeepEqual(messages, decodedMessages) {
		t.Errorf("expected\n%v\n, got\n%v", messages, decodedMessages)
	}
}
