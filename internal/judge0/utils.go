package judge0

import "encoding/base64"

func LanguageToLanguageID(language string) int {
	return languageIDMap[language]
}

func LanguageIDToLanguage(languageID int) string {
	for language, id := range languageIDMap {
		if id == languageID {
			return language
		}
	}
	return ""
}

func EncodeBase64(text string) string {
	return base64.StdEncoding.EncodeToString([]byte(text))
}

func DecodeBase64(encodedText string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encodedText)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func EncodeSubmissionInputs(submission Submission) Submission {
	submission.SourceCode = EncodeBase64(submission.SourceCode)
	submission.Stdin = EncodeBase64(submission.Stdin)
	//submission.ExpectedOutput = EncodeBase64(submission.ExpectedOutput)
	return submission
}

func EncodeInputSubmissionsInput(submissions []Submission) []Submission {
	for i, submission := range submissions {
		submission.SourceCode = EncodeBase64(submission.SourceCode)
		submission.Stdin = EncodeBase64(submission.Stdin)
		//submission.ExpectedOutput = EncodeBase64(submission.ExpectedOutput)
		submissions[i] = submission
	}
	return submissions
}
