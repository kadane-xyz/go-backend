package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/src/apierror"
	"kadane.xyz/go-backend/v2/src/middleware"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

const (
	maxFileSize  = 1 << 20 // 1 MB
	maxDimension = 3000
	minDimension = 500
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

type AvatarResponse struct {
	Data string `json:"data"`
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

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		apierror.SendError(w, http.StatusBadRequest, "Invalid content type")
		return
	}

	// Limit the max file size
	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize)

	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		apierror.SendError(w, http.StatusBadRequest, "File too large. Maximum size is 1MB")
		return
	}
	defer r.MultipartForm.RemoveAll()

	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Error getting image file")
		return
	}
	defer file.Close()

	if fileHeader.Size > maxFileSize {
		apierror.SendError(w, http.StatusBadRequest, "File too large. Maximum size is 1MB")
		return
	}

	imageData, err := io.ReadAll(file)
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Error reading image file")
		return
	}

	if err := validateImage(imageData); err != nil {
		apierror.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// s3 bucket upload file and return url
	_, err = h.AWSClient.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket: &h.AWSBucketAvatar,
		Key:    &userId,
		Body:   bytes.NewReader(imageData),
	})
	if err != nil {
		log.Println("Error uploading avatar: ", err)
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

	response := AvatarResponse{Data: url}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// validateImage checks image type and dimensions
func validateImage(imageData []byte) error {
	img, format, err := image.DecodeConfig(bytes.NewReader(imageData))
	if err != nil {
		return fmt.Errorf("invalid image format")
	}

	if format != "jpeg" && format != "png" {
		return fmt.Errorf("unsupported image format. Only JPEG and PNG are supported")
	}

	if img.Width > maxDimension || img.Height > maxDimension {
		return fmt.Errorf("image dimensions too large. Maximum is %dx%d", maxDimension, maxDimension)
	}

	if img.Width < minDimension || img.Height < minDimension {
		return fmt.Errorf("image dimensions too small. Minimum is %dx%d", minDimension, minDimension)
	}

	return nil
}
