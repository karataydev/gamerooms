package game

import (
	"context"
	"log"

	"github.com/karataymarufemre/gamerooms/internal/message"
	"github.com/redis/go-redis/v9"
)

type RoomService interface {
	NewRoom() Room
}

type RedisRoomService struct {
	rdb *redis.Client
}

func NewRedisRoomService(rdb *redis.Client) *RedisRoomService {
	return &RedisRoomService{
		rdb: rdb,
	}
}

func (r *RedisRoomService) NewRoom() Room {
	room := NewRoom(r.SendToClient)
	go r.Subscribe(room)
	go room.Loop()
	return room
}

func (r *RedisRoomService) Subscribe(room *GameRoom) {
	r.rdb.Set(room.ctx, "exists" + room.Id(), room.Id(), 0)
	log.Printf("room started: %s", room.Id())
	for {
		strSlice, _ := r.rdb.BRPop(room.ctx, 0, room.Id()).Result()
		log.Printf("message recieved from queue for room %s: %s", strSlice[0], strSlice[1])
		m := message.FromStr(strSlice[1])
		room.msg <- m
	}
}

func (r *RedisRoomService) SendToClient(ctx context.Context, clientId string, msg *message.Message) {
	if err := r.rdb.LPush(ctx, clientId, msg.ToJson()).Err(); err != nil {
		log.Printf("message could not send to client message queue: %s, message: %v, error: %v", clientId, msg, err)
	}
}
