package chains

import (
	"time"

	"github.com/iotaledger/hive.go/app"
)

type ParametersChains struct {
	BroadcastUpToNPeers              int           `default:"2" usage:"number of peers an offledger request is broadcasted to"`
	BroadcastInterval                time.Duration `default:"5s" usage:"time between re-broadcast of offledger requests"`
	APICacheTTL                      time.Duration `default:"300s" usage:"time to keep processed offledger requests in api cache"`
	PullMissingRequestsFromCommittee bool          `default:"true" usage:"whether or not to pull missing requests from other committee members"`
	DeriveAliasOutputByQuorum        bool          `default:"true" usage:"false means we propose own AliasOutput, true - by majority vote."`
	PipeliningLimit                  int           `default:"-1" usage:"-1 -- infinite, 0 -- disabled, X -- build the chain if there is up to X transactions unconfirmed by L1."`
	ConsensusDelay                   time.Duration `default:"500ms" usage:"Minimal delay between consensus runs."`
	RecoveryTimeout                  time.Duration `default:"20s" usage:"Time after which another consensus attempt is made."`
	RedeliveryPeriod                 time.Duration `default:"2s" usage:"the resend period for msg."`
	PrintStatusPeriod                time.Duration `default:"3s" usage:"the period to print consensus instance status."`
	ConsensusInstsInAdvance          int           `default:"3" usage:""`
	AwaitReceiptCleanupEvery         int           `default:"100" usage:"for every this number AwaitReceipt will be cleaned up"`
}

type ParametersWAL struct {
	LoadToStore bool   `default:"false" usage:"load blocks from \"write-ahead log\" to the store on node start-up"`
	Enabled     bool   `default:"true" usage:"whether the \"write-ahead logging\" is enabled"`
	Path        string `default:"waspdb/wal" usage:"the path to the \"write-ahead logging\" folder"`
}

type ParametersValidator struct {
	Address string `default:"" usage:"bech32 encoded address to identify the node (as access node on gov contract and to collect validator fee payments)"`
}

type ParametersStateManager struct {
	BlockCacheMaxSize                 int           `default:"1000" usage:"how many blocks may be stored in cache before old ones start being deleted"`
	BlockCacheBlocksInCacheDuration   time.Duration `default:"1h" usage:"how long should the block stay in block cache before being deleted"`
	BlockCacheBlockCleaningPeriod     time.Duration `default:"1m" usage:"how often should the block cache be cleaned"`
	StateManagerGetBlockRetry         time.Duration `default:"3s" usage:"how often get block requests should be repeated"`
	StateManagerRequestCleaningPeriod time.Duration `default:"1s" usage:"how often requests waiting for response should be checked for expired context"`
	StateManagerTimerTickPeriod       time.Duration `default:"1s" usage:"how often timer tick fires in state manager"`
	PruningMinStatesToKeep            int           `default:"10000" usage:"this number of states will always be available in the store; if 0 - store pruning is disabled"`
	PruningMaxStatesToDelete          int           `default:"1000" usage:"on single store pruning attempt at most this number of states will be deleted"`
}

type ParametersSnapshotManager struct {
	SnapshotsToLoad []string `default:"" usage:"list of snapshots to load; can be either single block hash of a snapshot (if a single chain has to be configured) or list of '<chainID>:<blockHash>' to configure many chains"`
	Period          uint32   `default:"0" usage:"how often state snapshots should be made: 1000 meaning \"every 1000th state\", 0 meaning \"making snapshots is disabled\""`
	Delay           uint32   `default:"20" usage:"how many states should pass before snapshot is produced"`
	LocalPath       string   `default:"waspdb/snap" usage:"the path to the snapshots folder in this node's disk"`
	NetworkPaths    []string `default:"" usage:"the list of paths to the remote (http(s)) snapshot locations; each of listed locations must contain 'INDEX' file with list of snapshot files"`
}

var (
	ParamsChains          = &ParametersChains{}
	ParamsWAL             = &ParametersWAL{}
	ParamsValidator       = &ParametersValidator{}
	ParamsStateManager    = &ParametersStateManager{}
	ParamsSnapshotManager = &ParametersSnapshotManager{}
)

var params = &app.ComponentParams{
	Params: map[string]any{
		"chains":       ParamsChains,
		"wal":          ParamsWAL,
		"validator":    ParamsValidator,
		"stateManager": ParamsStateManager,
		"snapshots":    ParamsSnapshotManager,
	},
	Masked: nil,
}
