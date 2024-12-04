package api

import (
	"encoding/json"
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/src/apierror"
	"kadane.xyz/go-backend/v2/src/middleware"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

type RoomRequest struct {
	Name       string   `json:"name"`
	MaxPlayers int      `json:"maxPlayers"`
	Visibility string   `json:"visibility"`
	Problems   []string `json:"problems"`
	Difficulty string   `json:"difficulty"`
	Mode       string   `json:"mode"`
	Whitelist  []string `json:"whitelist"`
}

type RoomResponse struct {
	Data sql.Room `json:"data"`
}

type RoomsResponse struct {
	Data []sql.Room `json:"data"`
}

func (h *Handler) GetRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.PostgresQueries.GetRooms(r.Context())
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to get rooms")
		return
	}

	response := RoomsResponse{
		Data: rooms,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing user ID for room creation")
		return
	}

	var roomRequest RoomRequest
	err := json.NewDecoder(r.Body).Decode(&roomRequest)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	adminId, err := uuid.Parse(userId)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	problems := make([]pgtype.UUID, len(roomRequest.Problems))
	for i, problem := range roomRequest.Problems {
		id, err := uuid.Parse(problem)
		if err != nil {
			apierror.SendError(w, http.StatusBadRequest, "Invalid problem ID")
			return
		}
		problems[i] = pgtype.UUID{Bytes: id, Valid: true}
	}

	whitelist := make([]pgtype.UUID, len(roomRequest.Whitelist))
	for i, user := range roomRequest.Whitelist {
		id, err := uuid.Parse(user)
		if err != nil {
			apierror.SendError(w, http.StatusBadRequest, "Invalid user ID")
			return
		}
		whitelist[i] = pgtype.UUID{Bytes: id, Valid: true}
	}

	room, err := h.PostgresQueries.CreateRoom(r.Context(), sql.CreateRoomParams{
		Admin:      pgtype.UUID{Bytes: adminId, Valid: true},
		Name:       roomRequest.Name,
		MaxPlayers: int32(roomRequest.MaxPlayers),
		Visibility: sql.Visibility(roomRequest.Visibility),
		Problems:   problems,
		Difficulty: roomRequest.Difficulty,
		Mode:       sql.Mode(roomRequest.Mode),
		Whitelist:  whitelist,
	})

	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to create room")
		return
	}

	response := RoomResponse{
		Data: room,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetRoom(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing user ID")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing room ID")
		return
	}

	roomId, err := uuid.Parse(id)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	room, err := h.PostgresQueries.GetRoom(r.Context(), pgtype.UUID{Bytes: roomId, Valid: true})
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to get room")
		return
	}

	roomAdminId := uuid.UUID(room.Admin.Bytes).String()

	accountId, err := h.PostgresQueries.GetAccountIDByUsername(r.Context(), userId)
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to get account ID")
		return
	}

	accountIdBytes, err := uuid.Parse(accountId)
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to get account ID")
		return
	}

	if roomAdminId != accountId && !slices.Contains(room.Whitelist, pgtype.UUID{Bytes: accountIdBytes, Valid: true}) {
		apierror.SendError(w, http.StatusForbidden, "You are not permitted to view this room")
		return
	}

	response := RoomResponse{
		Data: room,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
