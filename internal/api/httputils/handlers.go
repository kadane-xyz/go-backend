package httputils

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	request "kadane.xyz/go-backend/v2/internal/api/requests"
	"kadane.xyz/go-backend/v2/internal/domain"
	"kadane.xyz/go-backend/v2/internal/errors"
)

type ApiResponse struct {
	Data any `json:"data"`
}

type ApiPaginatedResponse struct {
	Data       any               `json:"data"`
	Pagination domain.Pagination `json:"pagination"`
}

func DecodeJSONRequest[T any](r *http.Request) (T, error) {
	var request T
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		var empty T
		return empty, err
	}
	return request, nil
}

// { data: {} }
func EmptyDataResponse(w http.ResponseWriter) {
	response := map[string]any{
		"data": map[string]any{},
	}
	json.NewEncoder(w).Encode(response)
}

// { data: [] }
func EmptyDataArrayResponse(w http.ResponseWriter) {
	response := map[string]any{
		"data": []any{},
	}
	json.NewEncoder(w).Encode(response)
}

// Send a JSON response with the given status code and data
func SendJSONResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// Send a JSON response with the given status code and data
// { data: {} }
func SendJSONDataResponse(w http.ResponseWriter, statusCode int, data any) {
	SendJSONResponse(w, statusCode, ApiResponse{
		Data: data,
	})
}

func SendJSONPaginatedResponse(w http.ResponseWriter, statusCode int, data any, pagination domain.Pagination) {
	SendJSONDataResponse(w, statusCode, ApiPaginatedResponse{
		Data:       data,
		Pagination: pagination,
	})
}

func GetQueryParam(r *http.Request, param string) (*string, error) {
	if param == "" {
		return nil, errors.ErrInternalServer
	}

	params := r.URL.Query().Get(param)
	if params == "" {
		return nil, errors.ErrUnprocessableEntity
	}

	return &params, nil
}

func GetQueryParamStringArray(r *http.Request, param string) ([]string, error) {
	var params []string
	p, err := GetQueryParam(r, param)
	if err != nil {
		return nil, err
	}

	if p != nil && *p != "" {
		params = strings.Split(*p, ",")
	}

	return params, nil
}

func GetQueryParamInt32(r *http.Request, param string) (int32, error) {
	p, err := GetQueryParam(r, param)
	if err != nil {
		return 0, err
	}

	var paramInt32 int32
	if p != nil {
		paramInt, err := strconv.ParseInt(*p, 10, 32)
		if err != nil {
			return 0, err
		}
		if paramInt < 1 {
			return 0, errors.NewAppError(nil, "param: "+param+" is less than 1", http.StatusBadRequest)
		}

		paramInt32 = int32(paramInt)
	}

	return paramInt32, nil
}

func GetQueryParamOrder(r *http.Request) (*string, error) {
	p, err := GetQueryParam(r, "order")
	if err != nil {
		return nil, err
	}

	if p == nil || *p == "" {
		return nil, nil
	}

	if orderParam := request.RequestQueryParamOrder(*p); !orderParam.IsValid() {
		return nil, errors.NewApiError(nil, "order param is invalid", http.StatusBadRequest)
	}

	return p, nil
}

func GetURLParam(r *http.Request, param string) (*string, error) {
	p := chi.URLParam(r, param)
	if p == "" {
		return nil, errors.NewApiError(nil, "Missing "+param, http.StatusBadRequest)
	}

	return &p, nil
}

func GetURLParamInt32(r *http.Request, param string) (int32, error) {
	p, err := GetURLParam(r, param)
	if err != nil {
		return 0, err
	}

	paramInt, err := strconv.ParseInt(*p, 10, 32)
	if err != nil {
		return 0, err
	}
	paramInt32 := int32(paramInt)

	return paramInt32, nil
}

func GetURLParamInt64(r *http.Request, param string) (int64, error) {
	p, err := GetURLParam(r, param)
	if err != nil {
		return 0, err
	}

	paramInt, err := strconv.ParseInt(*p, 10, 64)
	if err != nil {
		return 0, err
	}

	return paramInt, nil
}
