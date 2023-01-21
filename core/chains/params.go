package chains

import (
	"time"

	"github.com/iotaledger/hive.go/core/app"
)

type ParametersChains struct {
	BroadcastUpToNPeers              int           `default:"2" usage:"number of peers an offledger request is broadcasted to"`
	BroadcastInterval                time.Duration `default:"5s" usage:"time between re-broadcast of offledger requests"`
	APICacheTTL                      time.Duration `default:"300s" usage:"time to keep processed offledger requests in api cache"`
	PullMissingRequestsFromCommittee bool          `default:"true" usage:"whether or not to pull missing requests from other committee members"`
}

type ParametersWAL struct {
	Enabled   bool   `default:"true" usage:"whether the WAL is enabled"`
	Directory string `default:"wal" usage:"the path for WAL blocks"`
}

var (
	ParamsChains = &ParametersChains{}
	ParamsWAL    = &ParametersWAL{}
)

var params = &app.ComponentParams{
	Params: map[string]any{
		"chains": ParamsChains,
		"wal":    ParamsWAL,
	},
	Masked: nil,
}
