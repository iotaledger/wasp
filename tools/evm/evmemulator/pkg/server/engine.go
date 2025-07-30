package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// --- JSON-RPC types (as above) ---

type PayloadStatus struct {
	Status          string  `json:"status"`
	LatestValidHash string  `json:"latestValidHash"`
	ValidationError *string `json:"validationError"` // pointer is important here otherwise it will be omitted and cause a validation error.
}

type ForkchoiceResult struct {
	PayloadStatus PayloadStatus `json:"payloadStatus"`
	PayloadID     string        `json:"payloadId"`
}

type JsonRPCResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  any         `json:"result"`
}

type ErrorResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type RpcRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      interface{}   `json:"id"`
}

func logPretty(prefix string, v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("%s (marshal error: %v)\n", prefix, err)
		return
	}
	fmt.Printf("%s\n%s\n", prefix, data)
}

func engineMockHandler(w http.ResponseWriter, r *http.Request) {
	var req RpcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	// Log input request
	logPretty("Input Request:", req)

	w.Header().Set("Content-Type", "application/json")

	var respBytes []byte
	var resp interface{}

	switch req.Method {
	case "engine_newPayloadV3", "engine_newPayloadV2", "engine_newPayloadV1":
		resp = JsonRPCResponse{
			Jsonrpc: "2.0",
			ID:      req.ID,
			Result: PayloadStatus{
				Status:          "VALID",
				LatestValidHash: "0x0000000000000000000000000000000000000000000000000000000000000000",
				ValidationError: nil,
			},
		}
	case "engine_forkchoiceUpdatedV3", "engine_forkchoiceUpdatedV2", "engine_forkchoiceUpdatedV1":
		resp = JsonRPCResponse{
			Jsonrpc: "2.0",
			ID:      req.ID,
			Result: ForkchoiceResult{
				PayloadStatus: PayloadStatus{
					Status:          "VALID",
					LatestValidHash: "0x0000000000000000000000000000000000000000000000000000000000000000",
					ValidationError: nil,
				},
				PayloadID: "0x0000000021f32cc1",
			},
		}
	default:
		var errResp ErrorResponse
		errResp.Jsonrpc = "2.0"
		errResp.ID = req.ID
		errResp.Error.Code = -32601
		errResp.Error.Message = "Method not implemented"
		resp = errResp
	}

	// Marshal response for logging and sending
	respBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		http.Error(w, "internal marshal error", http.StatusInternalServerError)
		return
	}
	// Log output response
	fmt.Printf("Output Response:\n%s\n", respBytes)

	// Send response (compact JSON)
	var compactBuf bytes.Buffer
	if err := json.Compact(&compactBuf, respBytes); err == nil {
		w.Write(compactBuf.Bytes())
	} else {
		w.Write(respBytes)
	}
}

func StartEngineMockServer(laddr string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", engineMockHandler)
	return &http.Server{
		Addr:    laddr,
		Handler: mux,
	}
}
