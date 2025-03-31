// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/iotaledger/hive.go/kvstore"
	hivedb "github.com/iotaledger/hive.go/kvstore/database"
	"github.com/iotaledger/hive.go/kvstore/rocksdb"
	"github.com/samber/lo"

	old_kvstore "github.com/iotaledger/hive.go/kvstore"
	old_database "github.com/nnikolash/wasp-types-exported/packages/database"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("usage: %s <db-1-dir> <db-2-dir>", os.Args[0])
	}

	db1Dir := os.Args[1]
	db2Dir := os.Args[2]

	db1Dir = lo.Must(filepath.Abs(db1Dir))
	db2Dir = lo.Must(filepath.Abs(db2Dir))

	db1KVS := ConnectToDB(db1Dir)
	db2KVS := ConnectToDB(db2Dir)

	fmt.Println("Comparing DB1 vs DB2...")
	printProgress := NewProgressPrinter()
	count := 0

	db1KVS.Iterate(nil, func(key, value1 []byte) bool {
		count++
		printProgress.Print()

		value2, err := db2KVS.Get(key)
		if err != nil {
			if errors.Is(err, kvstore.ErrKeyNotFound) {
				log.Printf("Key %x not found in DB2\n", key)
				return true
			}
			panic(err)
		}
		if !bytes.Equal(value1, value2) {
			log.Printf("Key %x has different values in DB1 and DB2: %x != %x", key, value1, value2)
			return true
		}

		return true
	})

	fmt.Printf("\nTotal keys compared: %v\n", count)

	fmt.Println("Comparing DB2 vs DB1...")
	count = 0
	printProgress = NewProgressPrinter()

	db2KVS.Iterate(nil, func(key, value2 []byte) bool {
		count++
		printProgress.Print()

		value1, err := db1KVS.Get(key)
		if err != nil {
			if errors.Is(err, kvstore.ErrKeyNotFound) {
				log.Printf("Key %x not found in DB1\n", key)
				return true
			}
			panic(err)
		}
		if !bytes.Equal(value1, value2) {
			log.Printf("Key %x has different values in DB1 and DB2: %x != %x", key, value1, value2)
			return true
		}

		return true
	})

	fmt.Printf("\nTotal keys compared: %v\n", count)
}

func ConnectToDB(dbDir string) old_kvstore.KVStore {
	log.Printf("Connecting to DB in %v\n", dbDir)

	rocksDatabase := lo.Must(rocksdb.OpenDBReadOnly(dbDir,
		rocksdb.IncreaseParallelism(runtime.NumCPU()-1),
		rocksdb.Custom([]string{
			"periodic_compaction_seconds=43200",
			"level_compaction_dynamic_level_bytes=true",
			"keep_log_file_num=2",
			"max_log_file_size=50000000", // 50MB per log file
		}),
	))

	db := old_database.New(
		dbDir,
		rocksdb.New(rocksDatabase),
		hivedb.EngineRocksDB,
		true,
		func() bool { panic("should not be called") },
	)

	kvs := db.KVStore()

	return kvs
}

func NewProgressPrinter(printingPeriod ...uint32) *ProgressPrinter {
	var period uint32 = 100000
	if len(printingPeriod) > 0 {
		period = printingPeriod[0]
	}

	return &ProgressPrinter{period: period}
}

type ProgressPrinter struct {
	Count  uint32
	period uint32
}

func (p *ProgressPrinter) Print() {
	p.Count++
	if p.Count%p.period == 0 {
		fmt.Printf("\rProcessed: %v         ", p.Count)
	}
}
