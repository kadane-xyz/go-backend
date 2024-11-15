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
	"net/mail"
	neturl "net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
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
	AvatarUrl string `json:"avatarUrl"`
	Level     int    `json:"level"`
}

type AccountResponse struct {
	Data Account `json:"data"`
}

type AccountsResponse struct {
	Data []Account `json:"data"`
}

type AccountAttributes struct {
	ID           string `json:"id,omitempty"`
	Bio          string `json:"bio,omitempty"`
	ContactEmail string `json:"contactEmail,omitempty"`
	Location     string `json:"location,omitempty"`
	RealName     string `json:"realName,omitempty"`
	GithubUrl    string `json:"githubUrl,omitempty"`
	LinkedinUrl  string `json:"linkedinUrl,omitempty"`
	FacebookUrl  string `json:"facebookUrl,omitempty"`
	InstagramUrl string `json:"instagramUrl,omitempty"`
	TwitterUrl   string `json:"twitterUrl,omitempty"`
	School       string `json:"school,omitempty"`
}

type AccountAttributesWithAccount struct {
	ID         string            `json:"id"`
	Username   string            `json:"username"`
	Email      string            `json:"email"`
	AvatarUrl  string            `json:"avatarUrl,omitempty"`
	Level      int               `json:"level"`
	Attributes AccountAttributes `json:"attributes,omitempty"`
}

type AccountAttributesResponse struct {
	Data AccountAttributesWithAccount `json:"data"`
}

type AccountAttributesUpdateResponse struct {
	Data AccountAttributes `json:"data"`
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

// POST: /accounts
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

	// Create account attributes in the database
	id, err := h.PostgresQueries.CreateAccountAttributes(r.Context(), sql.CreateAccountAttributesParams{
		ID:           createAccountRequest.ID,
		Bio:          pgtype.Text{String: "", Valid: true},
		Location:     pgtype.Text{String: "", Valid: true},
		RealName:     pgtype.Text{String: "", Valid: true},
		GithubUrl:    pgtype.Text{String: "", Valid: true},
		LinkedinUrl:  pgtype.Text{String: "", Valid: true},
		FacebookUrl:  pgtype.Text{String: "", Valid: true},
		InstagramUrl: pgtype.Text{String: "", Valid: true},
		TwitterUrl:   pgtype.Text{String: "", Valid: true},
		School:       pgtype.Text{String: "", Valid: true},
	})
	if err != nil {
		log.Println("Error creating account attributes: ", err)
		apierror.SendError(w, http.StatusInternalServerError, "Error creating account attributes")
		return
	}

	log.Println("Account attributes id: ", id)

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

// GET: /accounts/id
func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	accountId := chi.URLParam(r, "id")
	if accountId == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing account id")
		return
	}

	account, err := h.PostgresQueries.GetAccountAttributesWithAccount(r.Context(), accountId)
	if err != nil {
		log.Println("Error getting account: ", err)
		apierror.SendError(w, http.StatusInternalServerError, "Error getting account")
		return
	}

	response := AccountAttributesResponse{Data: AccountAttributesWithAccount{
		ID:        account.ID,
		Username:  account.Username,
		Email:     account.Email,
		AvatarUrl: account.AvatarUrl.String,
		Level:     int(account.Level.Int32),
		Attributes: AccountAttributes{
			Bio:          account.Bio.String,
			Location:     account.Location.String,
			RealName:     account.RealName.String,
			GithubUrl:    account.GithubUrl.String,
			LinkedinUrl:  account.LinkedinUrl.String,
			FacebookUrl:  account.FacebookUrl.String,
			InstagramUrl: account.InstagramUrl.String,
			TwitterUrl:   account.TwitterUrl.String,
			School:       account.School.String,
		},
	}}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// PUT: /accounts/id
func (h *Handler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	// Get account ID from URL parameters
	accountID := chi.URLParam(r, "id")
	if accountID == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing account ID")
		return
	}

	// Decode request body
	var requestAttrs AccountAttributes
	if err := json.NewDecoder(r.Body).Decode(&requestAttrs); err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid JSON format in request body")
		return
	}
	defer r.Body.Close()

	// Validate input fields
	if err := validateAccountAttributes(requestAttrs); err != nil {
		apierror.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get current account attributes
	currentAttrs, err := h.PostgresQueries.GetAccountAttributes(r.Context(), accountID)
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Error retrieving account attributes")
		return
	}

	newCurrentAttrs := AccountAttributes{
		ID:           currentAttrs.ID,
		Bio:          currentAttrs.Bio.String,
		ContactEmail: currentAttrs.ContactEmail.String,
		Location:     currentAttrs.Location.String,
		RealName:     currentAttrs.RealName.String,
		GithubUrl:    currentAttrs.GithubUrl.String,
		LinkedinUrl:  currentAttrs.LinkedinUrl.String,
		FacebookUrl:  currentAttrs.FacebookUrl.String,
		InstagramUrl: currentAttrs.InstagramUrl.String,
		TwitterUrl:   currentAttrs.TwitterUrl.String,
		School:       currentAttrs.School.String,
	}

	// Build update parameters
	updateParams := buildUpdateParams(requestAttrs, newCurrentAttrs)
	if !updateParams.HasChanges {
		apierror.SendError(w, http.StatusBadRequest, "No changes detected")
		return
	}

	// Update account in database
	_, err = h.PostgresQueries.UpdateAccountAttributes(r.Context(), updateParams.Params)
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to update account")
		return
	}

	account, err := h.PostgresQueries.GetAccountAttributesWithAccount(r.Context(), accountID)
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Error getting account")
		return
	}

	response := AccountAttributesResponse{Data: AccountAttributesWithAccount{
		ID:        account.ID,
		Username:  account.Username,
		Email:     account.Email,
		AvatarUrl: account.AvatarUrl.String,
		Level:     int(account.Level.Int32),
		Attributes: AccountAttributes{
			Bio:          account.Bio.String,
			Location:     account.Location.String,
			RealName:     account.RealName.String,
			GithubUrl:    account.GithubUrl.String,
			LinkedinUrl:  account.LinkedinUrl.String,
			FacebookUrl:  account.FacebookUrl.String,
			InstagramUrl: account.InstagramUrl.String,
			TwitterUrl:   account.TwitterUrl.String,
			School:       account.School.String,
		},
	}}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// AccountUpdates tracks which fields are being updated
type AccountUpdates struct {
	Bio          pgtype.Text
	ContactEmail pgtype.Text
	Location     pgtype.Text
	RealName     pgtype.Text
	GithubUrl    pgtype.Text
	LinkedinUrl  pgtype.Text
	FacebookUrl  pgtype.Text
	InstagramUrl pgtype.Text
	TwitterUrl   pgtype.Text
	School       pgtype.Text
	changes      bool
}

func (u *AccountUpdates) HasChanges() bool {
	return u.changes
}

type UpdateParamsResult struct {
	Params     sql.UpdateAccountAttributesParams
	HasChanges bool
}

// buildAccountUpdates compares request attributes with current attributes
// and returns an AccountUpdates with only the changed fields
func buildUpdateParams(req AccountAttributes, current AccountAttributes) UpdateParamsResult {
	result := UpdateParamsResult{
		Params: sql.UpdateAccountAttributesParams{
			ID: current.ID,
		},
		HasChanges: false,
	}

	// Helper function to check and set pgtype.Text fields
	setField := func(newVal, currentVal string) pgtype.Text {
		if newVal != "" { // If field is provided in request
			result.HasChanges = result.HasChanges || (newVal != currentVal)
			return pgtype.Text{String: newVal, Valid: true}
		}
		// Keep current value if no new value provided
		return pgtype.Text{String: currentVal, Valid: true}
	}

	// Update all fields, tracking changes
	result.Params.Bio = setField(req.Bio, current.Bio)
	result.Params.ContactEmail = setField(req.ContactEmail, current.ContactEmail)
	result.Params.Location = setField(req.Location, current.Location)
	result.Params.RealName = setField(req.RealName, current.RealName)
	result.Params.GithubUrl = setField(req.GithubUrl, current.GithubUrl)
	result.Params.LinkedinUrl = setField(req.LinkedinUrl, current.LinkedinUrl)
	result.Params.FacebookUrl = setField(req.FacebookUrl, current.FacebookUrl)
	result.Params.InstagramUrl = setField(req.InstagramUrl, current.InstagramUrl)
	result.Params.TwitterUrl = setField(req.TwitterUrl, current.TwitterUrl)
	result.Params.School = setField(req.School, current.School)

	return result
}

// validateAccountAttributes performs validation on account attributes
func validateAccountAttributes(attrs AccountAttributes) error {
	// Add any validation rules here, for example:
	if attrs.ContactEmail != "" {
		if !isValidEmail(attrs.ContactEmail) {
			return fmt.Errorf("invalid email format")
		}
	}

	// URL validation
	urls := map[string]string{
		"GitHub":    attrs.GithubUrl,
		"LinkedIn":  attrs.LinkedinUrl,
		"Facebook":  attrs.FacebookUrl,
		"Instagram": attrs.InstagramUrl,
		"Twitter":   attrs.TwitterUrl,
	}

	for platform, url := range urls {
		if url != "" && !isValidURL(url) {
			return fmt.Errorf("invalid %s URL format", platform)
		}
	}

	return nil
}

// Helper functions for validation
func isValidEmail(email string) bool {
	// Basic email validation
	_, err := mail.ParseAddress(email)
	return err == nil
}

func isValidURL(url string) bool {
	// Basic URL validation
	_, err := neturl.ParseRequestURI(url)
	return err == nil
}

// GET: /accounts/username
func (h *Handler) GetAccountByUsername(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing username")
		return
	}

	account, err := h.PostgresQueries.GetAccountByUsername(r.Context(), username)
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Error getting account")
		return
	}

	response := AccountResponse{Data: Account{
		ID:       account.ID,
		Username: account.Username,
		Email:    account.Email,
	}}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
