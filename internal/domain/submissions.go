package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/judge0"
)

type Submission struct {
	ID            uuid.UUID            `json:"id"`
	Stdout        string               `json:"stdout"`
	Time          string               `json:"time"`
	Memory        int32                `json:"memory"`
	Stderr        string               `json:"stderr"`
	CompileOutput string               `json:"compileOutput"`
	Message       string               `json:"message"`
	Status        sql.SubmissionStatus `json:"status"`
	Language      string               `json:"language"`
	// relations
	AccountID       string      `json:"accountId"`
	SubmittedCode   string      `json:"submittedCode"`
	SubmittedStdin  string      `json:"submittedStdin"`
	ProblemID       int32       `json:"problemId"`
	CreatedAt       time.Time   `json:"createdAt"`
	Starred         bool        `json:"starred"`
	FailedTestCase  RunTestCase `json:"failedTestCase,omitempty"`
	PassedTestCases int32       `json:"passedTestCases"`
	TotalTestCases  int32       `json:"totalTestCases"`
}

type SubmissionCreateParams struct {
	ID              uuid.UUID
	Stdout          string
	Time            string
	Memory          int32
	Stderr          string
	CompileOutput   string
	Message         string
	Status          string
	LanguageID      int32
	LanguageName    string
	AccountID       string
	ProblemID       int32
	SubmittedCode   string
	SubmittedStdin  string
	FailedTestCase  []byte
	PassedTestCases int32
	TotalTestCases  int32
}

type SubmissionCreateRequest struct {
	Language   string `json:"language"`
	SourceCode string `json:"sourceCode"`
	ProblemID  int32  `json:"problemId"`
}

type SubmissionGetParams struct {
	SubmissionID uuid.UUID
	UserID       string
}

type SubmissionsGetByUsernameParams struct {
	Username  string
	ProblemID int32
	Status    sql.SubmissionStatus
	Sort      sql.ProblemSort
	Order     sql.SortDirection
	Page      int32
	PerPage   int32
}

func FromSQLGetSubmissionByUsernameRow(row sql.GetSubmissionsByUsernameRow) (*Submission, error) {
	failedTestCase := RunTestCase{}
	err := json.Unmarshal(row.FailedTestCase, &failedTestCase)
	if err != nil {
		return nil, err
	}

	return &Submission{
		ID:            row.ID.Bytes,
		Stdout:        nullHandler(row.Stdout),
		Time:          nullHandler(row.Time),
		Memory:        nullHandler(row.Memory),
		Stderr:        nullHandler(row.Stderr),
		CompileOutput: nullHandler(row.CompileOutput),
		Message:       nullHandler(row.Message),
		Status:        row.Status,
		Language:      judge0.LanguageIDToLanguage(int(row.LanguageID)),
		// custom fields
		AccountID:       row.AccountID,
		SubmittedCode:   row.SubmittedCode,
		SubmittedStdin:  nullHandler(row.SubmittedStdin),
		ProblemID:       row.ProblemID,
		CreatedAt:       row.CreatedAt.Time,
		Starred:         row.Starred,
		FailedTestCase:  failedTestCase,
		PassedTestCases: nullHandler(row.PassedTestCases),
		TotalTestCases:  nullHandler(row.TotalTestCases),
	}, nil
}

func FromSQLGetSubmissionByUsernameRows(rows []sql.GetSubmissionsByUsernameRow) ([]*Submission, error) {
	submissions := []*Submission{}
	for _, row := range rows {
		submission, err := FromSQLGetSubmissionByUsernameRow(row)
		if err != nil {
			return nil, err
		}
		submissions = append(submissions, submission)
	}

	return submissions, nil
}

func FromSQLGetSubmissionRow(row sql.GetSubmissionRow) (*Submission, error) {
	failedTestCase := RunTestCase{}
	if row.FailedTestCase != nil {
		if err := json.Unmarshal(row.FailedTestCase, &failedTestCase); err != nil {
			return nil, err
		}
	}

	return &Submission{
		ID:            row.ID.Bytes,
		Stdout:        nullHandler(row.Stdout),
		Time:          nullHandler(row.Time),
		Memory:        nullHandler(row.Memory),
		Stderr:        nullHandler(row.Stderr),
		CompileOutput: nullHandler(row.CompileOutput),
		Message:       nullHandler(row.Message),
		Status:        row.Status,
		Language:      judge0.LanguageIDToLanguage(int(row.LanguageID)),
		// custom fields
		AccountID:       row.AccountID,
		SubmittedCode:   row.SubmittedCode,
		SubmittedStdin:  nullHandler(row.SubmittedStdin),
		ProblemID:       row.ProblemID,
		CreatedAt:       row.CreatedAt.Time,
		Starred:         row.Starred,
		FailedTestCase:  failedTestCase,
		PassedTestCases: nullHandler(row.PassedTestCases),
		TotalTestCases:  nullHandler(row.TotalTestCases),
	}, nil
}

func FromSQLSubmission(row sql.Submission) (*Submission, error) {

	failedTestCase := RunTestCase{}
	if row.FailedTestCase != nil {
		if err := json.Unmarshal(row.FailedTestCase, &failedTestCase); err != nil {
			return nil, err
		}
	}

	return &Submission{
		ID:            row.ID.Bytes,
		Stdout:        nullHandler(row.Stdout),
		Time:          nullHandler(row.Time),
		Memory:        nullHandler(row.Memory),
		Stderr:        nullHandler(row.Stderr),
		CompileOutput: nullHandler(row.CompileOutput),
		Message:       nullHandler(row.Message),
		Status:        row.Status,
		Language:      judge0.LanguageIDToLanguage(int(row.LanguageID)),
		// custom fields
		AccountID:       row.AccountID,
		SubmittedCode:   row.SubmittedCode,
		SubmittedStdin:  *row.SubmittedStdin,
		ProblemID:       row.ProblemID,
		CreatedAt:       row.CreatedAt.Time,
		FailedTestCase:  failedTestCase,
		PassedTestCases: *row.PassedTestCases,
		TotalTestCases:  *row.TotalTestCases,
	}, nil
}

func FromSQLSubmissions(rows []sql.Submission) ([]*Submission, error) {
	submissions := []*Submission{}

	for _, row := range rows {
		submission, err := FromSQLSubmission(row)
		if err != nil {
			return nil, err
		}
		submissions = (append(submissions, submission))
	}

	return submissions, nil
}
