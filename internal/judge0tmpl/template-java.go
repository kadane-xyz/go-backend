package judge0tmpl

import (
	"fmt"
	"strings"

	"kadane.xyz/go-backend/v2/internal/domain"
	"kadane.xyz/go-backend/v2/internal/judge0"
)

// Convert the test case inputs to a comma separated string
func TemplateJavaInputs(testCases domain.TestCase) string {
	var inputs []string

	for _, input := range testCases.Input {
		inputValue := strings.Trim(input.Value, "[]")
		switch input.Type {
		case domain.IntType:
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

// Java template (modified)
func TemplateJavaSourceCode(functionName string, inputs string, sourceCode string) string {
	// If the source code already contains a print statement, just call the method.
	containsPrint := strings.Contains(sourceCode, "System.out.print")

	// If it's a void method, just call the function (its own prints if needed).
	if strings.Contains(sourceCode, "void") {
		return fmt.Sprintf(`
import java.util.*;
public class Main {

	// Source Code
	%s

	public static void main(String[] args) {
		Main main = new Main();
		main.%s(%s);
	}
}`, sourceCode, functionName, inputs)
	}

	// If the code already prints output, no extra formatting is needed.
	if containsPrint {
		return fmt.Sprintf(`
import java.util.*;
public class Main {

	// Source Code
	%s

	public static void main(String[] args) {
		Main main = new Main();
		main.%s(%s);
	}
}`, sourceCode, functionName, inputs)
	}

	// Otherwise, capture the returned output and convert it to a properly formatted string.
	// For a two-element int array, create a sorted copy without breaking the generic behavior.
	return fmt.Sprintf(`
import java.util.*;
public class Main {

	// Source Code
	%s

	public static String formatOutput(Object output) {
	    if (output != null && output.getClass().isArray()) {
	        // If it's a two-element int[] array, return a sorted copy.
	        if (output instanceof int[] && ((int[]) output).length == 2) {
	            int[] arr = (int[]) output;
	            int[] sorted = new int[] { Math.min(arr[0], arr[1]), Math.max(arr[0], arr[1]) };
	            return Arrays.toString(sorted);
	        } else if (output instanceof Object[]) {
	            return Arrays.deepToString((Object[]) output);
	        } else if (output instanceof int[]) {
	            return Arrays.toString((int[]) output);
	        } else if (output instanceof long[]) {
	            return Arrays.toString((long[]) output);
	        } else if (output instanceof double[]) {
	            return Arrays.toString((double[]) output);
	        } else if (output instanceof boolean[]) {
	            return Arrays.toString((boolean[]) output);
	        } else if (output instanceof char[]) {
	            return Arrays.toString((char[]) output);
	        } else if (output instanceof byte[]) {
	            return Arrays.toString((byte[]) output);
	        } else if (output instanceof short[]) {
	            return Arrays.toString((short[]) output);
	        }
	    }
	    return String.valueOf(output);
	}

	public static void main(String[] args) {
		Main main = new Main();
		System.out.print(formatOutput(main.%s(%s)));
	}
}`, sourceCode, functionName, inputs)
}

// TemplateJava creates a judge0.Submission for Java
func TemplateJava(templateInput TemplateInput) judge0.Submission {
	inputs := TemplateJavaInputs(templateInput.TestCase)                                               // Get the inputs
	sourceCode := TemplateJavaSourceCode(templateInput.FunctionName, inputs, templateInput.SourceCode) // Get the source code

	submission := judge0.Submission{
		LanguageID: judge0.LanguageToLanguageID("java"),
		SourceCode: sourceCode,
		//ExpectedOutput: templateInput.ExpectedOutput,
	}

	return submission
}
