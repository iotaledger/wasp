// Package l1paramsfetcher provides functionality to fetch and cache L1 protocol parameters.
// It automatically refreshes parameters when the current epoch expires.
package l1paramsfetcher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/clients/iota-go/client"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/parameters"
)

// L1ParamsFetcher provides the latest version of L1Params, and
// automatically refreshes it when the epoch is out of date
type L1ParamsFetcher interface {
	GetOrFetchLatest(ctx context.Context) (*parameters.L1Params, error)
}

type l1ParamsFetcher struct {
	client client.IotaClient
	log    log.Logger
	mu     sync.Mutex
	latest *parameters.L1Params
}

// NewL1ParamsFetcher creates a new L1ParamsFetcher
func NewL1ParamsFetcher(iotaClient client.IotaClient, log log.Logger) L1ParamsFetcher {
	return &l1ParamsFetcher{
		client: iotaClient,
		log:    log.NewChildLogger("L1ParamsFetcher"),
	}
}

// GetOrFetchLatest returns the latest L1Params, or fetches it if necessary
func (f *l1ParamsFetcher) GetOrFetchLatest(ctx context.Context) (*parameters.L1Params, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.shouldFetch() {
		f.log.LogInfo("Fetching latest L1Params...")
		latest, err := FetchLatest(ctx, f.client)
		if err != nil {
			f.log.LogError("Failed to fetch latest L1Params", err)
			return nil, err
		}
		f.latest = latest
	}

	return f.latest, nil
}

func (f *l1ParamsFetcher) shouldFetch() bool {
	if f.latest == nil {
		return true
	}
	now := time.Now()
	start := time.Unix(f.latest.Protocol.EpochStartTimestampMs.Int64(), 0)
	duration := time.Duration(f.latest.Protocol.EpochDurationMs.Int64()) * time.Millisecond
	return now.After(start.Add(duration))
}

// FetchLatest fetches the latest L1Params from L1, retrying on failure
func FetchLatest(ctx context.Context, iotaClient client.IotaClient) (*parameters.L1Params, error) {
	return iotaclient.Retry(
		ctx,
		func() (*parameters.L1Params, error) {
			system, err := iotaClient.GetLatestIotaSystemState(ctx)
			if err != nil {
				return nil, fmt.Errorf("can't get latest system state: %w", err)
			}
			meta, err := iotaClient.GetCoinMetadata(ctx, iotagraphql.IotaCoinType.String())
			if err != nil {
				return nil, fmt.Errorf("can't get coin metadata: %w", err)
			}
			if meta.Decimals != parameters.BaseTokenDecimals {
				return nil, fmt.Errorf("unsupported decimals: %d", meta.Decimals)
			}
			return &parameters.L1Params{
				Protocol: &parameters.Protocol{
					Epoch:                 system.Epoch,
					ProtocolVersion:       system.ProtocolVersion,
					SystemStateVersion:    system.SystemStateVersion,
					ReferenceGasPrice:     system.ReferenceGasPrice,
					EpochStartTimestampMs: system.EpochStartTimestampMs,
					EpochDurationMs:       system.EpochDurationMs,
				},
				BaseToken: parameters.IotaCoinInfoFromL1Metadata(
					coin.BaseTokenType,
					meta,
					coin.Value(system.IotaTotalSupply.Uint64()),
				),
			}, nil
		},
		iotaclient.DefaultRetryCondition[*parameters.L1Params](),
		iotaclient.WaitForEffectsEnabled,
	)
}
