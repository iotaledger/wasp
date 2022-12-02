package snapshot

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/kvtest"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

func Test1(t *testing.T) {
	db := mapdb.NewMapDB()
	st := state.InitChainStore(db)

	rndKVStream := kvtest.NewRandStreamIterator(kvtest.RandStreamParams{
		Seed:       time.Now().UnixNano(),
		NumKVPairs: 1_000,
		MaxKey:     48,
		MaxValue:   128,
	})
	tm := util.NewTimer()
	count := 0
	totalBytes := 0

	sd := st.NewStateDraft(time.Now(), st.LatestBlock().L1Commitment())

	err := rndKVStream.Iterate(func(k []byte, v []byte) bool {
		sd.Set(kv.Key(k), v)
		count++
		totalBytes += len(k) + len(v) + 6
		return true
	})
	require.NoError(t, err)
	t.Logf("write %d kv pairs, %d Mbytes, to in-memory state took %v", count, totalBytes/(1024*1024), tm.Duration())

	tm = util.NewTimer()
	block := st.Commit(sd)
	st.SetLatest(block.TrieRoot())
	t.Logf("commit and save state to in-memory db took %v", tm.Duration())

	require.NoError(t, err)

	rdr := st.LatestState()

	chid := rdr.ChainID()
	stateidx := rdr.BlockIndex()
	ts := rdr.Timestamp()

	fname := FileName(chid, stateidx)
	t.Logf("file: %s", fname)

	tm = util.NewTimer()
	err = WriteSnapshot(rdr, "", ConsoleReportParams{
		Console:           os.Stdout,
		StatsEveryKVPairs: 1_000_000,
	})
	require.NoError(t, err)
	t.Logf("write snapshot took %v", tm.Duration())
	defer os.Remove(fname)

	v, err := ScanFile(fname)
	require.NoError(t, err)
	require.True(t, chid.Equals(v.ChainID))
	require.EqualValues(t, stateidx, v.StateIndex)
	require.True(t, ts.Equal(v.TimeStamp))
}
