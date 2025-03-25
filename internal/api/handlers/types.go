package api

import "kadane.xyz/go-backend/v2/src/sql/sql"

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

type FriendshipStatus string

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
