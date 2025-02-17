package api

import (
	"fmt"
	"strings"

	"kadane.xyz/go-backend/v2/src/judge0"
)

// Convert the test case inputs to a comma separated string
func TemplatePythonInputs(testCases TestCase) string {
	var inputs []string

	for _, input := range testCases.Input {
		inputValue := strings.Trim(input.Value, "[]")
		switch input.Type {
		case "string":
			// Python strings use single or double quotes
			inputs = append(inputs, fmt.Sprintf("'%s'", inputValue))
		case "boolean", "bool":
			// Python uses True/False with capital first letter
			if inputValue == "true" {
				inputs = append(inputs, "True")
			} else {
				inputs = append(inputs, "False")
			}
		case "float", "double":
			// Ensure float values have decimal point
			if !strings.Contains(inputValue, ".") {
				inputs = append(inputs, inputValue+".0")
			} else {
				inputs = append(inputs, inputValue)
			}
		case "int[]":
			inputs = append(inputs, fmt.Sprintf("[%s]", inputValue))
		case "float[]", "double[]":
			// Convert each number to float
			nums := strings.Split(inputValue, ",")
			floats := make([]string, len(nums))
			for i, num := range nums {
				if !strings.Contains(num, ".") {
					floats[i] = num + ".0"
				} else {
					floats[i] = num
				}
			}
			inputs = append(inputs, fmt.Sprintf("[%s]", strings.Join(floats, ",")))
		case "string[]":
			// Add quotes around each string in array
			strs := strings.Split(inputValue, ",")
			for i, s := range strs {
				strs[i] = fmt.Sprintf("'%s'", strings.Trim(s, "\"'"))
			}
			inputs = append(inputs, fmt.Sprintf("[%s]", strings.Join(strs, ",")))
		case "boolean[]", "bool[]":
			// Convert each boolean to Python syntax
			bools := strings.Split(inputValue, ",")
			for i, b := range bools {
				if b == "true" {
					bools[i] = "True"
				} else {
					bools[i] = "False"
				}
			}
			inputs = append(inputs, fmt.Sprintf("[%s]", strings.Join(bools, ",")))
		default:
			// Integers can pass through as-is
			inputs = append(inputs, inputValue)
		}
	}
	return strings.Join(inputs, ",")
}

// Python template
func TemplatePythonSourceCode(functionName string, inputs string, sourceCode string) string {
	return fmt.Sprintf(`
%s

%s(%s)
`, sourceCode, functionName, inputs)
}

func TemplatePython(templateInput TemplateInput) judge0.Submission {
	inputs := TemplatePythonInputs(templateInput.TestCase)                                               // Get the inputs
	sourceCode := TemplatePythonSourceCode(templateInput.FunctionName, inputs, templateInput.SourceCode) // Get the source code

	submission := judge0.Submission{
		LanguageID: judge0.LanguageToLanguageID("python"),
		SourceCode: sourceCode,
		//ExpectedOutput: templateInput.ExpectedOutput,
	}

	return submission
}
