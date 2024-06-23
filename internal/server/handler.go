package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/karataymarufemre/gamerooms/internal/game"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type ApiHandler struct {
	clientService game.ClientService
	roomService game.RoomService
}

func NewApiHandler(clientSrvice game.ClientService, roomService game.RoomService) *ApiHandler {
	return &ApiHandler{
		clientService: clientSrvice,
		roomService: roomService,
	}
}

func (h *ApiHandler) ServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws/connect", h.serveWebsocket)
	mux.HandleFunc("/create", h.newRoom)
	return mux
}


func (h *ApiHandler) serveWebsocket(w http.ResponseWriter, r *http.Request)  {
	ctx := context.Background()
	roomId := r.URL.Query().Get("roomId")
	userId := r.URL.Query().Get("userId")
	log.Printf("user %s connecting to room %s", userId, roomId)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	h.clientService.Connect(conn, ctx, roomId, userId)
}

type NewRoomResponse struct {
	RoomId string `json:"roomId"`
}

func (h *ApiHandler) newRoom(w http.ResponseWriter, r *http.Request) {
	roomId := h.roomService.NewRoom().Id()
	resp := NewRoomResponse{RoomId: roomId}
	ToJson(w, http.StatusOK, resp)
}


func ToJson(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
