package drafting

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/example/job-application-assistant/apps/api/internal/domain"
)

type Result struct {
	CoverLetter  string `json:"coverLetter"`
	MatchSummary string `json:"matchSummary"`
}

type Generator interface {
	Generate(ctx context.Context, posting domain.JobPosting, profile domain.Profile) (Result, error)
}

type TokenProvider func(context.Context) (string, error)

type Foundry struct {
	endpoint   string
	deployment string
	apiKey     string
	token      TokenProvider
	client     *http.Client
}

func NewFoundry(endpoint, deployment, apiKey string, token TokenProvider, client *http.Client) (*Foundry, error) {
	if strings.TrimSpace(endpoint) == "" || strings.TrimSpace(deployment) == "" {
		return nil, errors.New("Azure AI endpoint and deployment are required")
	}
	if apiKey == "" && token == nil {
		return nil, errors.New("Azure AI authentication is required")
	}
	if client == nil {
		client = http.DefaultClient
	}
	endpoint = strings.TrimSuffix(endpoint, "/")
	if !strings.HasSuffix(endpoint, "/openai/v1") {
		endpoint += "/openai/v1"
	}
	return &Foundry{endpoint: endpoint, deployment: deployment, apiKey: apiKey, token: token, client: client}, nil
}

func (f *Foundry) Generate(ctx context.Context, posting domain.JobPosting, profile domain.Profile) (Result, error) {
	prompt := fmt.Sprintf(`Create a tailored cover letter and a concise fit summary.

Rules:
- Use only facts found in the candidate profile. Never invent experience, metrics, skills, or relationships.
- Write a polished cover letter of 250-400 words in plain text.
- Do not include an address block or placeholders.
- Return only valid JSON with exactly two string properties: "coverLetter" and "matchSummary".

Candidate profile:
%s

Additional background:
%s

Job title: %s
Company: %s
Job description:
%s`, profile.ResumeText, profile.Background, posting.Title, posting.Company, posting.Description)

	payload := map[string]any{
		"model": f.deployment,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a careful job-application writing assistant."},
			{"role": "user", "content": prompt},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return Result{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, f.endpoint+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	if f.apiKey != "" {
		request.Header.Set("api-key", f.apiKey)
	} else {
		token, tokenErr := f.token(ctx)
		if tokenErr != nil {
			return Result{}, fmt.Errorf("get Azure AI token: %w", tokenErr)
		}
		request.Header.Set("Authorization", "Bearer "+token)
	}

	response, err := f.client.Do(request)
	if err != nil {
		return Result{}, fmt.Errorf("call Azure AI Foundry: %w", err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		return Result{}, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Result{}, fmt.Errorf("Azure AI Foundry returned HTTP %d: %s", response.StatusCode, safeError(responseBody))
	}
	var completion struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(responseBody, &completion); err != nil {
		return Result{}, fmt.Errorf("decode Azure AI response: %w", err)
	}
	if len(completion.Choices) == 0 {
		return Result{}, errors.New("Azure AI returned no choices")
	}
	content := strings.TrimSpace(completion.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	var result Result
	if err := json.Unmarshal([]byte(strings.TrimSpace(content)), &result); err != nil {
		return Result{}, fmt.Errorf("decode generated draft: %w", err)
	}
	if strings.TrimSpace(result.CoverLetter) == "" || strings.TrimSpace(result.MatchSummary) == "" {
		return Result{}, errors.New("Azure AI returned an incomplete draft")
	}
	return result, nil
}

func safeError(body []byte) string {
	value := strings.TrimSpace(string(body))
	if len(value) > 500 {
		value = value[:500]
	}
	return value
}
