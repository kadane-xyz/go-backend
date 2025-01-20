package api

import (
	"fmt"
	"strings"

	"kadane.xyz/go-backend/v2/src/judge0"
)

// Convert the type of the test case to the type of the Go language
// Convert the test case inputs to a comma separated string
func RunGoTemplateInputs(testCases TestCase) string {
	var inputs []string

	for _, input := range testCases.Input {
		inputValue := strings.Trim(input.Value, "[]")
		switch input.Type {
		case IntType:
			inputs = append(inputs, inputValue)
		case "float":
			inputs = append(inputs, fmt.Sprintf("float32(%s)", inputValue))
		case "double":
			inputs = append(inputs, fmt.Sprintf("float64(%s)", inputValue))
		case "int[]":
			inputs = append(inputs, fmt.Sprintf("[]int{%s}", inputValue))
		case "float[]":
			inputs = append(inputs, fmt.Sprintf("[]float32{%s}", inputValue))
		case "double[]":
			inputs = append(inputs, fmt.Sprintf("[]float64{%s}", inputValue))
		case "string[]":
			inputs = append(inputs, fmt.Sprintf("[]string{%s}", inputValue))
		case "bool[]":
			inputs = append(inputs, fmt.Sprintf("[]bool{%s}", inputValue))
		}
	}
	return strings.Join(inputs, ",")
}

// Golang template
func RunGoTemplateSourceCode(sourceCode string, functionName string, inputs string) string {
	return fmt.Sprintf(`
package main

func main() {
	%s(%s)
}

// Source Code
%s`, functionName, inputs, sourceCode)
}

func RunGoTemplate(runTemplateInput RunTemplateInput) judge0.Submission {
	inputs := RunGoTemplateInputs(runTemplateInput.TestCases)                                                 // Get the inputs
	sourceCode := RunGoTemplateSourceCode(runTemplateInput.SourceCode, runTemplateInput.FunctionName, inputs) // Get the source code

	submission := judge0.Submission{
		LanguageID:     judge0.LanguageToLanguageID("go"),
		SourceCode:     sourceCode,
		ExpectedOutput: runTemplateInput.ExpectedOutput,
	}

	return submission
}
