package main

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/database/dbmanager"
	"github.com/iotaledger/wasp/packages/iscp"
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
		values := scanFile(fname)
		createDb(values.ChainID)
	}
}

func scanFile(fname string) *snapshot.ScanValues {
	fmt.Printf("scaning snapshot file %s\n", fname)
	values, err := snapshot.ScanSnapshotForValues(fname)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Chain ID: %s\n", values.ChainID)
	fmt.Printf("State index: %d\n", values.StateIndex)
	fmt.Printf("Timestamp: %v\n", values.TimeStamp)
	fmt.Printf("Number of records: %d\n", values.NumRecords)
	fmt.Printf("Longest key: %d\n", values.MaxKeyLen)
	for i := 0; i < 200; i++ {
		if l, ok := values.KeyLen[i]; ok {
			fmt.Printf("key len %d: %d\n", i, l)
		}
	}
	for i := 0; i < 1000; i++ {
		if l, ok := values.ValueLen[i]; ok {
			fmt.Printf("value len %d: %d\n", i, l)
		}
	}
	return values
}

// TODO must come from config or CL
const dbDirectory = ""

func createDb(chainID *iscp.ChainID) {
	//log := logger.NewLogger("snapshot")
	//dbm := dbmanager.NewDBManager(log, false)

	dbDir := chainID.String()
	if _, err := os.Stat(dbDir); !os.IsNotExist(err) {
		fmt.Printf("directory %s already exists. Can't create new database\n", dbDir)
		os.Exit(1)
	}
	fmt.Printf("creating new database for chain ID %s\n", dbDir)
	_, err := dbmanager.NewDB(dbDir)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
