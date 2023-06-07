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
}

type ParametersWAL struct {
	Enabled bool   `default:"true" usage:"whether the \"write-ahead logging\" is enabled"`
	Path    string `default:"waspdb/wal" usage:"the path to the \"write-ahead logging\" folder"`
}

type ParametersStateManager struct {
	BlockCacheMaxSize                 int           `default:"1000" usage:"how many blocks may be stored in cache before old ones start being deleted"`
	BlockCacheBlocksInCacheDuration   time.Duration `default:"1h" usage:"how long should the block stay in block cache before being deleted"`
	BlockCacheBlockCleaningPeriod     time.Duration `default:"1m" usage:"how often should the block cache be cleaned"`
	StateManagerGetBlockRetry         time.Duration `default:"3s" usage:"how often get block requests should be repeated"`
	StateManagerRequestCleaningPeriod time.Duration `default:"1s" usage:"how often requests waiting for response should be checked for expired context"`
	StateManagerTimerTickPeriod       time.Duration `default:"1s" usage:"how often timer tick fires in state manager"`
	PruningMinStatesToKeep            int           `default:"10000" usage:"this number of states will always be available in the database"`
	PruningMaxStatesToDelete          int           `default:"1000" usage:"on single store pruning attempt at most this number of states will be deleted"`
}

var (
	ParamsChains       = &ParametersChains{}
	ParamsWAL          = &ParametersWAL{}
	ParamsStateManager = &ParametersStateManager{}
)

var params = &app.ComponentParams{
	Params: map[string]any{
		"chains":       ParamsChains,
		"wal":          ParamsWAL,
		"stateManager": ParamsStateManager,
	},
	Masked: nil,
}
