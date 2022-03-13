package main

import (
	"flag"
	"fmt"
	"github.com/iotaledger/wasp/packages/snapshot"
	"os"
)

func main() {
	fileToScan := flag.String("scan", "", "scan the snapshot file")
	flag.Parse()

	if *fileToScan == "" {
		fmt.Printf("USAGE: snapshot -scan <filename>")
		os.Exit(1)
	}
	values, err := snapshot.ScanSnapshotForValues(*fileToScan)
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
}
