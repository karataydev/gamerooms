package message

import (
	"encoding/json"
	"log"

	"github.com/karataymarufemre/gamerooms/internal/event"
	"github.com/redis/go-redis/v9"
)



type RoomRole int

const (
	PLAYER RoomRole = iota + 1
	ADMIN
)



type Message struct {
	From string
	Role RoomRole
	Content string
	Event event.Event
	Data *json.RawMessage
}


func FromRedis(msg *redis.Message) *Message {
	m := &Message{}	
	err := json.Unmarshal([]byte(msg.Payload), m)
	if err != nil {
		log.Println(err)
		return nil
	}
	return m
}

func FromStr(payload string) *Message {
	m := &Message{}	
	err := json.Unmarshal([]byte(payload), m)
	if err != nil {
		log.Println(err)
		return nil
	}
	return m

}

func (m *Message) ToJson() []byte {
	json, err := json.Marshal(m)
	if err != nil {
		log.Printf("could not convert message to json. message: %v", m)
	}
	return json
}
