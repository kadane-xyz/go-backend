package responses

import (
	"encoding/json"
	"time"

	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
	"kadane.xyz/go-backend/v2/internal/judge0"
)

type SubmissionResponse struct {
	domain.Submission
}

func FromDomainSubmissionCreateParamsToApiSubmissionResponse(submission *domain.SubmissionCreateParams) (*SubmissionResponse, error) {
	language := judge0.LanguageIDToLanguage(int(submission.LanguageID))

	var failedTestCase domain.RunTestCase
	if submission.FailedTestCase != nil {
		if err := json.Unmarshal(submission.FailedTestCase, &failedTestCase); err != nil {
			return nil, err
		}
	}

	return &SubmissionResponse{
		domain.Submission{
			ID:              submission.ID,
			Stdout:          submission.Stdout,
			Time:            submission.Time,
			Memory:          submission.Memory,
			Stderr:          submission.Stderr,
			CompileOutput:   submission.CompileOutput,
			Message:         submission.Message,
			Status:          sql.SubmissionStatus(submission.Status),
			Language:        language,
			AccountID:       submission.AccountID,
			SubmittedCode:   submission.SubmittedCode,
			SubmittedStdin:  "",
			ProblemID:       submission.ProblemID,
			CreatedAt:       time.Now(),
			FailedTestCase:  failedTestCase,
			PassedTestCases: submission.PassedTestCases,
			TotalTestCases:  submission.TotalTestCases,
		},
	}, nil
}
