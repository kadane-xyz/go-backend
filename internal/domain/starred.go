package domain

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type Starred struct {
	ID      interface{}
	Starred bool
}

type StarredProblem struct {
	ID          int
	Title       string
	Description string
	Tags        []string
	Difficulty  string
	Points      int
	Starred     bool
}

type StarredSolution struct {
	Id        int64
	Username  string
	Title     string
	Date      pgtype.Timestamp
	Tags      []string
	Body      string
	Votes     int32
	ProblemId int32
	Starred   bool
}

type StarredSubmission struct {
	Id            pgtype.UUID
	Token         string
	Stdout        string
	Time          string
	Memory        int32
	Stderr        string
	CompileOutput string
	Message       string
	Status        sql.SubmissionStatus
	Language      string
	// custom fields
	AccountID      string
	SubmittedCode  string
	SubmittedStdin string
	ProblemID      int32
	CreatedAt      time.Time
	Starred        bool
}

type StarredRequest struct {
	ID any `json:"id"` // can be int32 or string
}

type StarProblemParams struct {
	UserId    string
	ProblemId int32
}

type StarSolutionParams struct {
	UserId     string
	SolutionId string
}

type StarSubmissionParams struct {
	UserId       string
	SubmissionId string
}

func FromSQLGetStarredSolutionsRows(rows []sql.GetStarredSolutionsRow) ([]*StarredSolution, error) {
	solutions := []*StarredSolution{}
	for _, row := range rows {
		votes := int32(0)
		if row.Votes != nil {
			votes = *row.Votes
		}
		problemId := int32(0)
		if row.ProblemID != nil {
			problemId = *row.ProblemID
		}
		solution := &StarredSolution{
			Id:        row.ID,
			Username:  row.Username,
			Title:     row.Title,
			Date:      row.CreatedAt,
			Tags:      row.Tags,
			Body:      row.Body,
			Votes:     votes,
			ProblemId: problemId,
			Starred:   row.Starred,
		}

		solutions = append(solutions, solution)
	}

	return solutions, nil
}

func FromSQLGetStarredSubmissionRow(rows []sql.GetStarredSubmissionsRow) []*StarredSubmission {
	submissions := []*StarredSubmission{}
	for _, row := range rows {
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

		compileOutput := ""
		if row.CompileOutput != nil {
			compileOutput = *row.CompileOutput
		}

		message := ""
		if row.Message != nil {
			message = *row.Message
		}

		submittedStdin := ""
		if row.SubmittedStdin != nil {
			submittedStdin = *row.SubmittedStdin
		}

		submissions = append(submissions, &StarredSubmission{
			Id: row.ID,
			//Token: row.T
			Stdout:         stdout,
			Time:           time,
			Memory:         memory,
			Stderr:         stderr,
			CompileOutput:  compileOutput,
			Message:        message,
			Status:         row.Status,
			Language:       row.LanguageName,
			AccountID:      row.AccountID,
			SubmittedCode:  row.SubmittedCode,
			SubmittedStdin: submittedStdin,
			ProblemID:      row.ProblemID,
			CreatedAt:      row.CreatedAt.Time,
			Starred:        row.Starred,
		})
	}
	return submissions
}
