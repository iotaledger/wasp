package main

import (
	"fmt"
	"os"

	"github.com/iotaledger/wasp/packages/database/dbmanager"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/trie"
	"github.com/iotaledger/wasp/packages/snapshot"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/panicutil"
)

const usage = "USAGE: snapshot [-create | -scanfile | -restoredb] <filename>\n"

func main() {
	if len(os.Args) < 3 {
		fmt.Printf(usage)
		os.Exit(1)
	}
	cmd := os.Args[1]
	param := os.Args[2]
	switch cmd {
	case "-create", "--create":
		chainID, err := iscp.ChainIDFromString(param)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}
		createSnapshot(chainID)
	case "-scanfile", "--scanfile":
		scanFile(param)
	case "-restoredb", "--restoredb":
		fmt.Printf("creating db from snapshot file %s\n", param)
		prop := scanFile(param)
		restoreDb(prop)
	case "-verify", "--verify":
		fmt.Printf("verifying state against snapshot file %s\n", param)
		prop := scanFile(param)
		verify(prop)
	default:
		fmt.Printf(usage)
		os.Exit(1)
	}
}

func scanFile(fname string) *snapshot.FileProperties {
	fmt.Printf("scaning snapshot file %s\n", fname)
	tm := util.NewTimer()
	prop, err := snapshot.ScanFile(fname)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("scan file took %v\n", tm.Duration())
	fmt.Printf("Chain ID: %s\n", prop.ChainID)
	fmt.Printf("State index: %d\n", prop.StateIndex)
	fmt.Printf("Timestamp: %v\n", prop.TimeStamp)
	fmt.Printf("Number of records: %d\n", prop.NumRecords)
	fmt.Printf("Total bytes: %d MB\n", prop.Bytes/(1024*1024))
	fmt.Printf("Longest key: %d\n", prop.MaxKeyLen)
	return prop
}

func restoreDb(prop *snapshot.FileProperties) {
	kvstream, err := kv.OpenKVStreamFile(prop.FileName)
	if err != nil {
		fmt.Printf("error: %d\n", err)
		os.Exit(1)
	}
	dbDir := prop.ChainID.String()
	if _, err := os.Stat(dbDir); !os.IsNotExist(err) {
		fmt.Printf("directory %s already exists. Can't create new database\n", dbDir)
		os.Exit(1)
	}
	fmt.Printf("creating new database for chain ID %s\n", prop.ChainID)
	db, err := dbmanager.NewDB(dbDir)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	st := state.NewVirtualState(db.NewStore())

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
	if err != nil {
		fmt.Printf("error: %d\n", err)
		os.Exit(1)
	}
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
	if err != nil {
		fmt.Printf("error: %d\n", err)
		os.Exit(1)
	}
	dbDir := prop.ChainID.String()
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		fmt.Printf("directory %s does not exists\n", dbDir)
		os.Exit(1)
	}
	fmt.Printf("verifying database for chain ID %s\n", prop.ChainID)

	db, err := dbmanager.NewDB(dbDir)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	st := state.NewVirtualState(db.NewStore())
	c := trie.RootCommitment(st.TrieNodeStore())
	fmt.Printf("root commitment is %s\n", c)

	var chainID *iscp.ChainID
	err = panicutil.CatchPanic(func() {
		chainID = st.ChainID()
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	if !prop.ChainID.Equals(chainID) {
		fmt.Printf("chain IDs in db and in file do not match the state in the database: %s != %s\n", chainID, prop.ChainID)
		os.Exit(1)
	}

	const reportEach = 100_000
	var count int
	var errW error

	tm := util.NewTimer()

	const useCachedNodeStore = true
	const clearCacheEach = 100_000

	var nodeStore trie.NodeStore
	if useCachedNodeStore {
		nodeStore = st.TrieNodeStore()
	} else {
		// this is uo to 3-4 times slower
		glb := coreutil.NewChainStateSync()
		glb.SetSolidIndex(st.BlockIndex())
		rdr := st.OptimisticStateReader(glb)
		rdr.SetBaseline()
		nodeStore = rdr.TrieNodeStore()
	}

	err = kvstream.Iterate(func(k, v []byte) bool {
		proof := state.CommitmentModel.Proof(k, nodeStore)
		if errW = proof.Validate(c, v); errW != nil {
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
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	if errW != nil {
		fmt.Printf("error: %v\n", errW)
		os.Exit(1)
	}
	fmt.Printf("verified total %d records. It took %v\n", count, tm.Duration())

	fmt.Printf("success: file %s match the database\n", prop.FileName)
}

func createSnapshot(chainID *iscp.ChainID) {
	dbDir := chainID.String()
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		fmt.Printf("directory %s does not exists\n", dbDir)
		os.Exit(1)
	}
	fmt.Printf("creating shapshot for chain ID %s\n", chainID)

	db, err := dbmanager.NewDB(dbDir)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	st := state.NewVirtualState(db.NewStore())

	var stateIndex uint32
	err = panicutil.CatchPanic(func() {
		stateIndex = st.BlockIndex()
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	fname := fmt.Sprintf("%s.%d.snapshot", chainID, stateIndex)
	fmt.Printf("will be writing to file %s\n", fname)

	kvwriter, err := kv.CreateKVStreamFile(fname)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	defer kvwriter.File.Close()

	const reportEach = 100_000
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
		os.Exit(1)
	}
	if errW != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	count, byteCount := kvwriter.Stats()
	fmt.Printf("wrote TOTAL %d key/value pairs, %d bytes. It took %v\n", count, byteCount, tm.Duration())
}
