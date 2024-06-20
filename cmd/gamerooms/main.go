package main

import (
	"context"
	"log"
	"net/http"

	"github.com/karataymarufemre/gamerooms/internal/game"
	"github.com/karataymarufemre/gamerooms/internal/server"
	"github.com/redis/go-redis/v9"
)


func main() {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	err := rdb.Ping(ctx).Err()
	if err != nil {
		panic("redis connection error => " + err.Error())
	}

	clientService := game.NewRedisClientService(rdb)
	roomService := game.NewRedisRoomService(rdb)
	apiHandler := server.NewApiHandler(clientService, roomService)

	err = http.ListenAndServe(":8080", apiHandler.ServeMux())
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	
}
