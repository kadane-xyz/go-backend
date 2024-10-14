package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/coder/websocket"
	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

type Message struct {
	Type      string      `json:"type"`
	RoomID    pgtype.Int8  `json:"room_id"`
	Content   string      `json:"content"`
	AccountID pgtype.Int8  `json:"account_id"`
	Username  string      `json:"username"`
	ProblemID pgtype.Int8  `json:"problem_id"`
}

func HandleWebSocket(manager *Manager, queries *sql.Queries) http.HandlerFunc {
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
					// Insert new room into database
					roomID, err := queries.CreateRoom(r.Context(), sql.CreateRoomParams{
						Name:            msg.Content,
						ProblemID:       msg.ProblemID,
						MaxParticipants: 4,
						TimeLimit:       3600,
						CreatorID:       msg.AccountID,
					})
					if err != nil {
						log.Printf("Error creating room: %v", err)
						sendResponse(conn, "Error creating room")
						continue
					}
					
					room := manager.CreateRoom(roomID, msg.Content, msg.ProblemID, 4, 3600, msg.AccountID)
					err = queries.AddParticipantToRoom(r.Context(), sql.AddParticipantToRoomParams{
						RoomID:    pgtype.Int8{Int64: roomID, Valid: true},
						AccountID: msg.AccountID,
					})
					if err != nil {
						log.Printf("Error adding participant to room: %v", err)
					}
					
					room.AddClient(msg.AccountID, conn)
					sendResponse(conn, fmt.Sprintf("Room created and joined: %d", roomID))

				case "join":
					room, exists := manager.GetRoom(msg.RoomID)
					if !exists {
						sendResponse(conn, fmt.Sprintf("Room not found: %d", msg.RoomID.Int64))
						continue
					}
					room.AddClient(msg.AccountID, conn)
					
					// Insert participant into room_participant table
					err := queries.AddParticipantToRoom(r.Context(), sql.AddParticipantToRoomParams{
						RoomID:    msg.RoomID,
						AccountID: msg.AccountID,
					})
					if err != nil {
						log.Printf("Error adding participant to room: %v", err)
					}
					
					sendResponse(conn, fmt.Sprintf("Joined room: %d", msg.RoomID.Int64))

				case "leave":
					room, exists := manager.GetRoom(msg.RoomID)
					if !exists {
						sendResponse(conn, fmt.Sprintf("Room not found: %d", msg.RoomID.Int64))
						continue
					}
					room.RemoveClient(msg.AccountID)
					
					// Update participant status in room_participant table
					err := queries.UpdateParticipantStatus(r.Context(), sql.UpdateParticipantStatusParams{
						RoomID:    msg.RoomID,
						AccountID: msg.AccountID,
						Status:    "inactive",
					})
					if err != nil {
						log.Printf("Error updating participant status: %v", err)
					}
					
					sendResponse(conn, fmt.Sprintf("Left room: %d", msg.RoomID.Int64))

				case "message":
					room, exists := manager.GetRoom(msg.RoomID)
					if !exists {
						sendResponse(conn, fmt.Sprintf("Room not found: %d", msg.RoomID.Int64))
						continue
					}
					
					// Insert message into room_message table
					err := queries.SaveRoomMessage(r.Context(), sql.SaveRoomMessageParams{
						RoomID:    msg.RoomID,
						AccountID: msg.AccountID,
						Content:   msg.Content,
					})
					if err != nil {
						log.Printf("Error saving message: %v", err)
					}
					
					room.Broadcast([]byte(msg.Content))

				case "list_clients":
					room, exists := manager.GetRoom(msg.RoomID)
					if !exists {
						sendResponse(conn, fmt.Sprintf("Room not found: %d", msg.RoomID.Int64))
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
