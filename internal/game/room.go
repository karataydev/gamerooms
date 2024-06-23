package game

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

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
	ctx context.Context

	gameLoopTicker time.Duration
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
	vote *Vote
	room *GameRoom
}



type Vote struct {
	To string
	Skip bool
}

func NewRoom(sendFunc func(context.Context, string, *message.Message)) *GameRoom {
	return &GameRoom{
		id: uuid.NewString(),
		players: make(map[string]*Player),
		sendFunc: sendFunc,
		msg: make(chan *message.Message),
		ctx: context.Background(),
	}
}

func (r *GameRoom) Id() string {
	return r.id
}

func (r *GameRoom) AddPlayer(clientId string) {
	r.players[clientId] = &Player{id: clientId, state: NotReady, room: r}
}

func (r *GameRoom) sendToClient(clientId string, msg *message.Message) {
	r.sendFunc(r.ctx, clientId, msg)
}

func (r *GameRoom) Loop() {
	r.waitAllPlayersReady()
	r.sendToAllClients(&message.Message{Event: "GameStarted"})
	fmt.Println("All players ready")
	ticker := time.NewTicker(r.gameLoopTicker)
	for {
		select {
		case <- ticker.C:
			ticker.Reset(r.gameLoopTicker)
		}
	}
}

func (r *GameRoom) sendToAllClients(m *message.Message) {
	for _, p := range r.players {
		r.sendToClient(p.id, m)
	}
}

func (r *GameRoom) OnVote(msg *message.Message) {
	var v Vote
	err := json.Unmarshal(*msg.Data, &v)
	if err != nil {
		log.Printf("could not parse vote message. err : %v", err)
	}
	p := r.players[msg.From]
	_, ok := r.players[v.To]
	if v.Skip || ok  {
		p.vote = &v
	} else {
		log.Printf("player not found. player id: %v", v.To)
	}
	
}


func (r *GameRoom) waitAllPlayersReady() {
	for !r.isAllPlayersReady() {
		select {
		case m, _ := <- r.msg:
			fmt.Println(m)
			switch m.Event {
			case event.Connect:
				r.AddPlayer(m.From)
			case event.Ready:
				r.players[m.From].state = Ready
			default:
				r.sendInvalidEvent(m, event.Connect, event.Ready)
			}
		}
	}
}

type InvalidEvent struct {
	AvailableEvents []event.Event
}

func (r *GameRoom) sendInvalidEvent(m *message.Message, availableEvents ...event.Event) {
	invM := &message.Message{Event: event.InvalidEvent}
	events, _ := json.Marshal(&InvalidEvent{AvailableEvents: availableEvents})
	invM.Data = (*json.RawMessage)(&events)
	r.sendToClient(m.From, invM)
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

func (r *GameRoom) DayLoop() {
	
}

func (r *GameRoom) NightLoop() {
}


func (r *GameRoom) VotingPhase() {
	for {
		select {
		case m, _ := <- r.msg:
			switch m.Event {
			case event.Vote:
				r.OnVote(m)
			default:
				r.sendInvalidEvent(m, event.Vote)
			}

		}
	}
}
