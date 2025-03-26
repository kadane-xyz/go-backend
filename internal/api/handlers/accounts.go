package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
	"io"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/internal/api/httputils"
	"kadane.xyz/go-backend/v2/internal/database/dbaccessors"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/errors"
)

const (
	maxFileSize  = 1 << 20 // 1 MB
	maxDimension = 3000
	minDimension = 500
)

type AccountHandler struct {
	accessor dbaccessors.AccountAccessor
}

func NewAccountHandler(accessor dbaccessors.AccountAccessor) *AccountHandler {
	return &AccountHandler{accessor: accessor}
}

type AccountUpdateRequest struct {
	Bio          *string `json:"bio,omitempty"`
	ContactEmail *string `json:"contactEmail,omitempty"`
	Location     *string `json:"location,omitempty"`
	RealName     *string `json:"realName,omitempty"`
	GithubUrl    *string `json:"githubUrl,omitempty"`
	LinkedinUrl  *string `json:"linkedinUrl,omitempty"`
	FacebookUrl  *string `json:"facebookUrl,omitempty"`
	InstagramUrl *string `json:"instagramUrl,omitempty"`
	TwitterUrl   *string `json:"twitterUrl,omitempty"`
	School       *string `json:"school,omitempty"`
	WebsiteUrl   *string `json:"websiteUrl,omitempty"`
}

type AccountAttributes struct {
	ID                 string `json:"id,omitempty"`
	Bio                string `json:"bio,omitempty"`
	ContactEmail       string `json:"contactEmail,omitempty"`
	Location           string `json:"location,omitempty"`
	RealName           string `json:"realName,omitempty"`
	GithubUrl          string `json:"githubUrl,omitempty"`
	LinkedinUrl        string `json:"linkedinUrl,omitempty"`
	FacebookUrl        string `json:"facebookUrl,omitempty"`
	InstagramUrl       string `json:"instagramUrl,omitempty"`
	TwitterUrl         string `json:"twitterUrl,omitempty"`
	School             string `json:"school,omitempty"`
	WebsiteUrl         string `json:"websiteUrl,omitempty"`
	FriendCount        int64  `json:"friends,omitempty"`
	BlockedCount       int64  `json:"blockedUsers,omitempty"`
	FriendRequestCount int64  `json:"friendRequests,omitempty"`
}

type Account struct {
	ID           string           `json:"id"`
	Username     string           `json:"username"`
	Email        string           `json:"email"`
	AvatarUrl    string           `json:"avatarUrl,omitempty"`
	Level        int32            `json:"level"`
	CreatedAt    time.Time        `json:"createdAt"`
	FriendStatus FriendshipStatus `json:"friendStatus,omitempty"`
	Plan         sql.AccountPlan  `json:"plan"`
	IsAdmin      bool             `json:"isAdmin"`
	Attributes   interface{}      `json:"attributes"`
}

type AccountValidation struct {
	Plan sql.AccountPlan `json:"plan"`
}

func ValidateGetAccountsFiltered(r *http.Request) (sql.ListAccountsWithAttributesFilteredParams, error) {
	usernames := r.URL.Query().Get("usernames")
	var usernamesFilter []string
	if usernames != "" {
		usernamesFilter = strings.Split(usernames, ",")
	}

	locations := r.URL.Query().Get("locations")
	var locationsFilter []string
	if locations != "" {
		locationsFilter = strings.Split(locations, ",")
	}

	sort := r.URL.Query().Get("sort")
	if sort != "level" {
		sort = ""
	}

	order := r.URL.Query().Get("order")
	if order == "asc" {
		order = "ASC"
	} else {
		order = "DESC"
	}

	return sql.ListAccountsWithAttributesFilteredParams{
		UsernamesFilter:   usernamesFilter,
		LocationsFilter:   locationsFilter,
		Sort:              sort,
		SortDirection:     order,
		IncludeAttributes: true,
	}, nil
}

// GET: /accounts
// Get all accounts with filtering
func (h *AccountHandler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	params, err := ValidateGetAccountsFiltered(r)
	if err != nil {
		errors.SendError(w, err, http.StatusBadRequest, "Failed to validate get accounts filtered")
		return
	}

	accounts, err := h.accessor.ListAccountsWithAttributesFiltered(r.Context(), params)
	if err != nil {
		errors.SendError(w, err, http.StatusInternalServerError, "Failed to get accounts filtered")
		return
	}

	// accounts response
	response := []Account{}
	for _, account := range accounts {
		if account.Attributes == nil {
			account.Attributes = AccountAttributes{}
		}

		response = append(response, Account{
			ID:         account.ID,
			Username:   account.Username,
			Email:      account.Email,
			CreatedAt:  account.CreatedAt.Time,
			AvatarUrl:  account.AvatarUrl.String,
			Level:      account.Level,
			Plan:       account.Plan,
			IsAdmin:    account.Admin,
			Attributes: account.Attributes,
		})
	}

	httputils.SendJSONDataResponse(w, http.StatusOK, response)
}

type CreateAccountRequest struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func ValidateCreateAccount(r *http.Request) (*sql.CreateAccountParams, *errors.ApiError) {
	createAccountRequest, apiErr := httputils.DecodeJSONRequest[CreateAccountRequest](r)
	if apiErr != nil {
		return nil, errors.NewUnprocessableEntityError("Invalid request body")
	}

	// Validate input fields
	if createAccountRequest.ID == "" {
		return nil, errors.NewBadRequestError("Missing account id")
	}
	if createAccountRequest.Username == "" {
		return nil, errors.NewBadRequestError("Missing username")
	}
	if createAccountRequest.Email == "" {
		return nil, errors.NewBadRequestError("Missing email")
	}

	// Validate email format
	if !isValidEmail(createAccountRequest.Email) {
		return nil, errors.NewBadRequestError("Invalid email format")
	}
	return &sql.CreateAccountParams{
		ID:       createAccountRequest.ID,
		Username: createAccountRequest.Username,
		Email:    createAccountRequest.Email,
	}, nil
}

// POST: /accounts
func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	admin := httputils.GetClientAdmin(w, r)
	if !admin {
		errors.SendError(w, nil, http.StatusForbidden, "You are not authorized to create accounts")
		return
	}

	// Validate request body
	createAccountRequest, apiErr := ValidateCreateAccount(r)
	if apiErr != nil {
		apiErr.Send(w)
		return
	}

	// Create account in the database
	err := h.accessor.CreateAccount(r.Context(), *createAccountRequest)
	if err != nil {
		apiErr.Send(w)
		return
	}

	// Create account attributes in the database
	_, err = h.accessor.CreateAccountAttributes(r.Context(), sql.CreateAccountAttributesParams{
		ID:           createAccountRequest.ID,
		Bio:          pgtype.Text{String: "", Valid: true},
		ContactEmail: pgtype.Text{String: "", Valid: true},
		Location:     pgtype.Text{String: "", Valid: true},
		RealName:     pgtype.Text{String: "", Valid: true},
		GithubUrl:    pgtype.Text{String: "", Valid: true},
		LinkedinUrl:  pgtype.Text{String: "", Valid: true},
		FacebookUrl:  pgtype.Text{String: "", Valid: true},
		InstagramUrl: pgtype.Text{String: "", Valid: true},
		TwitterUrl:   pgtype.Text{String: "", Valid: true},
		School:       pgtype.Text{String: "", Valid: true},
		WebsiteUrl:   pgtype.Text{String: "", Valid: true},
	})
	if err != nil {
		apiErr.Send(w)
		return
	}

	account, err := h.accessor.GetAccount(r.Context(), sql.GetAccountParams{
		ID:                createAccountRequest.ID,
		IncludeAttributes: true,
	})
	if err != nil {
		apiErr.Send(w)
		return
	}

	// Send response
	httputils.SendJSONDataResponse(w, http.StatusCreated, account)
}

// POST: /accounts/avatar
// Uploads an avatar image to S3 bucket and stores the URL in the accounts table
func (h *AccountHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		errors.SendError(w, errors.NewBadRequestError("Invalid content type"), http.StatusBadRequest, "Invalid content type")
		return
	}

	// Limit the max file size
	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize)

	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		errors.SendError(w, errors.NewBadRequestError("File too large. Maximum size is 1MB"), http.StatusBadRequest, "File too large. Maximum size is 1MB")
		return
	}
	defer r.MultipartForm.RemoveAll()

	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		errors.SendError(w, errors.NewBadRequestError("Error getting image file"), http.StatusBadRequest, "Error getting image file")
		return
	}
	defer file.Close()

	if fileHeader.Size > maxFileSize {
		errors.SendError(w, errors.NewBadRequestError("File too large. Maximum size is 1MB"), http.StatusBadRequest, "File too large. Maximum size is 1MB")
		return
	}

	imageData, err := io.ReadAll(file)
	if err != nil {
		errors.SendError(w, errors.NewInternalServerError("Error reading image file"), http.StatusInternalServerError, "Error reading image file")
		return
	}

	if err := validateImage(imageData); err != nil {
		errors.SendError(w, errors.NewBadRequestError(err.Error()), http.StatusBadRequest, err.Error())
		return
	}

	// s3 bucket upload file and return url
	_, err = h.AWSClient.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket: &h.AWSBucketAvatar,
		Key:    &userId,
		Body:   bytes.NewReader(imageData),
	})
	if err != nil {
		errors.SendError(w, errors.NewInternalServerError("Error uploading avatar"), http.StatusInternalServerError, "Error uploading avatar")
		return
	}

	// Use CloudFront URL instead of S3 public URL
	url := h.CloudFrontUrl + "/" + userId

	// store image url in accounts table
	err = h.PostgresQueries.UpdateAvatar(r.Context(), sql.UpdateAvatarParams{
		ID:        userId,
		AvatarUrl: pgtype.Text{String: url, Valid: true},
	})
	if err != nil {
		errors.SendError(w, errors.NewInternalServerError("Error updating avatar url"), http.StatusInternalServerError, "Error updating avatar url")
		return
	}

	response := AvatarResponse{Data: url}

	httputils.SendJSONDataResponse(w, http.StatusCreated, response)
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
func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	accountId := chi.URLParam(r, "id")
	if accountId == "" {
		errors.SendError(w, errors.NewBadRequestError("Missing account id"), http.StatusBadRequest, "Missing account id")
		return
	}

	attributes := r.URL.Query().Get("attributes")
	if attributes == "" {
		attributes = "false"
	}

	account, err := h.PostgresQueries.GetAccount(r.Context(), sql.GetAccountParams{
		ID:                accountId,
		IncludeAttributes: attributes == "true",
	})
	if err != nil {
		httputils.EmptyDataResponse(w)
		return
	}

	response := AccountResponse{Data: Account{
		ID:         account.ID,
		Username:   account.Username,
		Email:      account.Email,
		CreatedAt:  account.CreatedAt.Time,
		AvatarUrl:  account.AvatarUrl.String,
		Level:      account.Level,
		Plan:       account.Plan,
		IsAdmin:    account.Admin,
		Attributes: account.Attributes,
	}}

	httputils.SendJSONDataResponse(w, http.StatusOK, response)
}

// PUT: /accounts/id
func (h *AccountHandler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	// Get account ID from URL parameters
	accountID := chi.URLParam(r, "id")
	if accountID == "" {
		errors.SendError(w, errors.NewBadRequestError("Missing account ID"), http.StatusBadRequest, "Missing account ID")
		return
	}

	// Decode request body
	var requestAttrs AccountUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&requestAttrs); err != nil {
		errors.SendError(w, errors.NewBadRequestError("Invalid JSON format in request body"), http.StatusBadRequest, "Invalid JSON format in request body")
		return
	}
	defer r.Body.Close()

	// Validate input fields
	if err := validateAccountAttributes(requestAttrs); err != nil {
		errors.SendError(w, errors.NewBadRequestError(err.Error()), http.StatusBadRequest, err.Error())
		return
	}

	// Get current account attributes or create new relation if none exist
	currentAttrs, err := h.PostgresQueries.GetAccountAttributes(r.Context(), accountID)
	if err != nil {
		h.PostgresQueries.CreateAccountAttributes(r.Context(), sql.CreateAccountAttributesParams{
			ID: accountID,
		})
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
		WebsiteUrl:   currentAttrs.WebsiteUrl.String,
	}

	// Build update parameters
	updateParams := buildUpdateParams(requestAttrs, newCurrentAttrs)
	if !updateParams.HasChanges {
		errors.SendError(w, errors.NewBadRequestError("No changes detected"), http.StatusBadRequest, "No changes detected")
		return
	}

	// Update account in database
	_, err = h.PostgresQueries.UpdateAccountAttributes(r.Context(), updateParams.Params)
	if err != nil {
		errors.SendError(w, errors.NewInternalServerError("Failed to update account"), http.StatusInternalServerError, "Failed to update account")
		return
	}

	// Get updated account attributes
	account, err := h.PostgresQueries.GetAccount(r.Context(), sql.GetAccountParams{
		ID:                accountID,
		IncludeAttributes: true,
	})
	if err != nil {
		errors.SendError(w, errors.NewInternalServerError("Error getting account"), http.StatusInternalServerError, "Error getting account")
		return
	}

	// Prepare response
	response := AccountResponse{Data: Account{
		ID:         account.ID,
		Username:   account.Username,
		Email:      account.Email,
		CreatedAt:  account.CreatedAt.Time,
		AvatarUrl:  account.AvatarUrl.String,
		Level:      account.Level,
		Plan:       account.Plan,
		IsAdmin:    account.Admin,
		Attributes: account.Attributes,
	}}

	// Send response
	httputils.SendJSONDataResponse(w, http.StatusOK, response)
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
	WebsiteUrl   pgtype.Text
	changes      bool
}

func (u *AccountUpdates) HasChanges() bool {
	return u.changes
}

type UpdateParamsResult struct {
	Params     sql.UpdateAccountAttributesParams
	HasChanges bool
}

// buildUpdateParams compares request attributes with current attributes
// and returns an UpdateParamsResult with all provided fields, including empty strings
func buildUpdateParams(req AccountUpdateRequest, current AccountAttributes) UpdateParamsResult {
	result := UpdateParamsResult{
		Params: sql.UpdateAccountAttributesParams{
			ID: current.ID,
		},
		HasChanges: false,
	}

	// Helper function to check and set pgtype.Text fields
	setField := func(newVal *string, currentVal string) pgtype.Text {
		if newVal != nil { // If field was provided in request (including empty string)
			result.HasChanges = result.HasChanges || (*newVal != currentVal)
			return pgtype.Text{String: *newVal, Valid: true}
		}
		// Keep current value if field not provided in request
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
	result.Params.WebsiteUrl = setField(req.WebsiteUrl, current.WebsiteUrl)

	return result
}

// validateAccountAttributes performs validation on account attributes
func validateAccountAttributes(attrs AccountUpdateRequest) error {
	// Only validate non-empty email addresses
	if attrs.ContactEmail != nil && *attrs.ContactEmail != "" {
		if !isValidEmail(*attrs.ContactEmail) {
			return fmt.Errorf("invalid email format")
		}
	}

	// Only validate non-empty locations
	if attrs.Location != nil && *attrs.Location != "" && len(*attrs.Location) > 2 {
		return fmt.Errorf("location field too long")
	}

	// Helper function to check string length
	checkLength := func(value *string, fieldName string) error {
		if value != nil && len(*value) > 50 {
			return fmt.Errorf("%s field too long", fieldName)
		}
		return nil
	}

	// Validate length limits for provided fields
	fields := map[string]*string{
		"Bio":          attrs.Bio,
		"ContactEmail": attrs.ContactEmail,
		"RealName":     attrs.RealName,
		"GithubUrl":    attrs.GithubUrl,
		"LinkedinUrl":  attrs.LinkedinUrl,
		"TwitterUrl":   attrs.TwitterUrl,
		"FacebookUrl":  attrs.FacebookUrl,
		"InstagramUrl": attrs.InstagramUrl,
		"School":       attrs.School,
		"WebsiteUrl":   attrs.WebsiteUrl,
	}

	for fieldName, value := range fields {
		if err := checkLength(value, fieldName); err != nil {
			return err
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

// DELETE: /accounts/id
func (h *AccountHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	accountId := chi.URLParam(r, "id")
	if accountId == "" {
		errors.SendError(w, errors.NewBadRequestError("Missing account ID"), http.StatusBadRequest, "Missing account ID")
		return
	}

	err := h.PostgresQueries.DeleteAccount(r.Context(), accountId)
	if err != nil {
		errors.SendError(w, errors.NewInternalServerError("Error deleting account"), http.StatusInternalServerError, "Error deleting account")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET: /accounts/username
func (h *AccountHandler) GetAccountByUsername(w http.ResponseWriter, r *http.Request) {
	userID, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	username := chi.URLParam(r, "username")
	if username == "" {
		errors.SendError(w, errors.NewBadRequestError("Missing username"), http.StatusBadRequest, "Missing username")
		return
	}

	attributes := r.URL.Query().Get("attributes")
	if attributes == "" {
		attributes = "false"
	}

	// check if account exists
	account, err := h.PostgresQueries.GetAccountByUsername(r.Context(), sql.GetAccountByUsernameParams{
		Username:          username,
		UserID:            userID,
		IncludeAttributes: attributes == "true",
	})
	if err != nil {
		httputils.EmptyDataResponse(w)
		return
	}

	response := AccountResponse{Data: Account{
		ID:           account.ID,
		Username:     account.Username,
		Email:        account.Email,
		CreatedAt:    account.CreatedAt.Time,
		AvatarUrl:    account.AvatarUrl.String,
		Level:        account.Level,
		Plan:         account.Plan,
		IsAdmin:      account.Admin,
		FriendStatus: FriendshipStatus(account.FriendStatus),
		Attributes:   account.Attributes,
	}}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GET: /accounts/validate
func (h *AccountHandler) GetAccountValidation(w http.ResponseWriter, r *http.Request) {
	accountPlan, err := GetClientPlan(w, r)
	if err != nil {
		return
	}

	response := AccountValidationResponse{
		Data: AccountValidation{
			Plan: accountPlan,
		},
	}

	httputils.SendJSONDataResponse(w, http.StatusOK, response)
}
