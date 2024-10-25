package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/src/apierror"
	"kadane.xyz/go-backend/v2/src/middleware"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

type Account struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	AvatarUrl string `json:"avatar_url"`
}

type AccountResponse struct {
	Data Account `json:"data"`
}

type AccountsResponse struct {
	Data []Account `json:"data"`
}

// GET: /accounts
func (h *Handler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.PostgresQueries.GetAccounts(r.Context())
	if err != nil {
		log.Println("Error getting accounts: ", err)
		apierror.SendError(w, http.StatusInternalServerError, "Error getting accounts")
	}

	accountsResponse := AccountsResponse{Data: []Account{}}
	for _, account := range accounts {
		accountsResponse.Data = append(accountsResponse.Data, Account{
			ID:       account.ID,
			Username: account.Username,
			Email:    account.Email,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accountsResponse)
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
	var createAccountRequest CreateAccountRequest
	err := json.NewDecoder(r.Body).Decode(&createAccountRequest)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid JSON format in request body")
		return
	}

	// Validate input fields
	if createAccountRequest.ID == "" {
		apierror.SendError(w, http.StatusBadRequest, "Account ID cannot be empty")
		return
	}
	if createAccountRequest.Username == "" {
		apierror.SendError(w, http.StatusBadRequest, "Username cannot be empty")
		return
	}
	if createAccountRequest.Email == "" {
		apierror.SendError(w, http.StatusBadRequest, "Email cannot be empty")
		return
	}

	// Validate email format
	if !isValidEmail(createAccountRequest.Email) {
		apierror.SendError(w, http.StatusBadRequest, "Invalid email format")
		return
	}

	// Create account in the database
	err = h.PostgresQueries.CreateAccount(r.Context(), sql.CreateAccountParams{
		ID:       createAccountRequest.ID,
		Username: createAccountRequest.Username,
		Email:    createAccountRequest.Email,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique_violation
				if pgErr.ConstraintName == "account_username_key" {
					apierror.SendError(w, http.StatusConflict, "Username already exists")
				} else if pgErr.ConstraintName == "account_email_key" {
					apierror.SendError(w, http.StatusConflict, "Email already exists")
				} else {
					apierror.SendError(w, http.StatusConflict, "Account with this ID already exists")
					return
				}
			default:
				log.Printf("Error creating account: %v", err)
				apierror.SendError(w, http.StatusInternalServerError, "Failed to create account")
				return
			}
		} else if errors.Is(err, pgx.ErrNoRows) {
			apierror.SendError(w, http.StatusNotFound, "Account not found")
			return
		} else {
			log.Printf("Error creating account: %v", err)
			apierror.SendError(w, http.StatusInternalServerError, "Failed to create account")
			return
		}
	}

	// Prepare response
	response := AccountResponse{
		Data: Account{
			ID:       createAccountRequest.ID,
			Username: createAccountRequest.Username,
			Email:    createAccountRequest.Email,
		},
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Helper function to validate email format
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(email)
}

// POST: /accounts/avatar
// Uploads an avatar image to S3 bucket and stores the URL in the accounts table
func (h *Handler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing user id")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Error reading request body")
		return
	}
	defer r.Body.Close()

	decodedContent, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Error decoding base64 string")
		return
	}

	// s3 bucket upload file and return url
	_, err = h.AWSClient.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket: &h.AWSBucketAvatar,
		Key:    &userId,
		Body:   bytes.NewReader(decodedContent),
	})
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Error uploading avatar")
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
		apierror.SendError(w, http.StatusInternalServerError, "Error updating avatar url")
		return
	}

	w.WriteHeader(http.StatusCreated)
}
