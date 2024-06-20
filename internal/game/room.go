package game

import (
	"context"
	"fmt"
	//"time"

	"github.com/google/uuid"
	"github.com/karataymarufemre/gamerooms/internal/event"
	"github.com/karataymarufemre/gamerooms/internal/message"
	"github.com/redis/go-redis/v9"
)


type Room interface {
	Id() string
}

type GameRoom struct {
	id string
	players map[string]*Player
	rdb *redis.Client
}

type PlayerState int

const (
	NOT_READY PlayerState = iota
	READY
)

type Player struct {
	id string
	state PlayerState
}

func NewRoom(rdb *redis.Client) *GameRoom {
	return &GameRoom{
		id: uuid.NewString(),
		players: make(map[string]*Player),
		rdb: rdb,
	}
}

func (r *GameRoom) Id() string {
	return r.id
}

func (r *GameRoom) AddPlayer(clientId string) {
	r.players[clientId] = &Player{id: clientId, state: NOT_READY}
}

func (r *GameRoom) Run() {
	ctx := context.Background()
	subs := r.rdb.Subscribe(ctx, r.id)
	// ticker := time.NewTicker(pingPeriod)
	// wait everyone to be ready
	for !r.isAllPlayersReady() {
		select {
		case msg, _ := <- subs.Channel():
			m := message.FromRedis(msg)
			fmt.Println(m)
			switch m.Event {
			case event.Connect:
				r.AddPlayer(m.From)
			case event.Ready:
				r.players[m.From].state = READY
			}
		}
	}
	fmt.Println("All players ready")
	// TODO continue here   
	for {

	}
}

func (r *GameRoom) isAllPlayersReady() bool {
	isAllReady := true
	if len(r.players) == 0 {
		return false
	}
	for _, p := range r.players {
		if p.state != READY {
			isAllReady = false
			break
		}
	}
	return isAllReady
}

