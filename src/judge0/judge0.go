package judge0

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"kadane.xyz/go-backend/v2/src/config"
)

type Judge0Client struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

type Submission struct {
	SourceCode           string `json:"source_code"` // plain string that will be base64 encoded
	LanguageID           int    `json:"language_id"`
	CompilerOptions      string `json:"compiler_options"`
	CommandLineArguments string `json:"command_line_arguments"`
	Stdin                string `json:"stdin"` // plain string that will be base64 encoded
	//ExpectedOutput       string `json:"expected_output"` // plain string that will be base64 encoded
}

type SubmissionBatch struct {
	Submissions []Submission `json:"submissions"`
}

type SubmissionResponse struct {
	Token string `json:"token"`
}

type SubmissionResult struct {
	Stdout        string `json:"stdout"`
	Stderr        string `json:"stderr"`
	CompileOutput string `json:"compile_output"`
	Message       string `json:"message"`
	ExitCode      int    `json:"exit_code"`
	ExitSignal    int    `json:"exit_signal"`
	Status        struct {
		ID          int    `json:"id"`
		Description string `json:"description"`
	} `json:"status"`
	CreatedAt  string `json:"created_at"`
	FinishedAt string `json:"finished_at"`
	Token      string `json:"token"`
	Time       string `json:"time"`
	WallTime   string `json:"wall_time"`
	Memory     int    `json:"memory"`
	Language   struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"language"`
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

// default values for retry and wait times
const (
	initialRetryDelay = 50 * time.Millisecond
	maxRetryDelay     = 500 * time.Millisecond
	maxWaitTime       = 30 * time.Second
)

func NewJudge0Client(cfg *config.Config) *Judge0Client {
	return &Judge0Client{
		BaseURL: cfg.Judge0Url,
		Token:   cfg.Judge0Token,
		Client:  &http.Client{},
	}
}

func (c *Judge0Client) CreateSubmissionBatchAndWait(submissions []Submission) ([]SubmissionResult, error) {
	var wg sync.WaitGroup
	results := make([]SubmissionResult, len(submissions))
	errors := make(chan error, len(submissions))

	for i, submission := range submissions {
		wg.Add(1)
		go func(i int, submission Submission) {
			defer wg.Done()
			resp, err := c.CreateSubmissionAndWait(submission)
			if err != nil {
				errors <- fmt.Errorf("submission %d error: %s", i, err.Error())
				return
			}
			results[i] = *resp
		}(i, submission)
	}

	wg.Wait()
	close(errors)

	if len(errors) > 0 {
		return nil, fmt.Errorf("one or more submissions failed")
	}

	return results, nil
}

func (c *Judge0Client) CreateSubmissionAndWait(submission Submission) (*SubmissionResult, error) {
	// First base64 encode submission
	submission = EncodeSubmissionInputs(submission)
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
	url := fmt.Sprintf("%s/submissions/batch?base64_encoded=true", c.BaseURL)

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
	// return as base64 encoded string with fields *
	url := fmt.Sprintf("%s/submissions/%s?base64_encoded=true&fields=*", c.BaseURL, token)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result SubmissionResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	resultDecoded := result
	resultDecoded.Stdout, err = DecodeBase64(result.Stdout)
	if err != nil {
		return nil, fmt.Errorf("error decoding stdout: %w", err)
	}
	resultDecoded.Stderr, err = DecodeBase64(result.Stderr)
	if err != nil {
		return nil, fmt.Errorf("error decoding stderr: %w", err)
	}
	resultDecoded.CompileOutput, err = DecodeBase64(result.CompileOutput)
	if err != nil {
		return nil, fmt.Errorf("error decoding compile output: %w", err)
	}
	resultDecoded.Message, err = DecodeBase64(result.Message)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %w", err)
	}

	return &resultDecoded, nil
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
