package snapshot

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/trie"
	"github.com/iotaledger/wasp/packages/state"
	"io"
	"path"
)

type ConsoleReportParams struct {
	Console           io.Writer
	StatsEveryKVPairs int
}

// WriteSnapshotToFile dumps k/v pairs of the state into the
// file named '<chainID>.<block index>.snapshot' in a directory
// Keys are not sorted therefore the result in general is not deterministic
func WriteSnapshotToFile(reader state.OptimisticStateReader, dir string, p ...ConsoleReportParams) error {
	par := ConsoleReportParams{
		Console:           io.Discard,
		StatsEveryKVPairs: 100,
	}
	if len(p) > 0 {
		par = p[0]
	}
	chid, err := reader.ChainID()
	if err != nil {
		return err
	}
	blockIndex, err := reader.BlockIndex()
	if err != nil {
		return err
	}
	c := trie.RootCommitment(reader.TrieNodeStore())
	fmt.Fprintf(par.Console, "[WriteSnapshotToFile] chinID: %s\n", chid)
	fmt.Fprintf(par.Console, "[WriteSnapshotToFile] block index: %d\n", blockIndex)
	fmt.Fprintf(par.Console, "[WriteSnapshotToFile] state commitment: %s\n", c)

	fname := fmt.Sprintf("%s.%d.snapshot", chid, blockIndex)
	fmt.Fprintf(par.Console, "[WriteSnapshotToFile] will be writing snapshot to file '%s'\n", fname)

	snapshot, err := CreateSnapshotFile(path.Join(dir, fname))
	if err != nil {
		fmt.Fprintf(par.Console, "[WriteSnapshotToFile] error: %v\n", err)
		return err
	}
	defer snapshot.File.Close()

	var errW error
	err = reader.KVStoreReader().Iterate("", func(k kv.Key, v []byte) bool {
		if errW = snapshot.WriteKeyValue([]byte(k), v); errW != nil {
			return false
		}
		if par.StatsEveryKVPairs > 0 {
			kvCount, bCount := snapshot.Stats()
			if kvCount%par.StatsEveryKVPairs == 0 {
				fmt.Fprintf(par.Console, "[WriteSnapshotToFile] k/v pairs: %d, bytes: %d\n", kvCount, bCount)
			}
		}
		return true
	})
	if err != nil {
		fmt.Fprintf(par.Console, "[WriteSnapshotToFile] error: %v\n", err)
		return err
	}
	if errW != nil {
		fmt.Fprintf(par.Console, "[WriteSnapshotToFile] error: %v\n", errW)
		return errW
	}
	kvCount, bCount := snapshot.Stats()
	fmt.Fprintf(par.Console, "[WriteSnapshotToFile] ---- TOTAL: k/v pairs: %d, bytes: %d\n", kvCount, bCount)

	return nil
}
