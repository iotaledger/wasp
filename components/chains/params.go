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

type ParametersSnapshotter struct {
	Period uint32 `default:"0" usage:"how often state snapshots should be made: 1000 meaning \"every 1000th state\", 0 meaning \"making snapshots is disabled\""`
	Path   string `default:"waspdb/snap" usage:"the path to the snapshots folder"`
}

var (
	ParamsChains      = &ParametersChains{}
	ParamsWAL         = &ParametersWAL{}
	ParamsSnapshotter = &ParametersSnapshotter{}
)

var params = &app.ComponentParams{
	Params: map[string]any{
		"chains":    ParamsChains,
		"wal":       ParamsWAL,
		"snapshots": ParamsSnapshotter,
	},
	Masked: nil,
}
