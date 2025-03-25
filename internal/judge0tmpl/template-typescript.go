package judge0tmpl

import (
	"fmt"
	"strings"

	"kadane.xyz/go-backend/v2/internal/judge0"
)

// Convert the test case inputs to a comma separated string
func TemplateTypescriptInputs(testCases TestCase) string {
	var inputs []string

	for _, input := range testCases.Input {
		inputValue := strings.Trim(input.Value, "[]")
		switch input.Type {
		case "float", "double", "int", "boolean":
			inputs = append(inputs, inputValue) // TypeScript uses number for both
		case "int[]", "float[]", "double[]", "string[]", "boolean[]":
			inputs = append(inputs, fmt.Sprintf("[%s]", inputValue))
		case "string":
			inputs = append(inputs, fmt.Sprintf("\"%s\"", inputValue))
		}
	}
	return strings.Join(inputs, ",")
}

// Typescript template
func TemplateTypescriptSourceCode(functionName string, inputs string, sourceCode string) string {
	return fmt.Sprintf(`
%s

console.log(%s(%s))
`, sourceCode, functionName, inputs)
}

func TemplateTypescript(templateInput TemplateInput) judge0.Submission {
	inputs := TemplateTypescriptInputs(templateInput.TestCase)                                               // Get the inputs
	sourceCode := TemplateTypescriptSourceCode(templateInput.FunctionName, inputs, templateInput.SourceCode) // Get the source code

	submission := judge0.Submission{
		LanguageID: judge0.LanguageToLanguageID("typescript"),
		SourceCode: sourceCode,
		//ExpectedOutput: templateInput.ExpectedOutput,
	}

	return submission
}
