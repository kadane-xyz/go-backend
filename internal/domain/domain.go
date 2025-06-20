package domain

import (
	"encoding/json"

	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type Pagination struct {
	Page      int32 `json:"page"`
	PerPage   int32 `json:"perPage"`
	DataCount int32 `json:"dataCount"`
	LastPage  int32 `json:"lastPage"`
}

type TestCaseType string

const (
	IntType         TestCaseType = "int"
	FloatType       TestCaseType = "float"
	DoubleType      TestCaseType = "double"
	StringType      TestCaseType = "string"
	BoolType        TestCaseType = "bool"
	IntArrayType    TestCaseType = "int[]"
	FloatArrayType  TestCaseType = "float[]"
	DoubleArrayType TestCaseType = "double[]"
	StringArrayType TestCaseType = "string[]"
	BoolArrayType   TestCaseType = "bool[]"
)

type TestCaseInput struct {
	Name  string       `json:"name"`  // name of the input
	Type  TestCaseType `json:"type"`  // type of the input
	Value string       `json:"value"` // value of the input
}

type TestCase struct {
	Description string          `json:"description"`
	Input       []TestCaseInput `json:"input"`
	Output      string          `json:"output"`
	Visibility  sql.Visibility  `json:"visibility"`
}

const (
	FriendshipStatusNone            FriendshipStatus = "none"
	FriendshipStatusFriend          FriendshipStatus = "friend"
	FriendshipStatusBlocked         FriendshipStatus = "blocked"
	FriendshipStatusRequestSent     FriendshipStatus = "requestSent"
	FriendshipStatusRequestReceived FriendshipStatus = "requestReceived"
)

type VoteRequest struct {
	Vote sql.VoteType `json:"vote"`
}

// nullHandler returns *ptr if ptr != nil, otherwise the zero value of T.
func nullHandler[T any](ptr *T) T {
	if ptr != nil {
		return *ptr
	}
	var zero T
	return zero
}

// jsonArrayHandler returns []*T for handling interface{} types
func jsonArrayHandler[T any](raw []byte) ([]*T, error) {
	var hs []T
	err := json.Unmarshal(raw, &hs)
	if err != nil {
		return nil, err
	}

	out := make([]*T, len(hs))
	for i := range hs {
		out[i] = &hs[i]
	}

	return out, nil
}

// jsonHandler returns *T for handling interface{} types
func jsonHandler[T any](raw []byte) (*T, error) {
	var hs T
	err := json.Unmarshal(raw, &hs)
	if err != nil {
		return nil, err
	}

	return &hs, nil
}
