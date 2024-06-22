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
	ctx := context.Background()
	room := NewRoom(r.SendToClient)
	go r.Subscribe(ctx, room)
	go room.Loop(ctx)
	return room
}

func (r *RedisRoomService) Subscribe(ctx context.Context, room *GameRoom) {
	for {
		strSlice, _ := r.rdb.BRPop(ctx, 0, room.Id()).Result()
		log.Printf("message recieved from queue for room %s: %s", strSlice[0], strSlice[1])
		m := message.FromStr(strSlice[1])
		room.msg <- m
	}
}

func (r *RedisRoomService) SendToClient(ctx context.Context, clientId string, msg *message.Message) {
	if err := r.rdb.LPush(ctx, clientId, msg).Err(); err != nil {
		log.Printf("message could not send to client message queue: %s, message: %v", clientId, msg)
	}
}
