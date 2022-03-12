package snapshot

import (
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/stretchr/testify/require"
	"math/rand"
	"os"
	"testing"
	"time"
)

func genRndDict(n int) dict.Dict {
	ret := dict.New()

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < n; i++ {
		f := rnd.Intn(1_000_000) % 5
		k := make([]byte, f*10)
		rnd.Read(k)
		f = rnd.Intn(1_000_000) % 10
		v := make([]byte, f*10)
		rnd.Read(v)
		ret.Set(kv.Key(k), v)
	}
	return ret
}

func Test1(t *testing.T) {
	db := mapdb.NewMapDB()
	chainID := testmisc.RandChainID()
	st, err := state.CreateOriginState(db, chainID)
	require.NoError(t, err)

	for k, v := range genRndDict(100) {
		st.KVStore().Set(k, v)
	}
	err = st.Save()
	require.NoError(t, err)

	glb := coreutil.NewChainStateSync()
	glb.SetSolidIndex(0)
	ordr := state.NewOptimisticStateReader(db, glb)
	ordr.SetBaseline()

	err = WriteSnapshotToFile(ordr, "", ConsoleReportParams{
		Console:           os.Stdout,
		StatsEveryKVPairs: 10,
	})
	require.NoError(t, err)
}
