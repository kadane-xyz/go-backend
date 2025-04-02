package requests

import "kadane.xyz/go-backend/v2/internal/domain"

type AdminProblemRunRequest struct {
	FunctionName string            `json:"functionName"`
	Solutions    map[string]string `json:"solutions"` // ["language": "sourceCode"]
	TestCase     domain.TestCase   `json:"testCase"`
}
