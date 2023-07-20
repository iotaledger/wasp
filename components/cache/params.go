package cache

import (
	"time"

	"github.com/iotaledger/hive.go/app"
)

// ParametersDatabase contains the definition of the parameters used by the ParametersDatabase.
type ParametersCache struct {
	// Engine defines the used database engine (rocksdb/mapdb).
	CacheSize string `default:"512MiB" usage:"cache size"`

	// Engine defines the used database engine (rocksdb/mapdb).
	CacheStatsInterval time.Duration `default:"30s" usage:"interval for printing cache statistics"`

	// DebugSkipHealthCheck defines whether to ignore the check for corrupted databases.
	CacheEnabled bool `default:"true" usage:"enables or disables caching of states"`
}

var ParamsCache = &ParametersCache{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"cache": ParamsCache,
	},
	Masked: nil,
}
