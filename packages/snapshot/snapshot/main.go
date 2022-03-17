package main

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/database/dbmanager"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/trie"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/panicutil"
	"os"

	"github.com/iotaledger/wasp/packages/snapshot"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("USAGE: snapshot [-scan || -createdb] <filename>\n")
		os.Exit(1)
	}
	cmd := os.Args[1]
	fname := os.Args[2]
	switch cmd {
	case "-scan", "--scan":
		scanFile(fname)
	case "-createdb", "--createdb":
		fmt.Printf("creating db from snapshot file %s\n", fname)
		prop := scanFile(fname)
		createDb(prop)
	case "-verify", "--verify":
		fmt.Printf("veryfying state against snapshot file %s\n", fname)
		prop := scanFile(fname)
		verify(prop)
	default:
		fmt.Printf("USAGE: snapshot [-scan || -createdb] <filename>\n")
		os.Exit(1)
	}
}

func scanFile(fname string) *snapshot.FileProperties {
	fmt.Printf("scaning snapshot file %s\n", fname)
	values, err := snapshot.ScanFile(fname)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Chain ID: %s\n", values.ChainID)
	fmt.Printf("State index: %d\n", values.StateIndex)
	fmt.Printf("Timestamp: %v\n", values.TimeStamp)
	fmt.Printf("Number of records: %d\n", values.NumRecords)
	fmt.Printf("Longest key: %d\n", values.MaxKeyLen)
	//for i := 0; i < 200; i++ {
	//	if l, ok := values.KeyLen[i]; ok {
	//		fmt.Printf("key len %d: %d\n", i, l)
	//	}
	//}
	//for i := 0; i < 1000; i++ {
	//	if l, ok := values.ValueLen[i]; ok {
	//		fmt.Printf("value len %d: %d\n", i, l)
	//	}
	//}
	return values
}

func createDb(prop *snapshot.FileProperties) {
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

	const persistEach = 10000

	var count int
	var errW error
	err = kvstream.Iterate(func(k []byte, v []byte) bool {
		st.KVStore().Set(kv.Key(k), v)
		count++
		if count%persistEach == 0 {
			if errW = st.Save(); errW != nil {
				return false
			}
			fmt.Printf("committed %d total records to database\n", count)
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
	fmt.Printf("committed %d total records to database\n", count)

	c := trie.RootCommitment(st.TrieAccess())
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
	fmt.Printf("veyfying database for chain ID %s\n", prop.ChainID)

	db, err := dbmanager.NewDB(dbDir)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	st := state.NewVirtualState(db.NewStore())
	c := trie.RootCommitment(st.TrieAccess())
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

	const reportEach = 10000
	var count int
	var errW error

	err = kvstream.Iterate(func(k, v []byte) bool {
		proof := state.CommitmentModel.Proof(k, st.TrieAccess())
		if errW = proof.Validate(c, v); errW != nil {
			return false
		}
		count++
		if count%reportEach == 0 {
			fmt.Printf("verified total %d records\n", count)
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
	fmt.Printf("verified total %d records\n", count)

	fmt.Printf("success: file %s match the database\n", prop.FileName)
}
