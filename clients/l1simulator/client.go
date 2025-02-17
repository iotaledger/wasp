package l1simulator

// this client is required to control the iota l1 simulator
// https://github.com/lmoe/iota-simulator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

// Client represents the API client
type Client struct {
	clients.L1Client
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new API client
func NewClient(l1config clients.L1Config, controlApiUrl string) *Client {
	return &Client{
		baseURL:  controlApiUrl,
		L1Client: clients.NewL1Client(l1config, &iotaclient.WaitParams{Attempts: 10, DelayBetweenAttempts: 2}),
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
func (c *Client) Health(ctx context.Context) error {
	resp, err := c.httpClient.Get(c.baseURL + "/")
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	return nil
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

func (c *Client) ProgressL1() error {
	_, err := c.CreateCheckpoint()
	if err != nil {
		return err
	}

	err = c.AdvanceClock(1000)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Publish(
	ctx context.Context,
	req iotaclient.PublishRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	ret, err := c.L1Client.Publish(ctx, req)
	if err != nil {
		return nil, err
	}
	err = c.ProgressL1()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *Client) MergeCoins(
	ctx context.Context,
	req iotaclient.MergeCoinsRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	ret, err := c.L1Client.MergeCoins(ctx, req)
	if err != nil {
		return nil, err
	}
	err = c.ProgressL1()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *Client) ExecuteTransactionBlock(
	ctx context.Context,
	req iotaclient.ExecuteTransactionBlockRequest,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	ret, err := c.L1Client.ExecuteTransactionBlock(ctx, req)
	if err != nil {
		return nil, err
	}
	err = c.ProgressL1()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *Client) DryRunTransaction(
	ctx context.Context,
	txDataBytes iotago.Base64Data,
) (*iotajsonrpc.DryRunTransactionBlockResponse, error) {
	ret, err := c.L1Client.DryRunTransaction(ctx, txDataBytes)
	if err != nil {
		return nil, err
	}
	err = c.ProgressL1()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *Client) SplitCoin(
	ctx context.Context,
	req iotaclient.SplitCoinRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	ret, err := c.L1Client.SplitCoin(ctx, req)
	if err != nil {
		return nil, err
	}
	err = c.ProgressL1()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *Client) SignAndExecuteTransaction(
	ctx context.Context,
	req *iotaclient.SignAndExecuteTransactionRequest,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	ret, err := c.L1Client.SignAndExecuteTransaction(ctx, req)
	if err != nil {
		return nil, err
	}
	err = c.ProgressL1()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *Client) PublishContract(
	ctx context.Context,
	signer iotasigner.Signer,
	modules []*iotago.Base64Data,
	dependencies []*iotago.Address,
	gasBudget uint64,
	options *iotajsonrpc.IotaTransactionBlockResponseOptions,
) (*iotajsonrpc.IotaTransactionBlockResponse, *iotago.PackageID, error) {
	ret, packageID, err := c.PublishContract(ctx, signer, modules, dependencies, gasBudget, options)
	if err != nil {
		return nil, nil, err
	}
	err = c.ProgressL1()
	if err != nil {
		return nil, nil, err
	}
	return ret, packageID, nil
}

func (c *Client) RequestFunds(ctx context.Context, address cryptolib.Address) error {
	err := c.L1Client.RequestFunds(ctx, address)
	if err != nil {
		return err
	}
	err = c.ProgressL1()
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) DeployISCContracts(ctx context.Context, signer iotasigner.Signer) (iotago.PackageID, error) {
	ret, err := c.L1Client.DeployISCContracts(ctx, signer)
	if err != nil {
		return iotago.PackageID{}, err
	}
	err = c.ProgressL1()
	if err != nil {
		return iotago.PackageID{}, err
	}
	return ret, nil
}
