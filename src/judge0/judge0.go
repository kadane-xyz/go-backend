package judge0

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"kadane.xyz/go-backend/v2/src/config"
)

type Judge0Client struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

type Submission struct {
	SourceCode     string `json:"source_code"`
	LanguageID     int    `json:"language_id"`
	Stdin          string `json:"stdin,omitempty"`
	ExpectedOutput string `json:"expected_output,omitempty"`
	Wait           bool   `json:"wait,omitempty"`
}

type SubmissionResponse struct {
	Token string `json:"token"`
}

type SubmissionResult struct {
	Stdout        string `json:"stdout"`
	Time          string `json:"time"`
	Memory        int    `json:"memory"`
	Stderr        string `json:"stderr"`
	Token         string `json:"token"`
	CompileOutput string `json:"compile_output"`
	Message       string `json:"message"`
	Status        struct {
		ID          int    `json:"id"`
		Description string `json:"description"`
	} `json:"status"`
}

var languageIDMap = map[string]int{
	"cpp":        54,
	"go":         60,
	"java":       62,
	"javascript": 63,
	"python":     71,
	"typescript": 74,
}

func LanguageToLanguageID(language string) int {
	return languageIDMap[language]
}

func NewJudge0Client(cfg *config.Config) *Judge0Client {
	return &Judge0Client{
		BaseURL: cfg.Judge0Url,
		Token:   cfg.Judge0Token,
		Client:  &http.Client{},
	}
}

func (c *Judge0Client) CreateSubmission(submission Submission) (*SubmissionResponse, error) {
	url := fmt.Sprintf("%s/submissions?base64_encoded=true&wait=%v", c.BaseURL, submission.Wait)

	jsonData, err := json.Marshal(submission)
	if err != nil {
		return nil, fmt.Errorf("error marshaling submission: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Token", c.Token)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var submissionResp SubmissionResponse
	err = json.Unmarshal(body, &submissionResp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &submissionResp, nil
}

func (c *Judge0Client) GetSubmission(token string) (*SubmissionResult, error) {
	url := fmt.Sprintf("%s/submissions/%s?base64_encoded=false", c.BaseURL, token)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	req.Header.Set("X-Auth-Token", c.Token)

	resp, err := c.Client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result SubmissionResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &result, nil
}

func (c *Judge0Client) GetLanguages() ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/languages", c.BaseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Auth-Token", c.Token)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var languages []map[string]interface{}
	err = json.Unmarshal(body, &languages)
	if err != nil {
		return nil, err
	}

	return languages, nil
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
