package domain

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/errors"
	"kadane.xyz/go-backend/v2/internal/judge0"
)

type Submission struct {
	Id            uuid.UUID            `json:"id"`
	Stdout        string               `json:"stdout"`
	Time          time.Duration        `json:"time"`
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

type SubmissionCreateParams struct {
	Id              uuid.UUID
	Stdout          string
	Time            time.Time
	Memory          int32
	Stderr          string
	CompileOutput   string
	Message         string
	Status          string
	LanguageId      int32
	LanguageName    string
	AccountId       string
	ProblemId       int32
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
	SubmissionId uuid.UUID
	UserId       string
}

type SubmissionsGetByUsernameParams struct {
	Username  string
	ProblemId int32
	Status    sql.SubmissionStatus
	Sort      sql.ProblemSort
	Order     sql.SortDirection
	Page      int32
	PerPage   int32
}

// TransformSubmissionResults converts database results to API response format
func TransformSubmissionResults(submissions []sql.GetSubmissionsByUsernameRow) ([]domain.Submission, error) {
	submissionResults := make([]domain.Submission, 0, len(submissions))

	for _, submission := range submissions {
		submissionId := uuid.UUID(submission.ID.Bytes)
		submissionFailedTestCase := domain.RunTestCase{}

		err := json.Unmarshal(submission.FailedTestCase, &submissionFailedTestCase)
		if err != nil {
			return nil, errors.NewAppError(err, "Failed to unmarshal failed test case", http.StatusInternalServerError)

		}

		stdout := ""
		if submission.Stdout != nil {
			stdout = *submission.Stdout
		}

		submissionResults = append(submissionResults, domain.Submission{
			Id:              submissionId.String(),
			Stdout:          stdout,
			Time:            submission.Time,
			Memory:          submission.Memory,
			Stderr:          submission.Stderr,
			CompileOutput:   submission.CompileOutput,
			Message:         submission.Message,
			Status:          submission.Status,
			Language:        judge0.LanguageIDToLanguage(int(submission.LanguageID)),
			AccountID:       submission.AccountID,
			SubmittedCode:   submission.SubmittedCode,
			SubmittedStdin:  submission.SubmittedStdin,
			ProblemID:       submission.ProblemID,
			CreatedAt:       submission.CreatedAt.Time,
			Starred:         submission.Starred,
			FailedTestCase:  submissionFailedTestCase,
			PassedTestCases: submission.PassedTestCases,
			TotalTestCases:  submission.TotalTestCases,
		})
	}

	return submissionResults, nil
}

func FromSQLGetSubmissionByUsernameRow(row sql.GetSubmissionsByUsernameRow) (*Submission, error) {
	stdout := ""
	if row.Stdout != nil {
		stdout = *row.Stdout
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
		Id:            row.ID.Bytes,
		Stdout:        stdout,
		Time:          row.Time.Duration,
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

func FromSQLGetSubmissionRow(row sql.GetSubmissionRow) (*Submission, error) {
	stdout := ""
	if row.Stdout != nil {
		stdout = *row.Stdout
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
	passedTestCases := int32(0)
	if row.PassedTestCases != nil {
		passedTestCases = *row.PassedTestCases
	}
	totalTestCases := int32(0)
	if row.TotalTestCases != nil {
		totalTestCases = *row.TotalTestCases
	}
	failedTestCase := RunTestCase{}
	if row.FailedTestCase != nil {
		if err := json.Unmarshal(row.FailedTestCase, &failedTestCase); err != nil {
			return nil, err
		}
	}

	return &Submission{
		Id:            row.ID.Bytes,
		Stdout:        stdout,
		Time:          row.Time.Time,
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

func FromSQLSubmission(row sql.Submission) (*Submission, error) {

	failedTestCase := RunTestCase{}
	if row.FailedTestCase != nil {
		if err := json.Unmarshal(row.FailedTestCase, &failedTestCase); err != nil {
			return nil, err
		}
	}

	return &Submission{
		Id:            row.ID.Bytes,
		Stdout:        *row.Stdout,
		Time:          row.Time.Time,
		Memory:        *row.Memory,
		Stderr:        *row.Stderr,
		CompileOutput: *row.CompileOutput,
		Message:       *row.Message,
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
