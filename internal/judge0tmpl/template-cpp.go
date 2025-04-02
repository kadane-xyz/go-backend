package judge0tmpl

import (
	"fmt"
	"strings"

	"kadane.xyz/go-backend/v2/internal/domain"
	"kadane.xyz/go-backend/v2/internal/judge0"
)

// Convert the test case inputs to a comma separated string
func TemplateCppInputs(testCase domain.TestCase) string {
	var inputs []string

	for _, input := range testCase.Input {
		switch input.Type {
		case "int[]":
			inputValue := strings.Trim(input.Value, "[]")
			inputs = append(inputs, fmt.Sprintf("vector<int>{%s}", inputValue))
		case "int":
			inputs = append(inputs, input.Value)
		case "string":
			inputs = append(inputs, fmt.Sprintf("%q", input.Value))
		case "string[]":
			inputValue := strings.Trim(input.Value, "[]")
			// Split by comma and wrap each element in quotes
			elements := strings.Split(inputValue, ",")
			for i, e := range elements {
				elements[i] = fmt.Sprintf("%q", strings.TrimSpace(e))
			}
			inputs = append(inputs, fmt.Sprintf("vector<string>{%s}", strings.Join(elements, ",")))
		default:
			inputs = append(inputs, input.Value)
		}
	}

	return strings.Join(inputs, ", ")
}

// C++ template
func TemplateCppSourceCode(functionName string, inputs string, sourceCode string) string {
	return fmt.Sprintf(`
#include <iostream>
#include <string>
#include <vector>
#include <algorithm>
#include <map>
#include <unordered_map>
#include <set>
#include <unordered_set>
#include <queue>
#include <deque>
#include <stack>
#include <cmath>
#include <cstdio>
#include <cstdlib>
#include <cstring>
using namespace std;

// Helper functions to print different types
template<typename T>
void printValue(const T& val) {
	cout << val;
}

template<typename T>
void printVector(const vector<T>& vec) {
	cout << "[";
	for (size_t i = 0; i < vec.size(); i++) {
		if (i > 0) cout << ",";
		printValue(vec[i]);
	}
	cout << "]";
}

template<typename T>
void printResult(const T& result) {
	printValue(result);
}

template<typename T>
void printResult(const vector<T>& result) {
	printVector(result);
}

// Source Code
%s

int main() {
	auto result = %s(%s);
	printResult(result);
	return 0;
}
`, sourceCode, functionName, inputs)
}

func TemplateCpp(templateInput TemplateInput) judge0.Submission {
	inputs := TemplateCppInputs(templateInput.TestCase)                                               // Get the inputs
	sourceCode := TemplateCppSourceCode(templateInput.FunctionName, inputs, templateInput.SourceCode) // Get the source code
	languageID := judge0.LanguageToLanguageID("cpp")

	submission := judge0.Submission{
		LanguageID: languageID,
		SourceCode: sourceCode,
		//ExpectedOutput: templateInput.ExpectedOutput,
	}

	return submission
}
