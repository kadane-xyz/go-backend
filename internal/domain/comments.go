package domain

import "time"

type Comment struct {
	ID         int64      `json:"id"`
	SolutionId int64      `json:"solutionId"`
	Body       string     `json:"body"`
	CreatedAt  time.Time  `json:"createdAt"`
	Votes      int32      `json:"votes"`
	ParentId   int64      `json:"parentId,omitempty"`
	Children   []*Comment `json:"children,omitempty"` // For nested child comments
}
