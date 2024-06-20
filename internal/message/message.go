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
