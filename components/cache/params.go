package cache

import (
	"time"

	"github.com/iotaledger/hive.go/app"
)

// ParametersDatabase contains the definition of the parameters used by the ParametersDatabase.
type ParametersCache struct {
	// CacheSize defines the maximum cache size
	CacheSize string `default:"64MiB" usage:"cache size"`

	// CacheStatsInterval is the interval for the statistics ticker
	CacheStatsInterval time.Duration `default:"30s" usage:"interval for printing cache statistics"`

	// CacheEnable enabled the cache
	Enabled bool `default:"true" usage:"whether the cache plugin is enabled"`
}

var ParamsCache = &ParametersCache{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"cache": ParamsCache,
	},
	Masked: nil,
}
