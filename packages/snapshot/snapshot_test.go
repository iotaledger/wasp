package snapshot

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

func Test1(t *testing.T) {
	db := mapdb.NewMapDB()
	st := origin.InitChain(state.NewStore(db), nil, 0)

	tm := util.NewTimer()
	count := 0
	totalBytes := 0

	latest, err := st.LatestBlock()
	require.NoError(t, err)
	sd, err := st.NewStateDraft(time.Now(), latest.L1Commitment())
	require.NoError(t, err)

	seed := time.Now().UnixNano()
	t.Log("seed:", seed)
	rnd := util.NewPseudoRand(seed)
	for i := 0; i < 1000; i++ {
		k := randByteSlice(rnd, 4+1, 48) // key is hname + key
		v := randByteSlice(rnd, 1, 128)

		sd.Set(kv.Key(k), v)
		count++
		totalBytes += len(k) + len(v) + 6
	}

	t.Logf("write %d kv pairs, %d Mbytes, to in-memory state took %v", count, totalBytes/(1024*1024), tm.Duration())

	tm = util.NewTimer()
	block := st.Commit(sd)
	err = st.SetLatest(block.TrieRoot())
	require.NoError(t, err)
	t.Logf("commit and save state to in-memory db took %v", tm.Duration())

	rdr, err := st.LatestState()
	require.NoError(t, err)

	stateidx := rdr.BlockIndex()
	ts := rdr.Timestamp()

	fname := FileName(stateidx)
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
	require.EqualValues(t, stateidx, v.StateIndex)
	require.True(t, ts.Equal(v.TimeStamp))
}

func randByteSlice(rnd *rand.Rand, minLength, maxLength int) []byte {
	n := rnd.Intn(maxLength-minLength) + minLength
	b := make([]byte, n)
	rnd.Read(b)
	return b
}
