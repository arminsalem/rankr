package project

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
)

type PublicProject struct {
	RepositoryID string `json:"repository_id"`
	Name         string `json:"name"`
	Owner        string `json:"owner"`
	Provider     string `json:"provider"`
	Status       string `json:"status"`
}

type HTTPClient struct {
	baseURL *url.URL
	client  *http.Client
}

func NewHTTPClient(baseURL string, timeout time.Duration) (*HTTPClient, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("project: parse base URL: %w", err)
	}
	if (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" {
		return nil, fmt.Errorf("project: base URL must be an absolute HTTP or HTTPS URL")
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("project: timeout must be positive")
	}
	return &HTTPClient{baseURL: base, client: &http.Client{Timeout: timeout}}, nil
}

func (c *HTTPClient) GetPublicProject(ctx context.Context, id uuid.UUID) (PublicProject, bool, error) {
	endpoint := c.baseURL.JoinPath("v1", "projects", "public", id.String())
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return PublicProject{}, false, fmt.Errorf("project: create request: %w", err)
	}
	request.Header.Set("Accept", "application/json")

	response, err := c.client.Do(request)
	if err != nil {
		return PublicProject{}, false, fmt.Errorf("project: request public project: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return PublicProject{}, false, nil
	}
	if response.StatusCode != http.StatusOK {
		return PublicProject{}, false, fmt.Errorf("project: unexpected HTTP status: %d", response.StatusCode)
	}

	var result PublicProject
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&result); err != nil {
		return PublicProject{}, false, fmt.Errorf("project: decode public project: %w", err)
	}
	return result, true, nil
}
