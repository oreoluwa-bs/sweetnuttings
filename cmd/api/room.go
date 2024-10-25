package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gorilla/websocket"
)

var Users = make(map[string]*User)
var Rooms = make(map[string]*ChatRoom)

type ChatRoom struct {
	ID               string                   `json:"id"`
	Name             string                   `json:"name"`
	Members          map[string]User          `json:"members"`
	Clients          map[*websocket.Conn]bool `json:"-"`
	broadcastChannel chan []Message           `json:"-"`
}

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsOnline bool   `json:"is_online"`
	messages []string
}

type Message struct {
	MemberID string `json:"member_id"`
	RoomID   string `json:"room_id"`
	Message  string `json:"message"`
}

type JoinRoomRequest struct {
	RoomID   string `json:"room_id"`
	JoinerID string `json:"joiner_id"`
}

func NewChatRoom(name string) *ChatRoom {
	return &ChatRoom{
		ID:      strings.ReplaceAll(strings.ToLower(name), " ", "_"),
		Name:    name,
		Members: make(map[string]User),
		Clients: make(map[*websocket.Conn]bool),
	}
}

func (cr *ChatRoom) Join(client *websocket.Conn, u User) {
	cr.Members[u.ID] = u
	cr.Clients[client] = true

	msg := Message{
		MemberID: "server",
		RoomID:   cr.ID,
		Message:  fmt.Sprintf("%s has joined the chat", u.Name),
	}

	cr.BroadcastMessage(msg)
}

func (cr *ChatRoom) Leave(u User) {
	delete(cr.Members, u.ID)

	msg := Message{
		MemberID: "server",
		RoomID:   cr.ID,
		Message:  fmt.Sprintf("%s has left the chat", u.Name),
	}
	cr.BroadcastMessage(msg)
}

func (cr *ChatRoom) BroadcastMessage(msg Message) {

	b, _ := json.Marshal(msg)

	var result map[string]interface{}
	json.Unmarshal(b, &result)

	s := &Command{
		Action: "room-message",
		Data:   result,
	}

	for client := range cr.Clients {
		err := client.WriteJSON(s)
		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}
