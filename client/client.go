package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type WaspClient struct {
	httpClient http.Client
	baseURL    string
}

const AdminRoutePrefix = "adm"

// NewWaspClient returns a new *WaspClient with the given baseURL and httpClient.
func NewWaspClient(baseURL string, httpClient ...http.Client) *WaspClient {
	if !strings.Contains(baseURL, "://") {
		baseURL = "http://" + baseURL
	}
	if len(httpClient) > 0 {
		return &WaspClient{baseURL: baseURL, httpClient: httpClient[0]}
	}
	return &WaspClient{baseURL: baseURL}
}

type ErrorResponse struct {
	Message string `json:"message"`
	code    int
}

func NewErrorResponse(code int, message string) *ErrorResponse {
	return &ErrorResponse{Message: message, code: code}
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("%d: %s", e.code, e.Message)
}

func (e *ErrorResponse) Code() int {
	return e.code
}

func IsNotFound(e error) bool {
	er, ok := e.(*ErrorResponse)
	return ok && er.Code() == http.StatusNotFound
}

func processResponse(res *http.Response, decodeTo interface{}) error {
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated {
		if decodeTo != nil {
			return json.Unmarshal(resBody, decodeTo)
		} else {
			return nil
		}
	}

	errRes := &ErrorResponse{
		code: res.StatusCode,
	}
	if err := json.Unmarshal(resBody, errRes); err != nil {
		errRes.Message = http.StatusText(res.StatusCode)
	}
	errRes.code = res.StatusCode
	return errRes
}

func (c *WaspClient) do(method string, route string, reqObj interface{}, resObj interface{}) error {
	// marshal request object
	var data []byte
	if reqObj != nil {
		var err error
		data, err = json.Marshal(reqObj)
		if err != nil {
			return err
		}
	}

	// construct request
	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s", c.baseURL, route), func() io.Reader {
		if data == nil {
			return nil
		}
		return bytes.NewReader(data)
	}())
	if err != nil {
		return err
	}

	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// make the request
	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Request failed: %v", err)
	}

	// write response into response object
	return processResponse(res, resObj)
}

// BaseURL returns the baseURL of the client.
func (c *WaspClient) BaseURL() string {
	return c.baseURL
}
