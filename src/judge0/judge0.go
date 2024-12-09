package judge0

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

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
	Stdin          []byte `json:"stdin,omitempty"`
	ExpectedOutput []byte `json:"expected_output,omitempty"`
	Wait           bool   `json:"wait,omitempty"`
}

type SubmissionBatch struct {
	Submissions []Submission `json:"submissions"`
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
	Language      struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"language"`
	Message string `json:"message"`
	Status  struct {
		ID          int    `json:"id"`
		Description string `json:"description"`
	} `json:"status"`
}

type PaginationMeta struct {
	CurrentPage int `json:"current_page"`
	NextPage    int `json:"next_page"`
	PrevPage    int `json:"prev_page"`
	TotalPages  int `json:"total_pages"`
	TotalCount  int `json:"total_count"`
}

type SubmissionsResponse struct {
	Submissions []SubmissionResult `json:"submissions"`
	Meta        PaginationMeta     `json:"meta"`
}

type SubmissionBatchResponse struct {
	Submissions []SubmissionResponse
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

const (
	initialRetryDelay = 50 * time.Millisecond
	maxRetryDelay     = 500 * time.Millisecond
	maxWaitTime       = 30 * time.Second
)

func (c *Judge0Client) CreateSubmissionBatchAndWait(submissions []Submission) ([]SubmissionResult, error) {
	// First create the submission without waiting
	resp, err := c.CreateSubmissionBatch(submissions)
	if err != nil {
		return nil, fmt.Errorf("batch creation error: %w", err)
	}

	// Create a map to track processed submissions
	results := make([]SubmissionResult, 0, len(submissions))
	pendingTokens := make(map[string]bool)

	// Initialize pending tokens
	for _, submission := range resp.Submissions {
		pendingTokens[submission.Token] = true
	}

	// Quick first check after submission
	for token := range pendingTokens {
		result, err := c.GetSubmission(token)
		if err == nil && result.Status.ID >= 3 {
			results = append(results, *result)
			delete(pendingTokens, token)
		}
	}

	// Then poll remaining submissions with exponential backoff
	startTime := time.Now()
	currentDelay := initialRetryDelay

	for len(pendingTokens) > 0 {
		if time.Since(startTime) > maxWaitTime {
			return nil, fmt.Errorf("submission batch timed out after %v", maxWaitTime)
		}

		time.Sleep(currentDelay)

		// Check each pending submission
		for token := range pendingTokens {
			result, err := c.GetSubmission(token)
			if err != nil {
				continue // Skip this token for now if there's an error
			}

			if result.Status.ID >= 3 {
				results = append(results, *result)
				delete(pendingTokens, token)
			}
		}

		// Adjust delay with gentler backoff
		currentDelay = time.Duration(float64(currentDelay) * 1.5)
		if currentDelay > maxRetryDelay {
			currentDelay = maxRetryDelay
		}
	}

	return results, nil
}

func (c *Judge0Client) CreateSubmissionAndWait(submission Submission) (*SubmissionResult, error) {
	// First create the submission without waiting
	resp, err := c.CreateSubmission(submission)
	if err != nil {
		return nil, err
	}

	// Quick first check after submission
	result, err := c.GetSubmission(resp.Token)
	if err == nil && result.Status.ID >= 3 {
		return result, nil
	}

	// Then poll with exponential backoff
	startTime := time.Now()
	currentDelay := initialRetryDelay

	for {
		if time.Since(startTime) > maxWaitTime {
			return nil, fmt.Errorf("submission timed out after %v", maxWaitTime)
		}

		time.Sleep(currentDelay)
		result, err := c.GetSubmission(resp.Token)
		if err != nil {
			return nil, err
		}

		if result.Status.ID >= 3 {
			return result, nil
		}

		// Smaller multiplication factor for gentler backoff
		currentDelay = time.Duration(float64(currentDelay) * 1.5)
		if currentDelay > maxRetryDelay {
			currentDelay = maxRetryDelay
		}
	}
}

func (c *Judge0Client) CreateSubmission(submission Submission) (*SubmissionResponse, error) {
	url := fmt.Sprintf("%s/submissions?base64_encoded=true", c.BaseURL)

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

func (c *Judge0Client) CreateSubmissionBatch(submissions []Submission) (*SubmissionBatchResponse, error) {
	url := fmt.Sprintf("%s/submissions/batch?base64_encoded=true&fields=*", c.BaseURL)

	submissionBatch := SubmissionBatch{
		Submissions: submissions,
	}

	jsonData, err := json.Marshal(submissionBatch)
	if err != nil {
		return nil, fmt.Errorf("error marshaling submission batch: %w", err)
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

	var submissionResponses []SubmissionResponse
	if err := json.Unmarshal(body, &submissionResponses); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &SubmissionBatchResponse{
		Submissions: submissionResponses,
	}, nil
}

func (c *Judge0Client) GetSubmission(token string) (*SubmissionResult, error) {
	url := fmt.Sprintf("%s/submissions/%s?base64_encoded=false&fields=stdout,time,memory,stderr,token,compile_output,message,status,language", c.BaseURL, token)

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

func (c *Judge0Client) GetSubmissions() (*SubmissionsResponse, error) {
	url := fmt.Sprintf("%s/submissions?base64_encoded=false&fields=*", c.BaseURL)

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

	log.Println(body)

	var response SubmissionsResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
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
