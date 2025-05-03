package domain

import (
	"encoding/json"
	"time"

	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/judge0"
)

type Submission struct {
	Id            string               `json:"id"`
	Token         string               `json:"token"`
	Stdout        string               `json:"stdout"`
	Time          string               `json:"time"`
	Memory        int32                `json:"memory"`
	Stderr        string               `json:"stderr"`
	CompileOutput string               `json:"compileOutput"`
	Message       string               `json:"message"`
	Status        sql.SubmissionStatus `json:"status"`
	Language      string               `json:"language"`
	// custom fields
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

type SubmissionRequest struct {
	Language   string `json:"language"`
	SourceCode string `json:"sourceCode"`
	ProblemID  int32  `json:"problemId"`
}

type SubmissionGetParams struct {
	Sort          sql.ProblemSort
	SortDirection sql.SortDirection
	UserId        string
	Username      string
	ProblemID     int64
	Status        sql.SubmissionStatus
}

func FromSQLGetSubmissionByUsernameRow(row sql.GetSubmissionsByUsernameRow) (*Submission, error) {
	stdout := ""
	if row.Stdout != nil {
		stdout = *row.Stdout
	}
	time := ""
	if row.Time != nil {
		time = *row.Time
	}
	memory := int32(0)
	if row.Memory != nil {
		memory = *row.Memory
	}
	stderr := ""
	if row.Stderr != nil {
		stderr = *row.Stderr
	}
	stdin := ""
	if row.SubmittedStdin != nil {
		stdin = *row.SubmittedStdin
	}
	compileOutput := ""
	if row.CompileOutput != nil {
		compileOutput = *row.CompileOutput
	}
	message := ""
	if row.Message != nil {
		message = *row.Message
	}
	failedTestCase := RunTestCase{}
	err := json.Unmarshal(row.FailedTestCase, &failedTestCase)
	if err != nil {
		return nil, err
	}
	passedTestCases := int32(0)
	if row.PassedTestCases != nil {
		passedTestCases = *row.PassedTestCases
	}
	totalTestCases := int32(0)
	if row.TotalTestCases != nil {
		totalTestCases = *row.TotalTestCases
	}
	return &Submission{
		Id:            row.ID.String(),
		Stdout:        stdout,
		Time:          time,
		Memory:        memory,
		Stderr:        stderr,
		CompileOutput: compileOutput,
		Message:       message,
		Status:        row.Status,
		Language:      judge0.LanguageIDToLanguage(int(row.LanguageID)),
		// custom fields
		AccountID:       row.AccountID,
		SubmittedCode:   row.SubmittedCode,
		SubmittedStdin:  stdin,
		ProblemID:       row.ProblemID,
		CreatedAt:       row.CreatedAt.Time,
		Starred:         row.Starred,
		FailedTestCase:  failedTestCase,
		PassedTestCases: passedTestCases,
		TotalTestCases:  totalTestCases,
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
