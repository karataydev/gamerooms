package game

import "github.com/redis/go-redis/v9"

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
	room := NewRoom(r.rdb)
	go room.Run()
	return room
}
