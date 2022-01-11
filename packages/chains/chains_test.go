package chains

import (
	"net"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

func TestBasic(t *testing.T) {
	logger := testlogger.NewLogger(t)
	db, err := database.NewMemDB()
	require.NoError(t, err)
	getOrCreateKVStore := func(chain *iscp.ChainID) kvstore.KVStore {
		return db.NewStore()
	}

	ch := New(logger, processors.NewConfig(), 10, time.Second, false, nil, getOrCreateKVStore)

	nconn := txstream.New("dummyID", logger, func() (addr string, conn net.Conn, err error) {
		return "", nil, xerrors.New("dummy dial error")
	})
	ch.Attach(nconn)
}
