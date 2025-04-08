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

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/internal/api/httputils"
	"kadane.xyz/go-backend/v2/internal/config"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
	"kadane.xyz/go-backend/v2/internal/errors"
	"kadane.xyz/go-backend/v2/internal/services"
)

const (
	maxFileSize  = 1 << 20 // 1 MB
	maxDimension = 3000
	minDimension = 500
)

type AccountHandler struct {
	service   services.AccountService
	awsClient *s3.Client
	config    *config.Config
}

func NewAccountHandler(service services.AccountService, awsClient *s3.Client, config *config.Config) *AccountHandler {
	return &AccountHandler{service: service, awsClient: awsClient, config: config}
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
		errors.SendError(w, http.StatusBadRequest, "Failed to validate get accounts filtered")
		return
	}

	accounts, err := h.accountService.ListAccounts(r.Context(), params)
	if err != nil {
		errors.SendError(w, http.StatusInternalServerError, "Failed to get accounts filtered")
		return
	}

	httputils.SendJSONDataResponse(w, http.StatusOK, accounts)
}

func ValidateCreateAccount(r *http.Request) (*sql.CreateAccountParams, *errors.ApiError) {
	createAccountRequest, apiErr := httputils.DecodeJSONRequest[domain.CreateAccountRequest](r)
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
		errors.SendError(w, http.StatusForbidden, "You are not authorized to create accounts")
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

// POST: /accounts/avatar
// Uploads an avatar image to S3 bucket and stores the URL in the accounts table
func (h *AccountHandler) UploadAccountAvatar(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := httputils.GetClientUserID(w, r)
	if err != nil {
		return
	}

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		errors.SendError(w, http.StatusBadRequest, "Invalid content type")
		return
	}

	// Limit the max file size
	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize)

	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		errors.SendError(w, http.StatusBadRequest, "File too large. Maximum size is 1MB")
		return
	}
	defer r.MultipartForm.RemoveAll()

	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		errors.SendError(w, http.StatusBadRequest, "Error getting image file")
		return
	}
	defer file.Close()

	if fileHeader.Size > maxFileSize {
		errors.SendError(w, http.StatusBadRequest, "File too large. Maximum size is 1MB")
		return
	}

	imageData, err := io.ReadAll(file)
	if err != nil {
		errors.SendError(w, http.StatusInternalServerError, "Error reading image file")
		return
	}

	if err := validateImage(imageData); err != nil {
		errors.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// s3 bucket upload file and return url
	_, err = h.awsClient.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket: &h.config.AWS.BucketAvatar,
		Key:    &userId,
		Body:   bytes.NewReader(imageData),
	})
	if err != nil {
		errors.SendError(w, http.StatusInternalServerError, "Error uploading avatar")
		return
	}

	// Use CloudFront URL instead of S3 public URL
	url := h.config.AWS.CloudFrontURL + "/" + userId

	// store image url in accounts table
	err = h.accessor.UploadAccountAvatar(r.Context(), sql.UpdateAccountAvatarParams{
		ID:        userId,
		AvatarUrl: pgtype.Text{String: url, Valid: true},
	})
	if err != nil {
		errors.SendError(w, http.StatusInternalServerError, "Error updating avatar url")
		return
	}

	httputils.SendJSONDataResponse(w, http.StatusCreated, url)
}

func ValidateGetAccount(r *http.Request) (sql.GetAccountParams, *errors.ApiError) {
	accountId := chi.URLParam(r, "id")
	if accountId == "" {
		return sql.GetAccountParams{}, errors.NewBadRequestError("Missing account id")
	}

	attributes := r.URL.Query().Get("attributes")
	if attributes == "" {
		attributes = "false"
	}

	return sql.GetAccountParams{
		ID:                accountId,
		IncludeAttributes: attributes == "true",
	}, nil
}

// GET: /accounts/id
func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	params, err := ValidateGetAccount(r)
	if err != nil {
		errors.SendError(w, http.StatusBadRequest, "Failed to validate get account")
		return
	}

	account, dbErr := h.accessor.GetAccount(r.Context(), params)
	if dbErr != nil {
		httputils.EmptyDataResponse(w)
		return
	}

	httputils.SendJSONDataResponse(w, http.StatusOK, account)
}

func ValidateUpdateAccount(r *http.Request) (sql.UpdateAccountAttributesParams, *errors.ApiError) {
	// Get account ID from URL parameters
	accountID := chi.URLParam(r, "id")
	if accountID == "" {
		return sql.UpdateAccountAttributesParams{}, errors.NewBadRequestError("Missing account ID")
	}

	// Decode request body
	var requestAttrs domain.AccountUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&requestAttrs); err != nil {
		return sql.UpdateAccountAttributesParams{}, errors.NewBadRequestError("Invalid JSON format in request body")
	}
	defer r.Body.Close()

	// Validate input fields
	if err := validateAccountAttributes(requestAttrs); err != nil {
		return sql.UpdateAccountAttributesParams{}, errors.NewBadRequestError(err.Error())
	}

	return sql.UpdateAccountAttributesParams{
		ID: accountID,
	}, nil
}

func prepareUpdateAccount(existingAccountAttributes sql.AccountAttribute, requestBody sql.UpdateAccountAttributesParams) (*sql.UpdateAccountAttributesParams, bool) {
	needsUpdate := false

	// Get existing account attributes
	updateParams := sql.UpdateAccountAttributesParams{
		ID:           existingAccountAttributes.ID,
		Bio:          existingAccountAttributes.Bio.String,
		ContactEmail: existingAccountAttributes.ContactEmail.String,
		Location:     existingAccountAttributes.Location.String,
		RealName:     existingAccountAttributes.RealName.String,
		GithubUrl:    existingAccountAttributes.GithubUrl.String,
		LinkedinUrl:  existingAccountAttributes.LinkedinUrl.String,
		FacebookUrl:  existingAccountAttributes.FacebookUrl.String,
		InstagramUrl: existingAccountAttributes.InstagramUrl.String,
		TwitterUrl:   existingAccountAttributes.TwitterUrl.String,
		School:       existingAccountAttributes.School.String,
		WebsiteUrl:   existingAccountAttributes.WebsiteUrl.String,
	}

	if requestBody.Bio != existingAccountAttributes.Bio.String {
		updateParams.Bio = requestBody.Bio
		needsUpdate = true
	}

	if requestBody.ContactEmail != existingAccountAttributes.ContactEmail.String {
		updateParams.ContactEmail = requestBody.ContactEmail
		needsUpdate = true
	}

	if requestBody.Location != existingAccountAttributes.Location.String {
		updateParams.Location = requestBody.Location
		needsUpdate = true
	}

	if requestBody.RealName != existingAccountAttributes.RealName.String {
		updateParams.RealName = requestBody.RealName
		needsUpdate = true
	}

	if requestBody.GithubUrl != existingAccountAttributes.GithubUrl.String {
		updateParams.GithubUrl = requestBody.GithubUrl
		needsUpdate = true
	}

	if requestBody.LinkedinUrl != existingAccountAttributes.LinkedinUrl.String {
		updateParams.LinkedinUrl = requestBody.LinkedinUrl
		needsUpdate = true
	}

	if requestBody.FacebookUrl != existingAccountAttributes.FacebookUrl.String {
		updateParams.FacebookUrl = requestBody.FacebookUrl
		needsUpdate = true
	}

	if requestBody.InstagramUrl != existingAccountAttributes.InstagramUrl.String {
		updateParams.InstagramUrl = requestBody.InstagramUrl
		needsUpdate = true
	}

	if requestBody.TwitterUrl != existingAccountAttributes.TwitterUrl.String {
		updateParams.TwitterUrl = requestBody.TwitterUrl
		needsUpdate = true
	}

	if requestBody.School != existingAccountAttributes.School.String {
		updateParams.School = requestBody.School
		needsUpdate = true
	}

	if requestBody.WebsiteUrl != existingAccountAttributes.WebsiteUrl.String {
		updateParams.WebsiteUrl = requestBody.WebsiteUrl
		needsUpdate = true
	}

	return &updateParams, needsUpdate
}

// PUT: /accounts/{id}
// Updates attributes for a given account
func (h *AccountHandler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	params, apiErr := ValidateUpdateAccount(r)
	if apiErr != nil {
		apiErr.Send(w)
		return
	}

	existingAttribute, err := h.accessor.GetAccountAttributes(r.Context(), params.ID)
	if err != nil {
		errors.SendError(w, http.StatusInternalServerError, "Error getting account attributes")
		return
	}

	// Build update parameters
	updateAttributes, needsUpdate := prepareUpdateAccount(existingAttribute, params)
	if needsUpdate {
		// Update account in database
		account, err := h.accessor.UpdateAccountAttributes(r.Context(), *updateAttributes)
		if err != nil {
			errors.SendError(w, http.StatusInternalServerError, "Failed to update account")
			return
		}

		httputils.SendJSONDataResponse(w, http.StatusOK, account)
	}

	// Get updated account attributes
	account, err := h.accessor.GetAccount(r.Context(), sql.GetAccountParams{
		ID:                params.ID,
		IncludeAttributes: true,
	})
	if err != nil {
		errors.SendError(w, http.StatusInternalServerError, "Error getting account")
		return
	}

	// Send response
	httputils.SendJSONDataResponse(w, http.StatusOK, account)
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
func buildUpdateParams(req domain.AccountUpdateRequest, current domain.AccountAttributes) UpdateParamsResult {
	result := UpdateParamsResult{
		Params: sql.UpdateAccountAttributesParams{
			ID: current.ID,
		},
		HasChanges: false,
	}

	// Helper function to check and set pgtype.Text fields
	setField := func(newVal *string, currentVal string) string {
		if newVal != nil { // If field was provided in request (including empty string)
			result.HasChanges = result.HasChanges || (*newVal != currentVal)
			return *newVal
		}
		// Keep current value if field not provided in request
		return currentVal
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
func validateAccountAttributes(attrs domain.AccountUpdateRequest) error {
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
		errors.SendError(w, http.StatusBadRequest, "Missing account ID")
		return
	}

	err := h.accessor.DeleteAccount(r.Context(), accountId)
	if err != nil {
		errors.SendError(w, http.StatusInternalServerError, "Error deleting account")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET: /accounts/username
func (h *AccountHandler) GetAccountByUsername(w http.ResponseWriter, r *http.Request) {
	userID, err := httputils.GetClientUserID(w, r)
	if err != nil {
		return
	}

	username := chi.URLParam(r, "username")
	if username == "" {
		errors.SendError(w, http.StatusBadRequest, "Missing username")
		return
	}

	attributes := r.URL.Query().Get("attributes")
	if attributes == "" {
		attributes = "false"
	}

	// check if account exists
	account, err := h.accessor.GetAccountByUsername(r.Context(), sql.GetAccountByUsernameParams{
		Username:          username,
		UserID:            userID,
		IncludeAttributes: attributes == "true",
	})
	if err != nil {
		httputils.EmptyDataResponse(w)
		return
	}

	httputils.SendJSONDataResponse(w, http.StatusOK, account)
}

// GET: /accounts/validate
func (h *AccountHandler) GetAccountValidation(w http.ResponseWriter, r *http.Request) {
	accountPlan, err := httputils.GetClientPlan(w, r)
	if err != nil {
		return
	}

	response := domain.AccountValidation{
		Plan: accountPlan,
	}

	httputils.SendJSONDataResponse(w, http.StatusOK, response)
}
