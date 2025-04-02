package requests

type ProblemRequest struct {
	Title        string               `json:"title"`
	Description  string               `json:"description"`
	FunctionName string               `json:"functionName"`
	Tags         []string             `json:"tags"`
	Difficulty   string               `json:"difficulty"`
	Code         ProblemRequestCode   `json:"code"`
	Hints        []ProblemRequestHint `json:"hints"`
	Points       int32                `json:"points"`
	Solutions    map[string]string    `json:"solutions"` // ["language": "sourceCode"]
	TestCases    []TestCase           `json:"testCases"`
}
