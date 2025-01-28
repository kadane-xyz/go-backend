package api

import (
	"fmt"
	"strings"

	"kadane.xyz/go-backend/v2/src/judge0"
)

// Convert the test case inputs to a comma separated string
func TemplateJavaInputs(testCases TestCase) string {
	var inputs []string

	for _, input := range testCases.Input {
		inputValue := strings.Trim(input.Value, "[]")
		switch input.Type {
		case IntType:
			inputs = append(inputs, inputValue)
		case "float":
			inputs = append(inputs, fmt.Sprintf("%sf", inputValue)) // Java float literal
		case "double":
			inputs = append(inputs, fmt.Sprintf("%sd", inputValue)) // Java double literal
		case "int[]":
			inputs = append(inputs, fmt.Sprintf("new int[]{%s}", inputValue))
		case "float[]":
			inputs = append(inputs, fmt.Sprintf("new float[]{%s}", inputValue))
		case "double[]":
			inputs = append(inputs, fmt.Sprintf("new double[]{%s}", inputValue))
		case "string[]":
			inputs = append(inputs, fmt.Sprintf("new String[]{%s}", inputValue))
		case "boolean[]":
			inputs = append(inputs, fmt.Sprintf("new boolean[]{%s}", inputValue))
		}
	}
	return strings.Join(inputs, ",")
}

// Java template
func TemplateJavaSourceCode(functionName string, inputs string, sourceCode string) string {
	// For void functions
	if strings.Contains(sourceCode, "void") {
		return fmt.Sprintf(`
public class Main {

	// Source Code
	%s

	public static void main(String[] args) {
		Main main = new Main();
		main.%s(%s);
	}
}`, sourceCode, functionName, inputs)
	}

	// For non-void functions
	return fmt.Sprintf(`
public class Main {

	// Source Code
	%s

	public static void main(String[] args) {
		Main main = new Main();
		System.out.print(main.%s(%s));
	}
}`, sourceCode, functionName, inputs)
}

func TemplateJava(templateInput TemplateInput) judge0.Submission {
	inputs := TemplateJavaInputs(templateInput.TestCases)                                              // Get the inputs
	sourceCode := TemplateJavaSourceCode(templateInput.FunctionName, inputs, templateInput.SourceCode) // Get the source code

	submission := judge0.Submission{
		LanguageID: judge0.LanguageToLanguageID("java"),
		SourceCode: sourceCode,
		//ExpectedOutput: templateInput.ExpectedOutput,
	}

	return submission
}
