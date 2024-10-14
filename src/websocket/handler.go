package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/coder/websocket"
)

type Message struct {
	Type      string `json:"type"`
	RoomID    string `json:"room_id"`
	Content   string `json:"content"`
	AccountID string `json:"account_id"`
}

func HandleWebSocket(manager *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			log.Printf("WebSocket accept error: %v", err)
			return
		}
		defer conn.Close(websocket.StatusInternalError, "")

		for {
			messageType, message, err := conn.Read(r.Context())
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				return
			}

			if messageType == websocket.MessageText {
				var msg Message
				if err := json.Unmarshal(message, &msg); err != nil {
					log.Printf("JSON unmarshal error: %v", err)
					continue
				}

				switch msg.Type {
				case "create":
					room := manager.CreateRoom(msg.RoomID)
					room.AddClient(msg.AccountID, conn)
					sendResponse(conn, "Room created and joined: "+msg.RoomID)
				case "join":
					room, exists := manager.GetRoom(msg.RoomID)
					if !exists {
						sendResponse(conn, "Room not found: "+msg.RoomID)
						continue
					}
					room.AddClient(msg.AccountID, conn)
					sendResponse(conn, "Joined room: "+msg.RoomID)
				case "leave":
					room, exists := manager.GetRoom(msg.RoomID)
					if !exists {
						sendResponse(conn, "Room not found: "+msg.RoomID)
						continue
					}
					room.RemoveClient(msg.AccountID)
					sendResponse(conn, "Left room: "+msg.RoomID)
				case "message":
					room, exists := manager.GetRoom(msg.RoomID)
					if !exists {
						sendResponse(conn, "Room not found: "+msg.RoomID)
						continue
					}
					room.Broadcast([]byte(msg.Content))
				case "list_clients":
					room, exists := manager.GetRoom(msg.RoomID)
					if !exists {
						sendResponse(conn, "Room not found: "+msg.RoomID)
						continue
					}
					clients := room.GetClients()
					clientsJSON, _ := json.Marshal(clients)
					sendResponse(conn, string(clientsJSON))
				default:
					sendResponse(conn, "Unknown message type: "+msg.Type)
				}
			}
		}
	}
}

func sendResponse(conn *websocket.Conn, message string) {
	response := Message{
		Type:    "response",
		Content: message,
	}
	responseJSON, _ := json.Marshal(response)
	conn.Write(context.Background(), websocket.MessageText, responseJSON)
}
