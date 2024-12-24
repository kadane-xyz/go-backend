package api

import "kadane.xyz/go-backend/v2/src/sql/sql"

type Pagination struct {
	Page      int64 `json:"page"`
	PerPage   int64 `json:"perPage"`
	DataCount int64 `json:"dataCount"`
	LastPage  int64 `json:"lastPage"`
}

type TestCase struct {
	Input      []byte         `json:"input"`  //base64 encoded
	Output     []byte         `json:"output"` //base64 encoded
	Visibility sql.Visibility `json:"visibility"`
}
