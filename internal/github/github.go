package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	apiBaseURL = "https://api.github.com"

	// TokenEnvVar is the environment variable holding a PAT with `Actions: read` (and `repo` for private targets)
	TokenEnvVar = "GITHUB_TOKEN"
)

var ErrTokenNotSet = errors.New(TokenEnvVar + " not set")

// WorkflowRun is the subset of a GitHub Actions workflow run used by the dashboard
type WorkflowRun struct {
	Status       string    `json:"status"`
	Conclusion   string    `json:"conclusion"`
	RunStartedAt time.Time `json:"run_started_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	HTMLURL      string    `json:"html_url"`
}

type Client struct {
	token   string
	baseURL string
	http    *http.Client
}

// NewClientFromEnv creates a Client authenticated with the GITHUB_TOKEN environment variable
func NewClientFromEnv() (*Client, error) {
	token := os.Getenv(TokenEnvVar)
	if token == "" {
		return nil, ErrTokenNotSet
	}
	return &Client{token: token, baseURL: apiBaseURL, http: &http.Client{Timeout: 10 * time.Second}}, nil
}

func (c *Client) get(ctx context.Context, path string, params url.Values, out any) error {
	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("GitHub API %d on %s: %s", resp.StatusCode, path, body)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// GetDefaultBranch returns the default branch of the repo (owner/name)
func (c *Client) GetDefaultBranch(ctx context.Context, repo string) (string, error) {
	var data struct {
		DefaultBranch string `json:"default_branch"`
	}
	if err := c.get(ctx, "/repos/"+repo, nil, &data); err != nil {
		return "", err
	}
	return data.DefaultBranch, nil
}

// GetLatestRun returns the most recent run of the workflow on the branch, or nil if it has never run
func (c *Client) GetLatestRun(ctx context.Context, repo, workflow, branch string) (*WorkflowRun, error) {
	var data struct {
		WorkflowRuns []WorkflowRun `json:"workflow_runs"`
	}
	params := url.Values{"branch": {branch}, "per_page": {"1"}}
	if err := c.get(ctx, fmt.Sprintf("/repos/%s/actions/workflows/%s/runs", repo, url.PathEscape(workflow)), params, &data); err != nil {
		return nil, err
	}
	if len(data.WorkflowRuns) == 0 {
		return nil, nil
	}
	return &data.WorkflowRuns[0], nil
}
