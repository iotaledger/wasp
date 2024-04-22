package miniclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/iotaledger/wasp/packages/testutil/privtangle_sui/miniclient/types"
)

type MiniClient struct {
	Host string
}

func NewMiniClient(host string) *MiniClient {
	return &MiniClient{
		Host: host,
	}
}

type JsonRPCRequestBody struct {
	JsonRPC string `json:"jsonrpc"`
	ID      uint32 `json:"id"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

func NewJsonRPCRequest(method string, params ...any) *JsonRPCRequestBody {
	return &JsonRPCRequestBody{
		JsonRPC: "2.0",
		ID:      1,
		Method:  method,
		Params:  params,
	}
}

func (j *JsonRPCRequestBody) JSON() ([]byte, error) {
	return json.Marshal(j)
}

func (j *JsonRPCRequestBody) BodyReader() *bytes.Reader {
	req, err := j.JSON()
	if err != nil {
		panic(err)
	}

	return bytes.NewReader(req)
}

func (c *MiniClient) doClientRequest(ctx context.Context, body *JsonRPCRequestBody, timeout time.Duration) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Host, body.BodyReader())
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{
		Timeout: timeout,
	}

	return client.Do(req)
}

func (c *MiniClient) GetLatestSuiSystemState(ctx context.Context) (*types.SuiX_GetLatestSuiSystemState, error) {
	body := NewJsonRPCRequest("suix_getLatestSuiSystemState")

	res, err := c.doClientRequest(ctx, body, 10*time.Second)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	result := types.SuiX_GetLatestSuiSystemState{}
	err = json.Unmarshal(resBody, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
