package chains

import (
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/database"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	logger := testlogger.NewLogger(t)
	db, err := database.NewMemDB()
	require.NoError(t, err)
	getOrCreateKVStore := func(chain *iscp.ChainID) kvstore.KVStore {
		return db.NewStore()
	}

	_ = New(logger, processors.NewConfig(), 10, time.Second, false, nil, getOrCreateKVStore)
}
