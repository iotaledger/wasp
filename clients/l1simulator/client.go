package l1simulator

// this client is required to control the iota l1 simulator
// https://github.com/lmoe/iota-simulator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents the API client
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: time.Second * 30,
		},
	}
}

type GasCostSummary struct {
	ComputationCost         string `json:"computationCost"`
	ComputationCostBurned   string `json:"computationCostBurned"`
	StorageCost             string `json:"storageCost"`
	StorageRebate           string `json:"storageRebate"`
	NonRefundableStorageFee string `json:"nonRefundableStorageFee"`
}

// CheckpointSummary represents the summary part of a checkpoint
type CheckpointSummary struct {
	Epoch                      uint64         `json:"epoch"`
	SequenceNumber             uint64         `json:"sequence_number"`
	NetworkTotalTransactions   uint64         `json:"network_total_transactions"`
	ContentDigest              string         `json:"content_digest"`
	PreviousDigest             *string        `json:"previous_digest"`
	EpochRollingGasCostSummary GasCostSummary `json:"epoch_rolling_gas_cost_summary"`
	TimestampMs                uint64         `json:"timestamp_ms"`
	CheckpointCommitments      []string       `json:"checkpoint_commitments"`
	EndOfEpochData             interface{}    `json:"end_of_epoch_data"`
	VersionSpecificData        []uint8        `json:"version_specific_data"`
}

// AuthorityStrongQuorumSignInfo represents the authority signature information
type AuthorityStrongQuorumSignInfo struct {
	Epoch      uint64  `json:"epoch"`
	Signature  string  `json:"signature"`
	SignersMap []uint8 `json:"signers_map"`
}

// Checkpoint represents the complete checkpoint response
type Checkpoint struct {
	Summary   CheckpointSummary             `json:"summary"`
	Authority AuthorityStrongQuorumSignInfo `json:"authority"`
}

// AdvanceClockRequest represents the request body for advance_clock
type AdvanceClockRequest struct {
	Duration uint32 `json:"duration"`
}

// Health checks the API health
func (c *Client) Health() (string, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/")
	if err != nil {
		return "", fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body failed: %w", err)
	}

	return string(body), nil
}

// CreateCheckpoint creates a new checkpoint
func (c *Client) CreateCheckpoint() (*Checkpoint, error) {
	resp, err := c.httpClient.Post(c.baseURL+"/create_checkpoint", "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("create checkpoint request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var checkpoint Checkpoint
	if err := json.NewDecoder(resp.Body).Decode(&checkpoint); err != nil {
		return nil, fmt.Errorf("decoding response failed: %w", err)
	}

	return &checkpoint, nil
}

// AdvanceClock advances the clock by the specified duration
func (c *Client) AdvanceClock(duration uint32) error {
	payload := AdvanceClockRequest{Duration: duration}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling request failed: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/advance_clock", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("advance clock request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// AdvanceEpoch advances to the next epoch
func (c *Client) AdvanceEpoch() error {
	resp, err := c.httpClient.Post(c.baseURL+"/advance_epoch", "application/json", nil)
	if err != nil {
		return fmt.Errorf("advance epoch request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
