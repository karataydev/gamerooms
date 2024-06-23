package game

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/karataymarufemre/gamerooms/internal/event"
	"github.com/karataymarufemre/gamerooms/internal/message"
	"github.com/redis/go-redis/v9"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type ClientService interface {
	Connect(*websocket.Conn, context.Context, string, string)
}

type RedisClientService struct {
	rdb *redis.Client
}

func NewRedisClientService(rdb *redis.Client) *RedisClientService {
	return &RedisClientService{
		rdb: rdb,
	}
}

func (c *RedisClientService) Connect(conn *websocket.Conn, ctx context.Context, roomId, userId string) {
	client := &Client{
		id: userId,
		roomId: roomId,
		conn: conn,
		msg: make(chan *message.Message),
	}

	err := c.rdb.Get(ctx, "exists" + roomId).Err()
	if err != nil {
		log.Printf("room does not exist, roomId: %s", roomId)
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "room does not exist"))
	}
	connectMsg := &message.Message{From: userId, Role: message.PLAYER, Event: event.Connect}
	err = c.rdb.LPush(ctx, roomId, connectMsg.ToJson()).Err()
	conn.WriteMessage(websocket.TextMessage, []byte("Connected"))
	go c.Subscribe(client, conn, ctx)
	go c.Publish(client, conn, ctx)
}

// listen redis pub sub and call client On Message Recieve
func (c *RedisClientService) Subscribe(client *Client, conn *websocket.Conn, ctx context.Context) {
	go c.SubscribeQueue(client, conn, ctx)
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()
	for {
		select {
		case msg, ok := <- client.msg:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			conn.WriteMessage(websocket.TextMessage, msg.ToJson())
		case <- ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println("ping err")
				return
			}
		}
	}
}

func (c *RedisClientService) SubscribeQueue(client *Client, conn *websocket.Conn, ctx context.Context) {
	for {
		strSlice, _ := c.rdb.BRPop(ctx, 0, client.id).Result()
		log.Printf("message recieved from queue for client %s: %s", strSlice[0], strSlice[1])
		m := message.FromStr(strSlice[1])
		client.msg <- m
	}
}

// listen websocket connection and enqueue to redis queue
func (c *RedisClientService) Publish(client *Client, conn *websocket.Conn, ctx context.Context) {
	defer func() {
		conn.Close()
	}()
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		newMsg := client.FromClient(message)
		content, err := json.Marshal(newMsg)
		log.Println(newMsg)
		if err = c.rdb.LPush(ctx, client.roomId, content).Err(); err != nil {
			log.Printf("error: %v", err)
		}
	}
}

