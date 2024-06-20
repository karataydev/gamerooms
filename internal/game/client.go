package game

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
	"github.com/karataymarufemre/gamerooms/internal/message"
)

type Client struct {
	// client id
	id string

	// room
	roomId string

	// The websocket connection.
	conn *websocket.Conn
}

func(c *Client) OnMessageRecieve(m *message.Message) ([]byte, bool) {
	if m.Role == message.ADMIN {
		log.Printf("from admin")
	}
	log.Printf("Subscribe %s için kanaldan okunan => %v", c.id, m)
	return []byte("recieved"), true
}

func (c *Client) FromClient(msg []byte) *message.Message {
	newMsg := &message.Message{From: c.id, Role: message.PLAYER}
	json.Unmarshal(msg, newMsg)
	return newMsg
}
