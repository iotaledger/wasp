// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/iotaledger/goshimmer/packages/database"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/database/dbkeys"
	"github.com/iotaledger/wasp/packages/iscp"
)

var dbKeysNames = map[byte]string{
	dbkeys.ObjectTypeDBSchemaVersion:    "Schema Version",
	dbkeys.ObjectTypeChainRecord:        "Chain Record",
	dbkeys.ObjectTypeCommitteeRecord:    "Committee Record",
	dbkeys.ObjectTypeDistributedKeyData: "Distributed Key Data",
	dbkeys.ObjectTypeStateHash:          "State Hash",
	dbkeys.ObjectTypeBlock:              "Block",
	dbkeys.ObjectTypeStateVariable:      "State Variable",
	dbkeys.ObjectTypeNodeIdentity:       "Node Identity",
	dbkeys.ObjectTypeBlobCache:          "BlobCache",
	dbkeys.ObjectTypeBlobCacheTTL:       "BlobCacheTTL",
	dbkeys.ObjectTypeTrustedPeer:        "TrustedPeer",
}

// TODO should be a cli param
const dbpath = "/tmp/wasp-cluster/wasp0/waspdb"

func printDbEntries(dbDir fs.DirEntry) {
	if !dbDir.IsDir() {
		fmt.Printf("Not a directory, skipping %s\n", dbDir.Name())
		return
	}
	db, err := database.NewDB(fmt.Sprintf("%s/%s", dbpath, dbDir.Name()))
	if err != nil {
		panic(err)
	}
	store := db.NewStore()

	fmt.Printf("\n\n------------------ %s ------------------\n", dbDir.Name())
	accLen := 0

	hnameUsedSpace := make(map[iscp.Hname]int)
	hnameCount := make(map[iscp.Hname]int)

	dbKeysUsedSpace := make(map[byte]int)

	err = store.Iterate(kvstore.EmptyPrefix, func(k kvstore.Key, v []byte) bool {
		usedSpace := len(k) + len(v)
		accLen += usedSpace
		dbKeysUsedSpace[k[0]] += usedSpace
		if len(k) >= 5 {
			hn, err := iscp.HnameFromBytes(k[1:5])
			if err == nil {
				fmt.Printf("HName: %s, key len: %d \t", hn, len(k))
				hnameUsedSpace[hn] += usedSpace
				hnameCount[hn]++
			}
		}
		fmt.Printf("Key: %s - Value len: %d\n", k, len(v))
		return true
	})

	fmt.Printf("\n\n Total DB size: %d\n\n", accLen)

	for hn, space := range hnameUsedSpace {
		fmt.Printf("Hname: %s, %d entries, size: %d\n", hn, hnameCount[hn], space)
	}

	fmt.Printf("\n\n DB Usage per dbKeys: \n\n")
	for dbkey, space := range dbKeysUsedSpace {
		fmt.Printf("KEY: %s, space: %d\n", dbKeysNames[dbkey], space)
	}

	if err != nil {
		panic(err)
	}
}

func main() {
	subDirectories, err := os.ReadDir(dbpath)
	if err != nil {
		panic(err)
	}
	for _, dir := range subDirectories {
		printDbEntries(dir)
	}
}
