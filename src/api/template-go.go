package api

import (
	"fmt"
	"strings"

	"kadane.xyz/go-backend/v2/src/judge0"
)

// Convert the test case inputs to a comma separated string
func TemplateGoInputs(testCases TestCase) string {
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
func TemplateGoSourceCode(functionName string, inputs string, sourceCode string) string {
	return fmt.Sprintf(`
package main

// Source Code
%s

func main() {
	%s(%s)
}`, sourceCode, functionName, inputs)
}

func TemplateGo(templateInput TemplateInput) judge0.Submission {
	inputs := TemplateGoInputs(templateInput.TestCase)                                               // Get the inputs
	sourceCode := TemplateGoSourceCode(templateInput.FunctionName, inputs, templateInput.SourceCode) // Get the source code

	submission := judge0.Submission{
		LanguageID: judge0.LanguageToLanguageID("go"),
		SourceCode: sourceCode,
		//ExpectedOutput: templateInput.ExpectedOutput,
	}

	return submission
}
