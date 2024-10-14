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
	RoomID    int64       `json:"room_id"`
	Content   string      `json:"content"`
	AccountID int64       `json:"account_id"`
	Username  string      `json:"username"`
	ProblemID int64       `json:"problem_id"`
}

type Response struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

func int64ToPgInt8(i int64) pgtype.Int8 {
	return pgtype.Int8{Int64: i, Valid: true}
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
					log.Printf("Received message: %s", string(message))
					continue
				}
				log.Printf("Parsed message: %+v", msg)

				switch msg.Type {
				case "create":
					// Insert new room into database
					roomID, err := queries.CreateRoom(r.Context(), sql.CreateRoomParams{
						Name:            msg.Content,
						ProblemID:       int64ToPgInt8(msg.ProblemID),
						MaxParticipants: 4,
						TimeLimit:       3600,
						CreatorID:       int64ToPgInt8(msg.AccountID),
					})
					if err != nil {
						log.Printf("Error creating room: %v", err)
						if err := sendResponse(conn, "error", "Error creating room"); err != nil {
							log.Printf("Error sending response: %v", err)
						}
						continue
					}
					
					room := manager.CreateRoom(roomID, msg.Content, msg.ProblemID, 4, 3600, msg.AccountID)
					err = queries.AddParticipantToRoom(r.Context(), sql.AddParticipantToRoomParams{
						RoomID:    int64ToPgInt8(roomID),
						AccountID: int64ToPgInt8(msg.AccountID),
					})
					if err != nil {
						log.Printf("Error adding participant to room: %v", err)
					}
					
					room.AddClient(msg.AccountID, conn)
					if err := sendResponse(conn, "room_created", fmt.Sprintf("Room created and joined: Room ID: %d", roomID)); err != nil {
						log.Printf("Error sending response: %v", err)
					}

				case "join":
					room, exists := manager.GetRoom(msg.RoomID)
					if !exists {
						if err := sendResponse(conn, "error", fmt.Sprintf("Room not found: %d", msg.RoomID)); err != nil {
							log.Printf("Error sending response: %v", err)
						}
						continue
					}
					room.AddClient(msg.AccountID, conn)
					
					// Insert participant into room_participant table
					err := queries.AddParticipantToRoom(r.Context(), sql.AddParticipantToRoomParams{
						RoomID:    int64ToPgInt8(msg.RoomID),
						AccountID: int64ToPgInt8(msg.AccountID),
					})
					if err != nil {
						log.Printf("Error adding participant to room: %v", err)
					}
					
					if err := sendResponse(conn, "joined", fmt.Sprintf("Joined room: %d", msg.RoomID)); err != nil {
						log.Printf("Error sending response: %v", err)
					}

				case "leave":
					room, exists := manager.GetRoom(msg.RoomID)
					if !exists {
						if err := sendResponse(conn, "error", fmt.Sprintf("Room not found: %d", msg.RoomID)); err != nil {
							log.Printf("Error sending response: %v", err)
						}
						continue
					}
					room.RemoveClient(msg.AccountID)
					
					// Update participant status in room_participant table
					err := queries.UpdateParticipantStatus(r.Context(), sql.UpdateParticipantStatusParams{
						RoomID:    int64ToPgInt8(msg.RoomID),
						AccountID: int64ToPgInt8(msg.AccountID),
						Status:    "inactive",
					})
					if err != nil {
						log.Printf("Error updating participant status: %v", err)
						continue
					}
					
					if err := sendResponse(conn, "left", fmt.Sprintf("Left room: %d", msg.RoomID)); err != nil {
						log.Printf("Error sending response: %v", err)
					}

				case "message":
					room, exists := manager.GetRoom(msg.RoomID)
					if !exists {
						if err := sendResponse(conn, "error", fmt.Sprintf("Room not found: %d", msg.RoomID)); err != nil {
							log.Printf("Error sending response: %v", err)
						}
						continue
					}
					
					// Insert message into room_message table
					err := queries.SaveRoomMessage(r.Context(), sql.SaveRoomMessageParams{
						RoomID:    int64ToPgInt8(msg.RoomID),
						AccountID: int64ToPgInt8(msg.AccountID),
						Content:   msg.Content,
					})
					if err != nil {
						log.Printf("Error saving message: %v", err)
					}
					
					room.Broadcast([]byte(msg.Content))

				case "list_accounts":
					room, exists := manager.GetRoom(msg.RoomID)
					if !exists {
						if err := sendResponse(conn, "error", fmt.Sprintf("Room not found: %d", msg.RoomID)); err != nil {
							log.Printf("Error sending response: %v", err)
						}
						continue
					}
					clients := room.GetClients()
					clientsJSON, _ := json.Marshal(clients)
					if err := sendResponse(conn, "accounts", string(clientsJSON)); err != nil {
						log.Printf("Error sending response: %v", err)
					}

				default:
					if err := sendResponse(conn, "error", "Unknown message type: "+msg.Type); err != nil {
						log.Printf("Error sending response: %v", err)
					}
				}
			}
		}
	}
}

func sendResponse(conn *websocket.Conn, responseType string, content interface{}) error {
	response := struct {
		Type    string      `json:"type"`
		Content interface{} `json:"content"`
	}{
		Type:    responseType,
		Content: content,
	}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return err
	}
	return conn.Write(context.Background(), websocket.MessageText, responseJSON)
}
