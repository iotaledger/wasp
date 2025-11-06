package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"
)

type GraphQLClient struct {
	url    string
	client graphql.Client
}

func NewGraphQLClient(url string) *GraphQLClient {
	return NewGraphQLClientWithTimeout(url, 30*time.Second)
}

func NewGraphQLClientWithTimeout(url string, timeout time.Duration) *GraphQLClient {
	return &GraphQLClient{
		url: strings.TrimRight(url, "/"),
		client: graphql.NewClient(
			url,
			&http.Client{
				Timeout: timeout,
			},
		),
	}
}

// GetGraphQLClient returns the underlying GraphQL client for custom GraphQL queries.
func (c *GraphQLClient) GetGraphQLClient() graphql.Client {
	return c.client
}

// Query builds and executes a custom GraphQL query returning the raw response bytes.
func (c *GraphQLClient) Query(query string, variables map[string]interface{}) ([]byte, error) {
	reqBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	rawBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return rawBytes, nil
}
