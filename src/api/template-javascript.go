package api

import (
	"fmt"
	"strings"

	"kadane.xyz/go-backend/v2/src/judge0"
)

// Convert the test case inputs to a comma separated string
func TemplateJavascriptInputs(testCases TestCase) string {
	var inputs []string

	for _, input := range testCases.Input {
		inputValue := strings.Trim(input.Value, "[]")
		switch input.Type {
		case "string":
			// Strings need quotes
			inputs = append(inputs, fmt.Sprintf("\"%s\"", inputValue))
		case "int[]", "float[]", "double[]", "string[]", "boolean[]":
			// Arrays just need square brackets
			inputs = append(inputs, fmt.Sprintf("[%s]", inputValue))
		default:
			// Numbers and booleans can be passed as-is
			inputs = append(inputs, inputValue)
		}
	}
	return strings.Join(inputs, ",")
}

// Javascript template
func TemplateJavascriptSourceCode(functionName string, inputs string, sourceCode string) string {
	return fmt.Sprintf(`
%s

%s(%s)
`, sourceCode, functionName, inputs)
}

func TemplateJavascript(templateInput TemplateInput) judge0.Submission {
	inputs := TemplateJavascriptInputs(templateInput.TestCases)                                              // Get the inputs
	sourceCode := TemplateJavascriptSourceCode(templateInput.FunctionName, inputs, templateInput.SourceCode) // Get the source code

	submission := judge0.Submission{
		LanguageID: judge0.LanguageToLanguageID("javascript"),
		SourceCode: sourceCode,
		//ExpectedOutput: templateInput.ExpectedOutput,
	}

	return submission
}
