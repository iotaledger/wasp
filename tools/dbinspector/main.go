// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io/fs"
	"os"

	hivedb "github.com/iotaledger/hive.go/core/database"
	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/wasp/packages/common"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/isc"
)

var dbKeysNames = map[byte]string{
	common.ObjectTypeDBSchemaVersion:    "Schema Version",
	common.ObjectTypeChainRecord:        "Chain Record",
	common.ObjectTypeCommitteeRecord:    "Committee Record",
	common.ObjectTypeDistributedKeyData: "Distributed Key Data",
	common.ObjectTypeTrie:               "State Hash",
	common.ObjectTypeBlock:              "Block",
	common.ObjectTypeState:              "State Variable",
	common.ObjectTypeNodeIdentity:       "Node Identity",
	common.ObjectTypeBlobCache:          "BlobCache",
	common.ObjectTypeBlobCacheTTL:       "BlobCacheTTL",
	common.ObjectTypeTrustedPeer:        "TrustedPeer",
}

const defaultDbpath = "/tmp/wasp-cluster/wasp0/waspdb"

func printDbEntries(dbDir fs.DirEntry, dbpath string) {
	if !dbDir.IsDir() {
		fmt.Printf("Not a directory, skipping %s\n", dbDir.Name())
		return
	}

	db, err := database.DatabaseWithDefaultSettings(fmt.Sprintf("%s/%s", dbpath, dbDir.Name()), false, hivedb.EngineAuto, false)
	if err != nil {
		panic(err)
	}

	store := db.KVStore()

	fmt.Printf("\n\n------------------ %s ------------------\n", dbDir.Name())
	accLen := 0

	hnameUsedSpace := make(map[isc.Hname]int)
	hnameCount := make(map[isc.Hname]int)

	dbKeysUsedSpace := make(map[byte]int)

	err = store.Iterate(kvstore.EmptyPrefix, func(k kvstore.Key, v []byte) bool {
		usedSpace := len(k) + len(v)
		accLen += usedSpace
		dbKeysUsedSpace[k[0]] += usedSpace
		if len(k) >= 5 {
			hn, err := isc.HnameFromBytes(k[1:5])
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
	var dbpath string
	if len(os.Args) > 1 {
		dbpath = os.Args[1]
	} else {
		dbpath = defaultDbpath
	}
	subDirectories, err := os.ReadDir(dbpath)
	if err != nil {
		panic(err)
	}
	for _, dir := range subDirectories {
		printDbEntries(dir, dbpath)
	}
}
