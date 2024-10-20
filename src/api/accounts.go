package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/src/middleware"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

// GET: /accounts
func (h *Handler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Get Accounts"))
	return
}

func GetS3PublicURL(bucketName, region, objectKey string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, region, objectKey)
}

type CreateAccountRequest struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// Post: /accounts
func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var createAccountRequest CreateAccountRequest
	err = json.Unmarshal(body, &createAccountRequest)
	if err != nil {
		http.Error(w, "Error unmarshalling request body", http.StatusInternalServerError)
		return
	}

	err = h.PostgresQueries.CreateAccount(r.Context(), sql.CreateAccountParams{
		ID:       createAccountRequest.ID,
		Username: createAccountRequest.Username,
		Email:    createAccountRequest.Email,
	})
	if err != nil {
		log.Println("Error creating account: ", err)
		http.Error(w, "Error creating account", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// POST: /accounts/avatar
// Uploads an avatar image to S3 bucket and stores the URL in the accounts table
func (h *Handler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		http.Error(w, "Missing user id", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	decodedContent, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		http.Error(w, "Error decoding base64 string", http.StatusInternalServerError)
		return
	}

	// s3 bucket upload file and return url
	_, err = h.AWSClient.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket: &h.AWSBucketAvatar,
		Key:    &userId,
		Body:   bytes.NewReader(decodedContent),
	})
	if err != nil {
		log.Println("Error uploading avatar: ", err)
		http.Error(w, "Error uploading avatar", http.StatusInternalServerError)
		return
	}

	// Get image url
	url := GetS3PublicURL(h.AWSBucketAvatar, h.AWSRegion, userId)
	log.Println("Avatar URL: ", url)

	// store image url in accounts table
	err = h.PostgresQueries.UpdateAvatar(r.Context(), sql.UpdateAvatarParams{
		ID:        userId,
		AvatarUrl: pgtype.Text{String: url, Valid: true},
	})
	if err != nil {
		log.Println("Error updating avatar url: ", err)
		http.Error(w, "Error updating avatar url", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
