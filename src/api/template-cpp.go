package api

import (
	"fmt"
	"strings"

	"kadane.xyz/go-backend/v2/src/judge0"
)

// Convert the test case inputs to a comma separated string
func TemplateCppInputs(testCases TestCase) string {
	var inputs []string

	return strings.Join(inputs, ",")
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

// Source Code
%s

int main() {
	%s(%s);

	return 0;
}
`, sourceCode, functionName, inputs)
}

func TemplateCpp(templateInput TemplateInput) judge0.Submission {
	inputs := TemplateCppInputs(templateInput.TestCases)                                              // Get the inputs
	sourceCode := TemplateCppSourceCode(templateInput.FunctionName, inputs, templateInput.SourceCode) // Get the source code
	languageID := judge0.LanguageToLanguageID("cpp")

	submission := judge0.Submission{
		LanguageID: languageID,
		SourceCode: sourceCode,
		//ExpectedOutput: templateInput.ExpectedOutput,
	}

	return submission
}
