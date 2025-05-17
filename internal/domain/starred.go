package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type Starred struct {
	ID      interface{}
	Starred bool
}

// GetStarredProblem
type StarredProblem struct {
	Problem
	Starred bool
}

type StarredSolution struct {
	ID        int32
	Username  string
	Title     string
	Date      pgtype.Timestamp
	Tags      []string
	Body      string
	Votes     int32
	ProblemID int32
	Starred   bool
}

type StarredSubmission struct {
	ID            pgtype.UUID
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
	SolutionId int32
}

type StarSubmissionParams struct {
	UserId       string
	SubmissionId uuid.UUID
}

func FromSQLGetStarredSolutionsRows(rows []sql.GetStarredSolutionsRow) ([]*StarredSolution, error) {
	solutions := []*StarredSolution{}
	for _, row := range rows {
		solution := &StarredSolution{
			ID:        row.ID,
			Username:  row.Username,
			Title:     row.Title,
			Date:      row.CreatedAt,
			Tags:      row.Tags,
			Body:      row.Body,
			Votes:     nullHandler(row.Votes),
			ProblemID: nullHandler(row.ProblemID),
			Starred:   row.Starred,
		}

		solutions = append(solutions, solution)
	}

	return solutions, nil
}

func FromSQLGetStarredSubmissionRow(rows []sql.GetStarredSubmissionsRow) []*StarredSubmission {
	submissions := []*StarredSubmission{}
	for i, row := range rows {
		submissions[i] = &StarredSubmission{
			ID: row.ID,
			//Token: row.T
			Stdout:         nullHandler(row.Stdout),
			Time:           nullHandler(row.Time),
			Memory:         nullHandler(row.Memory),
			Stderr:         nullHandler(row.Stderr),
			CompileOutput:  nullHandler(row.CompileOutput),
			Message:        nullHandler(row.Message),
			Status:         row.Status,
			Language:       row.LanguageName,
			AccountID:      row.AccountID,
			SubmittedCode:  row.SubmittedCode,
			SubmittedStdin: nullHandler(row.SubmittedStdin),
			ProblemID:      row.ProblemID,
			CreatedAt:      row.CreatedAt.Time,
			Starred:        row.Starred,
		}
	}
	return submissions
}

func FromSQLGetStarredProblemsRow(rows []sql.GetStarredProblemsRow) []*StarredProblem {
	problems := []*StarredProblem{}
	for i, row := range rows {
		problems[i] = &StarredProblem{
			Problem: Problem{
				ID:           row.Problem.ID,
				Title:        row.Problem.Title,
				Description:  nullHandler(row.Problem.Description),
				FunctionName: row.Problem.FunctionName,
				Points:       row.Problem.Points,
				CreatedAt:    row.Problem.CreatedAt.Time,
				Difficulty:   row.Problem.Difficulty,
				Tags:         row.Problem.Tags,
			},
			Starred: row.Starred,
		}
	}

	return problems
}
