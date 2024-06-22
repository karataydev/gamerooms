package game

import (
	"context"
	"fmt"

	//"time"

	"github.com/google/uuid"
	"github.com/karataymarufemre/gamerooms/internal/event"
	"github.com/karataymarufemre/gamerooms/internal/message"
)


type Room interface {
	Id() string
}

type GameRoom struct {
	id string
	players map[string]*Player
	sendFunc func(context.Context, string, *message.Message)
	msg chan *message.Message
}

type PlayerState int
type PlayerRole int

const (
	NotReady PlayerState = iota
	Ready
)

const (
	Admin PlayerRole = iota
	Normal
)

type Player struct {
	id string
	state PlayerState
	role  PlayerRole
}

func NewRoom(sendFunc func(context.Context, string, *message.Message)) *GameRoom {
	return &GameRoom{
		id: uuid.NewString(),
		players: make(map[string]*Player),
		sendFunc: sendFunc,
		msg: make(chan *message.Message),
	}
}

func (r *GameRoom) Id() string {
	return r.id
}

func (r *GameRoom) AddPlayer(clientId string) {
	r.players[clientId] = &Player{id: clientId, state: NotReady}
}

func (r *GameRoom) sendToClient(ctx context.Context, clientId string, msg *message.Message) {
	r.sendFunc(ctx, clientId, msg)
}

func (r *GameRoom) Loop(ctx context.Context) {
	// ticker := time.NewTicker(pingPeriod)
	// wait everyone to be ready
	for !r.isAllPlayersReady() {
		select {
		case m, _ := <- r.msg:
			fmt.Println(m)
			switch m.Event {
			case event.Connect:
				r.AddPlayer(m.From)
			case event.Ready:
				r.players[m.From].state = Ready
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
		if p.state != Ready {
			isAllReady = false
			break
		}
	}
	return isAllReady
}

