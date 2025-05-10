package responses

import (
	"time"

	"kadane.xyz/go-backend/v2/internal/domain"
	"kadane.xyz/go-backend/v2/internal/judge0"
)

type SubmissionResponse struct {
	domain.Submission
}

func FromDomainSubmissionToApiSubmissionResponse(submission *domain.Submission) *SubmissionResponse {
	language := judge0.LanguageIDToLanguage(int(lastLanguageID))

	return &SubmissionResponse{
		domain.Submission{
			Id:              submissionId,
			Stdout:          avgSubmission.Stdout,
			Time:            avgSubmission.Time,
			Memory:          avgSubmission.Memory,
			Stderr:          avgSubmission.Stderr,
			CompileOutput:   avgSubmission.CompileOutput,
			Message:         avgSubmission.Message,
			Status:          avgSubmission.Status,
			Language:        language,
			AccountID:       userId,
			SubmittedCode:   request.SourceCode,
			SubmittedStdin:  "",
			ProblemID:       request.ProblemID,
			CreatedAt:       time.Now(),
			FailedTestCase:  failedTestCase,
			PassedTestCases: passedTestCases,
			TotalTestCases:  totalTestCases,
		},
	}
}

// Prepare response
