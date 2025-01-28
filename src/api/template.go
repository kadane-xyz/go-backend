package api

import "kadane.xyz/go-backend/v2/src/judge0"

type TemplateInput struct {
	Language       string   `json:"language"`
	FunctionName   string   `json:"functionName"`
	SourceCode     string   `json:"sourceCode"`
	ExpectedOutput string   `json:"expectedOutput"`
	Problem        Problem  `json:"problem"`
	TestCases      TestCase `json:"testCases"`
}

func TemplateCreate(templateInput TemplateInput) judge0.Submission {
	switch templateInput.Language {
	case "cpp":
		return TemplateCpp(templateInput)
	case "go":
		return TemplateGo(templateInput)
	case "java":
		return TemplateJava(templateInput)
	case "typescript":
		return TemplateTypescript(templateInput)
	}
	return judge0.Submission{}
}
