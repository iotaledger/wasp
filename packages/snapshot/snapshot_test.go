package snapshot

import (
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/kvtest"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func Test1(t *testing.T) {
	db := mapdb.NewMapDB()
	chainID := testmisc.RandChainID()
	st, err := state.CreateOriginState(db, chainID)
	require.NoError(t, err)

	rndKVStream := kvtest.NewRandStreamIterator(kvtest.RandStreamParams{
		Seed:       time.Now().UnixNano(),
		NumKVPairs: 1_000_000,
		MaxKey:     48,
		MaxValue:   128,
	})
	tm := util.NewTimer()
	count := 0
	totalBytes := 0
	err = rndKVStream.Iterate(func(k []byte, v []byte) bool {
		st.KVStore().Set(kv.Key(k), v)
		count++
		totalBytes += len(k) + len(v) + 6
		return true
	})
	t.Logf("write %d kv pairs, %d Mbytes, to in-memory state took %v", count, totalBytes/(1024*1024), tm.Duration())

	require.NoError(t, err)
	upd := state.NewStateUpdateWithBlockLogValues(1, time.Now(), testmisc.RandVectorCommitment())
	st.ApplyStateUpdate(upd)

	tm = util.NewTimer()
	err = st.Save()
	t.Logf("commit and save state to in-memory db took %v", tm.Duration())

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

	fname := FileName(chid, stateidx)
	t.Logf("file: %s", fname)

	tm = util.NewTimer()
	err = WriteSnapshot(ordr, "", ConsoleReportParams{
		Console:           os.Stdout,
		StatsEveryKVPairs: 1_000_000,
	})
	require.NoError(t, err)
	t.Logf("write snapshot took %v", tm.Duration())

	v, err := ScanFile(fname)
	require.NoError(t, err)
	require.True(t, chid.Equals(v.ChainID))
	require.EqualValues(t, stateidx, v.StateIndex)
	require.True(t, ts.Equal(v.TimeStamp))
}
