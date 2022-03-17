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
		f := rnd.Intn(1_000_000)%5 + 1
		k := make([]byte, f*10)
		rnd.Read(k)
		f = rnd.Intn(1_000_000) % 10
		v := make([]byte, f*10)
		rnd.Read(v)
		if len(v) == 0 {
			v = []byte{byte(rnd.Intn(1000) % 256)}
		}
		ret.Set(kv.Key(k), v)
	}
	return ret
}

func Test1(t *testing.T) {
	db := mapdb.NewMapDB()
	chainID := testmisc.RandChainID()
	st, err := state.CreateOriginState(db, chainID)
	require.NoError(t, err)

	for k, v := range genRndDict(1_000_000) {
		if len(v) == 0 {
			panic("len(v) == 0")
		}
		st.KVStore().Set(k, v)
	}
	upd := state.NewStateUpdateWithBlockLogValues(1, time.Now(), testmisc.RandVectorCommitment())
	st.ApplyStateUpdate(upd)
	err = st.Save()
	require.NoError(t, err)

	glb := coreutil.NewChainStateSync()
	glb.SetSolidIndex(0)
	ordr := state.NewOptimisticStateReader(db, glb)
	ordr.SetBaseline()

	chid, err := ordr.ChainID()
	require.NoError(t, err)
	stateidx, err := ordr.BlockIndex()
	require.NoError(t, err)
	ts, err := ordr.Timestamp()
	require.NoError(t, err)

	fname := SnapshotFileName(chid, stateidx)
	t.Logf("file: %s", fname)
	err = WriteSnapshot(ordr, "", ConsoleReportParams{
		Console:           os.Stdout,
		StatsEveryKVPairs: 100_000,
	})
	require.NoError(t, err)
	v, err := ScanFile(fname)
	require.NoError(t, err)
	require.True(t, chid.Equals(v.ChainID))
	require.EqualValues(t, stateidx, v.StateIndex)
	require.True(t, ts.Equal(v.TimeStamp))
}
