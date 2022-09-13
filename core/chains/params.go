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

type ParametersRawBlocks struct {
	Enabled   bool   `default:"false" usage:"whether the raw blocks plugin is enabled"`
	Directory string `default:"blocks" usage:"the raw blocks path"`
}

var (
	ParamsChains    = &ParametersChains{}
	ParamsRawBlocks = &ParametersRawBlocks{}
)

var params = &app.ComponentParams{
	Params: map[string]any{
		"chains":    ParamsChains,
		"rawBlocks": ParamsRawBlocks,
	},
	Masked: nil,
}
