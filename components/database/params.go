package database

import (
	"github.com/iotaledger/hive.go/app"
)

// ParametersDatabase contains the definition of the parameters used by the ParametersDatabase.
type ParametersDatabase struct {
	// Engine defines the used database engine (rocksdb/mapdb).
	Engine string `default:"rocksdb" usage:"the used database engine (rocksdb/mapdb)"`

	ChainState struct {
		// Path defines the path to the chain state databases folder.
		Path string `default:"waspdb/chains/data" usage:"the path to the chain state databases folder"`

		CacheSize uint64 `default:"33554432" usage:"size of the RocksDB block cache"`
	}

	// DebugSkipHealthCheck defines whether to ignore the check for corrupted databases.
	DebugSkipHealthCheck bool `default:"true" usage:"ignore the check for corrupted databases"`
}

var ParamsDatabase = &ParametersDatabase{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"db": ParamsDatabase,
	},
	Masked: nil,
}
