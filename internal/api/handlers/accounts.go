package handlers

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
	"io"
	"net/http"
	"net/mail"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/internal/api/httputils"
	"kadane.xyz/go-backend/v2/internal/config"
	"kadane.xyz/go-backend/v2/internal/database/repository"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
	"kadane.xyz/go-backend/v2/internal/errors"
	"kadane.xyz/go-backend/v2/internal/middleware"
)

const (
	maxFileSize  = 1 << 20 // 1 MB
	maxDimension = 3000
	minDimension = 500
)

type AccountHandler struct {
	repo      repository.AccountRepository
	awsClient *s3.Client
	config    *config.Config
}

func NewAccountHandler(repo repository.AccountRepository, awsClient *s3.Client, config *config.Config) *AccountHandler {
	return &AccountHandler{repo: repo, awsClient: awsClient, config: config}
}

func ValidateGetAccountsFiltered(r *http.Request) (*domain.AccountGetParams, error) {
	usernames, err := httputils.GetQueryParamStringArray(r, "usernames")
	if err != nil {
		return nil, err
	}

	locations, err := httputils.GetQueryParamStringArray(r, "locations")
	if err != nil {
		return nil, err
	}

	sort, err := httputils.GetQueryParam(r, "sort")
	if err != nil {
		return nil, err
	}
	if *sort != "level" {
		*sort = ""
	}

	order, err := httputils.GetQueryParamOrder(r)
	if err != nil {
		return nil, err
	}

	return &domain.AccountGetParams{
		UsernamesFilter:   usernames,
		LocationsFilter:   locations,
		Sort:              *sort,
		SortDirection:     sql.SortDirection(*order),
		IncludeAttributes: true,
	}, nil
}

// GET: /accounts
// Get all accounts with filtering
func (h *AccountHandler) GetAccounts(w http.ResponseWriter, r *http.Request) error {
	params, err := ValidateGetAccountsFiltered(r)
	if err != nil {
		return err
	}

	accounts, err := h.repo.ListAccounts(r.Context(), params)
	if err != nil {
		return errors.HandleDatabaseError(err, "accounts")
	}

	httputils.SendJSONDataResponse(w, http.StatusOK, accounts)

	return nil
}

func ValidateCreateAccount(r *http.Request) (*domain.AccountCreateRequest, error) {
	createAccountRequest, apiErr := httputils.DecodeJSONRequest[domain.AccountCreateRequest](r)
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

	return &createAccountRequest, nil
}

// POST: /accounts
func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) error {
	admin, err := middleware.GetContextUserIsAdmin(r.Context())
	if err != nil {
		return err
	}
	if !admin {
		return nil
	}

	// Validate request body
	createAccountRequest, err := ValidateCreateAccount(r)
	if err != nil {
		return err
	}

	// Create account in the database
	err = h.repo.CreateAccount(r.Context(), createAccountRequest)
	if err != nil {
		return errors.HandleDatabaseError(err, "account")
	}

	account, err := h.repo.GetAccount(r.Context(), &domain.AccountGetParams{
		ID:                createAccountRequest.ID,
		IncludeAttributes: true,
	})
	if err != nil {
		return errors.HandleDatabaseError(err, "account")
	}

	// Send response
	httputils.SendJSONDataResponse(w, http.StatusCreated, account)

	return nil
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
func (h *AccountHandler) UploadAccountAvatar(w http.ResponseWriter, r *http.Request) error {
	// Get userid from middleware context
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		return errors.NewApiError(nil, "Invalid content type", http.StatusBadRequest)
	}

	// Limit the max file size
	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize)

	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		return err
	}
	defer r.MultipartForm.RemoveAll()

	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		return err
	}
	defer file.Close()

	if fileHeader.Size > maxFileSize {
		return errors.NewApiError(nil, "File too large. Maximum size is 1MB", http.StatusRequestEntityTooLarge)
	}

	imageData, err := io.ReadAll(file)
	if err != nil {
		return errors.NewAppError(nil, "Error reading image file", http.StatusInternalServerError)
	}

	if err := validateImage(imageData); err != nil {
		return errors.NewAppError(nil, err.Error(), http.StatusInternalServerError)
	}

	// s3 bucket upload file and return url
	_, err = h.awsClient.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket: &h.config.AWS.BucketAvatar,
		Key:    &claims.UserID,
		Body:   bytes.NewReader(imageData),
	})
	if err != nil {
		return errors.NewAppError(nil, "Error uploading avatar", http.StatusInternalServerError)
	}

	// Use CloudFront URL instead of S3 public URL
	url := h.config.AWS.CloudFrontURL + "/" + claims.UserID

	// store image url in accounts table
	err = h.repo.UploadAccountAvatar(r.Context(), &domain.AccountAvatarParams{
		ID:        claims.UserID,
		AvatarUrl: url,
	})
	if err != nil {
		return err
	}

	httputils.SendJSONDataResponse(w, http.StatusCreated, url)

	return nil
}

func ValidateGetAccount(r *http.Request) (*domain.AccountGetParams, error) {
	idPtr, err := httputils.GetURLParam(r, "id")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	accountID := *idPtr

	attributes := false
	if attrPtr, err := httputils.GetQueryParamBool(r, "attributes"); err != nil && attrPtr != nil {
		attributes = *attrPtr
	}

	return &domain.AccountGetParams{
		ID:                accountID,
		IncludeAttributes: attributes,
	}, nil
}

// GET: /accounts/id
func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) error {
	params, err := ValidateGetAccount(r)
	if err != nil {
		return err
	}

	account, err := h.repo.GetAccount(r.Context(), params)
	if err != nil {
		httputils.EmptyDataResponse(w)
		return nil
	}

	httputils.SendJSONDataResponse(w, http.StatusOK, account)

	return nil
}

func ValidateUpdateAccount(r *http.Request) (*domain.AccountUpdateParams, error) {
	// Get account ID from URL parameters
	accountID, err := httputils.GetURLParam(r, "accountId")
	if err != nil {
		return nil, err
	}

	// Decode request body
	requestAttrs, err := httputils.DecodeJSONRequest[domain.AccountUpdateRequest](r)
	if err != nil {
		return nil, err
	}

	// Validate input fields
	if err := validateAccountAttributes(requestAttrs); err != nil {
		return nil, errors.NewBadRequestError(err.Error())
	}

	return &domain.AccountUpdateParams{
		ID: *accountID,
	}, nil
}

func prepareUpdateAccount(existingAccountAttributes *domain.AccountAttributes, requestBody *domain.AccountUpdateParams) (*domain.AccountUpdateParams, bool) {
	needsUpdate := false

	// Get existing account attributes
	updateParams := domain.AccountUpdateParams{
		ID: existingAccountAttributes.ID,
		AccountUpdateRequest: domain.AccountUpdateRequest{
			Bio:          &existingAccountAttributes.Bio,
			ContactEmail: &existingAccountAttributes.ContactEmail,
			Location:     &existingAccountAttributes.Location,
			RealName:     &existingAccountAttributes.RealName,
			GithubUrl:    &existingAccountAttributes.GithubUrl,
			LinkedinUrl:  &existingAccountAttributes.LinkedinUrl,
			FacebookUrl:  &existingAccountAttributes.FacebookUrl,
			InstagramUrl: &existingAccountAttributes.InstagramUrl,
			TwitterUrl:   &existingAccountAttributes.TwitterUrl,
			School:       &existingAccountAttributes.School,
			WebsiteUrl:   &existingAccountAttributes.WebsiteUrl,
		},
	}

	if requestBody.Bio != nil && *requestBody.Bio != existingAccountAttributes.Bio {
		updateParams.Bio = requestBody.Bio
		needsUpdate = true
	}

	if requestBody.ContactEmail != nil && *requestBody.ContactEmail != existingAccountAttributes.ContactEmail {
		updateParams.ContactEmail = requestBody.ContactEmail
		needsUpdate = true
	}

	if requestBody.Location != nil && *requestBody.Location != existingAccountAttributes.Location {
		updateParams.Location = requestBody.Location
		needsUpdate = true
	}

	if requestBody.RealName != nil && *requestBody.RealName != existingAccountAttributes.RealName {
		updateParams.RealName = requestBody.RealName
		needsUpdate = true
	}

	if requestBody.GithubUrl != nil && *requestBody.GithubUrl != existingAccountAttributes.GithubUrl {
		updateParams.GithubUrl = requestBody.GithubUrl
		needsUpdate = true
	}

	if requestBody.LinkedinUrl != nil && *requestBody.LinkedinUrl != existingAccountAttributes.LinkedinUrl {
		updateParams.LinkedinUrl = requestBody.LinkedinUrl
		needsUpdate = true
	}

	if requestBody.FacebookUrl != nil && *requestBody.FacebookUrl != existingAccountAttributes.FacebookUrl {
		updateParams.FacebookUrl = requestBody.FacebookUrl
		needsUpdate = true
	}

	if requestBody.InstagramUrl != nil && *requestBody.InstagramUrl != existingAccountAttributes.InstagramUrl {
		updateParams.InstagramUrl = requestBody.InstagramUrl
		needsUpdate = true
	}

	if requestBody.TwitterUrl != nil && *requestBody.TwitterUrl != existingAccountAttributes.TwitterUrl {
		updateParams.TwitterUrl = requestBody.TwitterUrl
		needsUpdate = true
	}

	if requestBody.School != nil && *requestBody.School != existingAccountAttributes.School {
		updateParams.School = requestBody.School
		needsUpdate = true
	}

	if requestBody.WebsiteUrl != nil && *requestBody.WebsiteUrl != existingAccountAttributes.WebsiteUrl {
		updateParams.WebsiteUrl = requestBody.WebsiteUrl
		needsUpdate = true
	}

	return &updateParams, needsUpdate
}

// PUT: /accounts/{id}
// Updates attributes for a given account
func (h *AccountHandler) UpdateAccount(w http.ResponseWriter, r *http.Request) error {
	params, err := ValidateUpdateAccount(r)
	if err != nil {
		return err
	}

	existingAttribute, err := h.repo.GetAccountAttributes(r.Context(), params.ID)
	if err != nil {
		return err
	}

	// Build update parameters
	updateAttributes, needsUpdate := prepareUpdateAccount(existingAttribute, params)
	if needsUpdate {
		// Update account in database
		account, err := h.repo.UpdateAccountAttributes(r.Context(), updateAttributes)
		if err != nil {
			return err
		}

		httputils.SendJSONDataResponse(w, http.StatusOK, account)

		return nil
	}

	// Get updated account attributes
	account, err := h.repo.GetAccount(r.Context(), &domain.AccountGetParams{
		ID:                params.ID,
		IncludeAttributes: true,
	})
	if err != nil {
		return err
	}

	// Send response
	httputils.SendJSONDataResponse(w, http.StatusOK, account)

	return nil
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
func (h *AccountHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) error {
	accountId, err := httputils.GetURLParam(r, "accountId")
	if err != nil {
		return err
	}

	err = h.repo.DeleteAccount(r.Context(), *accountId)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

// GET: /accounts/username
func (h *AccountHandler) GetAccountByUsername(w http.ResponseWriter, r *http.Request) error {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	username, err := httputils.GetURLParam(r, "username")
	if err != nil {
		return err
	}

	attributes, err := httputils.GetQueryParam(r, "attributes")
	if err != nil {
		return err
	}
	if *attributes == "" {
		*attributes = "false"
	}

	// check if account exists
	account, err := h.repo.GetAccountByUsername(r.Context(), sql.GetAccountByUsernameParams{
		Username:          *username,
		UserID:            claims.UserID,
		IncludeAttributes: *attributes == "true",
	})
	if err != nil {
		httputils.EmptyDataResponse(w)
		return err
	}

	httputils.SendJSONDataResponse(w, http.StatusOK, account)

	return nil
}

// GET: /accounts/validate
func (h *AccountHandler) GetAccountValidation(w http.ResponseWriter, r *http.Request) error {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	response := domain.AccountValidation{
		Plan: claims.Plan,
	}

	httputils.SendJSONDataResponse(w, http.StatusOK, response)

	return nil
}
