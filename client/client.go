package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/iotaledger/wasp/packages/webapi/model"
	"golang.org/x/xerrors"
)

// WaspClient allows to make requests to the Wasp web API.
type WaspClient struct {
	httpClient http.Client
	baseURL    string
	token      string
}

// NewWaspClient returns a new *WaspClient with the given baseURL and httpClient.
func NewWaspClient(baseURL string, httpClient ...http.Client) *WaspClient {
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	if len(httpClient) > 0 {
		return &WaspClient{baseURL: baseURL, httpClient: httpClient[0]}
	}
	return &WaspClient{baseURL: baseURL}
}

func (c *WaspClient) WithToken(token string) *WaspClient {
	c.token = token

	return c
}

func processResponse(res *http.Response, decodeTo interface{}) error {
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return xerrors.Errorf("unable to read response body: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		if decodeTo != nil {
			return json.Unmarshal(resBody, decodeTo)
		}
		return nil
	}

	errRes := &model.HTTPError{}
	if err := json.Unmarshal(resBody, errRes); err != nil {
		errRes.Message = http.StatusText(res.StatusCode)
	}
	errRes.StatusCode = res.StatusCode
	errRes.Message = string(resBody)
	return errRes
}

func (c *WaspClient) do(method, route string, reqObj, resObj interface{}) error {
	// marshal request object
	var data []byte
	if reqObj != nil {
		var err error
		data, err = json.Marshal(reqObj)
		if err != nil {
			return xerrors.Errorf("json.Marshal: %w", err)
		}
	}

	// construct request
	url := fmt.Sprintf("%s/%s", strings.TrimRight(c.baseURL, "/"), strings.TrimLeft(route, "/"))
	req, err := http.NewRequestWithContext(context.Background(), method, url, func() io.Reader {
		if data == nil {
			return nil
		}
		return bytes.NewReader(data)
	}())
	if err != nil {
		return xerrors.Errorf("http.NewRequest [%s %s]: %w", method, url, err)
	}

	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", c.token))
	}

	// make the request
	res, err := c.httpClient.Do(req)
	if err != nil {
		return xerrors.Errorf("%s %s: %w", method, url, err)
	}

	// write response into response object
	err = processResponse(res, resObj)
	if err != nil {
		return xerrors.Errorf("%s %s: %w", method, url, err)
	}
	return nil
}

// BaseURL returns the baseURL of the client.
func (c *WaspClient) BaseURL() string {
	return c.baseURL
}
