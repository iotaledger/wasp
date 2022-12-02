package main

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	hivedb "github.com/iotaledger/hive.go/core/database"
	"github.com/iotaledger/trie.go/trie"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/snapshot"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/panicutil"
)

// implements 'snap-cli', a snapshot tool for Wasp databases

const usage = `USAGE: snap-cli [-create | -scanfile | -restoredb | -verify] <filename>
or
USAGE: snap-cli -validate <chainID> <L1 API endpoint> <L2 API endpoint>
or
USAGE: snap-cli -extractblocks <chainID> <fromIndex> <target directory>
`

func main() {
	panic("TODO - this package is broken since the trie package was remove from trie.go")

	ensureMinimumArgs(3)
	cmd := os.Args[1]
	param := os.Args[2]
	switch cmd {
	case "-create":
		createSnapshot(param)
	case "-scanfile":
		scanFile(param)
	case "-restoredb":
		fmt.Printf("creating db from snapshot file %s\n", param)
		prop := scanFile(param)
		restoreDb(prop)
	case "-verify":
		fmt.Printf("verifying state against snapshot file %s\n", param)
		prop := scanFile(param)
		verify(prop)
	case "-validate":
		ensureMinimumArgs(4)
		// TODO implement me
		fmt.Printf("'validate' option is NOT IMPLEMENTED\n")
		os.Exit(1)
	case "-extractblocks":
		ensureMinimumArgs(5)
		fromInt, err := strconv.Atoi(os.Args[3])
		mustNoErr(err)
		extractBlocks(param, uint32(fromInt), os.Args[4])
	default:
		fmt.Printf("%s\n", usage)
		os.Exit(1)
	}
}

func ensureMinimumArgs(n int) {
	if len(os.Args) < n {
		fmt.Printf("%s\n", usage)
		os.Exit(1)
	}
}

func mustNoErr(err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

func dbdirFromSnapshotFile(fname string) string {
	psplit := strings.Split(fname, ".")
	if len(psplit) < 1 {
		fmt.Printf("error: cannot parse directory name\n")
		os.Exit(1)
	}
	return psplit[0]
}

func scanFile(fname string) *snapshot.FileProperties {
	fmt.Printf("scaning snapshot file %s\n", fname)
	dbDir := dbdirFromSnapshotFile(fname)
	fmt.Printf("assuming chainID and DB directory name is (taken from file name): %s\n", dbDir)
	tm := util.NewTimer()
	prop, err := snapshot.ScanFile(fname)
	mustNoErr(err)
	fmt.Printf("scan file took %v\n", tm.Duration())
	fmt.Printf("Chain ID (implied from directory name): %s\n", dbDir)
	fmt.Printf("State index: %d\n", prop.StateIndex)
	fmt.Printf("Timestamp: %v\n", prop.TimeStamp)
	fmt.Printf("Number of records: %d\n", prop.NumRecords)
	fmt.Printf("Total bytes: %d MB\n", prop.Bytes/(1024*1024))
	fmt.Printf("Longest key: %d\n", prop.MaxKeyLen)
	return prop
}

func restoreDb(prop *snapshot.FileProperties) {
	dbDir := dbdirFromSnapshotFile(prop.FileName)
	kvstream, err := kv.OpenKVStreamFile(prop.FileName)
	mustNoErr(err)
	if _, err := os.Stat(dbDir); !os.IsNotExist(err) {
		fmt.Printf("directory %s already exists. Can't create new database\n", dbDir)
		os.Exit(1)
	}
	fmt.Printf("creating new database for chain ID %s\n", dbDir)
	db, err := database.DatabaseWithDefaultSettings(dbDir, true, hivedb.EngineRocksDB, false)
	mustNoErr(err)
	st := state.NewVirtualState(db.KVStore())

	const persistEach = 1_000_000

	var count int
	var errW error
	tm := util.NewTimer()
	err = kvstream.Iterate(func(k []byte, v []byte) bool {
		st.KVStore().Set(kv.Key(k), v)
		count++
		if count%persistEach == 0 {
			if errW = st.Save(); errW != nil {
				return false
			}
			fmt.Printf("committed %d total records to database. It took %v\n", count, tm.Duration())
		}
		return true
	})
	mustNoErr(err)
	if err = st.Save(); errW != nil {
		fmt.Printf("error: %d\n", err)
		os.Exit(1)
	}
	fmt.Printf("committed %d total records to database. It took %v\n", count, tm.Duration())

	c := trie.RootCommitment(st.TrieNodeStore())
	fmt.Printf("Success. Root commitment: %s\n", c)
}

func verify(prop *snapshot.FileProperties) {
	kvstream, err := kv.OpenKVStreamFile(prop.FileName)
	mustNoErr(err)
	dbDir := dbdirFromSnapshotFile(prop.FileName)
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		fmt.Printf("directory %s does not exists\n", dbDir)
		os.Exit(1)
	}
	fmt.Printf("verifying database for chain ID/dbDir %s\n", dbDir)

	db, err := database.DatabaseWithDefaultSettings(dbDir, false, hivedb.EngineAuto, false)
	if err != nil {
		panic(err)
	}
	mustNoErr(err)

	st := state.NewVirtualState(db.KVStore())
	c := state.RootCommitment(st.TrieNodeStore())
	fmt.Printf("root commitment is %s\n", c)

	var chainID *isc.ChainID
	err = panicutil.CatchPanic(func() {
		chainID = st.ChainID()
	})
	mustNoErr(err)
	if !prop.ChainID.Equals(chainID) {
		fmt.Printf("chain IDs in db and in file do not match the state in the database")
		os.Exit(1)
	}

	const reportEach = 10_000
	var count int
	var errW error

	tm := util.NewTimer()

	const useCachedNodeStore = true
	const clearCacheEach = 100_000

	var nodeStore trie.NodeStore
	if useCachedNodeStore {
		nodeStore = st.TrieNodeStore()
	} else {
		// this is up to 3-4 times slower
		glb := coreutil.NewChainStateSync()
		glb.SetSolidIndex(st.BlockIndex())
		rdr := st.OptimisticStateReader(glb)
		rdr.SetBaseline()
		nodeStore = rdr.TrieNodeStore()
	}

	err = kvstream.Iterate(func(k, v []byte) bool {
		proof := state.GetMerkleProof(k, nodeStore)
		if errW = state.ValidateMerkleProof(proof, c, v); errW != nil {
			return false
		}
		count++
		if count%reportEach == 0 {
			took := tm.Duration()
			fmt.Printf("verified total %d records. Took %v, %d rec/sec\n", count, took, (1000*int64(count))/took.Milliseconds())
		}
		if useCachedNodeStore && count%clearCacheEach == 0 {
			_ = st.Save() // just clears trie cache, to prevent the whole trie coming to memory
		}
		return true
	})
	mustNoErr(err)
	mustNoErr(errW)
	fmt.Printf("verified total %d records. It took %v\n", count, tm.Duration())

	fmt.Printf("success: file %s match the database\n", prop.FileName)
}

func createSnapshot(dbDir string) {
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		fmt.Printf("directory %s does not exists\n", dbDir)
		os.Exit(1)
	}
	fmt.Printf("creating shapshot for directory/chain ID %s\n", dbDir)

	db, err := database.DatabaseWithDefaultSettings(dbDir, false, hivedb.EngineAuto, false)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	st := state.NewVirtualState(db.KVStore())

	var stateIndex uint32
	err = panicutil.CatchPanic(func() {
		stateIndex = st.BlockIndex()
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	fname := fmt.Sprintf("%s.%d.snapshot", dbDir, stateIndex)
	fmt.Printf("will be writing to file %s\n", fname)

	kvwriter, err := kv.CreateKVStreamFile(fname)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = kvwriter.File.Close() }()

	const reportEach = 10_000
	var errW error

	tm := util.NewTimer()

	err = st.KVStoreReader().Iterate("", func(k kv.Key, v []byte) bool {
		if errW = kvwriter.Write([]byte(k), v); errW != nil {
			return false
		}
		count, byteCount := kvwriter.Stats()
		if count%reportEach == 0 {
			fmt.Printf("wrote %d key/value pairs, %d bytes. It took %v\n", count, byteCount, tm.Duration())
		}
		return true
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		func() { _ = kvwriter.File.Close() }()
		os.Exit(1) //nolint:gocritic
	}
	if errW != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	count, byteCount := kvwriter.Stats()
	fmt.Printf("wrote TOTAL %d key/value pairs, %d bytes. It took %v\n", count, byteCount, tm.Duration())
}

func extractBlocks(dbDir string, from uint32, targetDir string) {
	fmt.Printf("Database directory/chainID: %s\n", dbDir)
	fmt.Printf("Extracting blocks starting from #%d\n", from)

	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		fmt.Printf("directory %s does not exists\n", dbDir)
		os.Exit(1)
	}
	fmt.Printf("using database for chain ID/dbDir %s\n", dbDir)
	fmt.Printf("writing files to directory '%s'\n", targetDir)

	db, err := database.DatabaseWithDefaultSettings(dbDir, false, hivedb.EngineAuto, false)
	mustNoErr(err)

	err = os.MkdirAll(targetDir, 0o777)
	mustNoErr(err)

	indices := make([]uint32, 0)
	store := db.KVStore()
	err = state.ForEachBlockIndex(store, func(blockIndex uint32) bool {
		indices = append(indices, blockIndex)
		return true
	})
	mustNoErr(err)
	sort.Slice(indices, func(i, j int) bool {
		return indices[i] < indices[j]
	})
	var blk state.Block
	for _, idx := range indices {
		if idx < from {
			continue
		}
		blk, err = state.LoadBlock(store, idx)
		if err != nil {
			fmt.Printf("error: failed to load block data for index #%d: %v\n", idx, err)
			continue
		}
		fname := snapshot.BlockFileName(dbDir, idx, blk.GetHash())
		fullName := path.Join(targetDir, fname)
		err = os.WriteFile(fullName, blk.Bytes(), 0o600)
		if err != nil {
			fmt.Printf("error: failed to write block data to file %s: %v\n", fullName, err)
			continue
		}
		fmt.Printf("block #%d -> %s\n", idx, fullName)
	}
}
