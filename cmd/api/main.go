package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Command struct {
	Action string                 `json:"action"`
	Data   map[string]interface{} `json:"data"`
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Command)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	http.HandleFunc("/ws", handleWebsocket)
	http.HandleFunc("GET /rooms", handleRooms)

	// go handleMessages()

	r := NewChatRoom("Public")
	r2 := NewChatRoom("Private Channel")
	Rooms[r.ID] = r
	Rooms[r2.ID] = r2

	Users["JohnCena"] = &User{
		ID:   "JohnCena",
		Name: "Johnathan Cena",
	}

	log.Println("Listening on port :3001")
	if err := http.ListenAndServe(":3001", nil); err != nil {
		log.Fatalf("ListenAndServe Error: %s", err)
	}

}

func handleWebsocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade Error: %s", err)
		return
	}

	defer func() {
		delete(clients, ws)
		ws.Close()
	}()

	clients[ws] = true

	for {
		var msg Command

		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("ReadJSON Error: %s", err)
			break
		}

		d, err := json.Marshal(msg.Data)
		if err != nil {
			// Handle error
			log.Println("Error marshalling data")
			return
		}

		switch msg.Action {
		case "join-room":
			var req JoinRoomRequest
			err = json.Unmarshal(d, &req)
			if err != nil {
				// Handle error
				log.Println("Error unmarshalling to JoinRoomRequest")
				return
			}

			r := Rooms[req.RoomID]
			r.Join(ws, *Users[req.JoinerID])
		case "message-room":

			var req Message
			err = json.Unmarshal(d, &req)
			if err != nil {
				// Handle error
				log.Println("Error unmarshalling to Message")
				return
			}

			r := Rooms[req.RoomID]
			msg := Message{
				MemberID: req.MemberID,
				RoomID:   req.RoomID,
				Message:  req.Message,
			}
			r.BroadcastMessage(msg)
		}

	}
}

// func handleMessages() {
// 	for {

// 	}
// }

func handleRooms(w http.ResponseWriter, r *http.Request) {
	var rooms = make([]*ChatRoom, 0)

	for _, cr := range Rooms {
		rooms = append(rooms, cr)
	}

	b, err := json.Marshal(rooms)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		w.Write([]byte("{\"message\":\"Error getting rooms\"}"))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)

}
