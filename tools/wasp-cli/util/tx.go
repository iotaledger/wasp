package util

import (
	"os"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func PostTransaction(tx *iotago.Transaction) {
	WithTransaction(func() (*iotago.Transaction, error) {
		panic("TODO implement")
		// return tx, config.GoshimmerClient().PostTransaction(tx)
	})
}

func WithTransaction(f func() (*iotago.Transaction, error)) *iotago.Transaction {
	tx, err := f()
	log.Check(err)
	logTx(tx, nil)

	if config.WaitForCompletion {
		panic("TODO implement")
		// log.Check(config.GoshimmerClient().WaitForConfirmation(tx.ID()))
	}

	return tx
}

func WithOffLedgerRequest(chainID *iscp.ChainID, f func() (*iscp.OffLedgerRequestData, error)) {
	req, err := f()
	log.Check(err)
	log.Printf("Posted off-ledger request (check result with: %s chain request %s)\n", os.Args[0], req.ID().Base58())
	if config.WaitForCompletion {
		log.Check(config.WaspClient().WaitUntilRequestProcessed(chainID, req.ID(), 1*time.Minute))
	}
}

func WithSCTransaction(chainID *iscp.ChainID, f func() (*iotago.Transaction, error), forceWait ...bool) *iotago.Transaction {
	tx, err := f()
	log.Check(err)
	logTx(tx, chainID)

	if config.WaitForCompletion || len(forceWait) > 0 {
		log.Printf("Waiting for tx requests to be processed...\n")
		log.Check(config.WaspClient().WaitUntilAllRequestsProcessed(chainID, tx, 1*time.Minute))
	}

	return tx
}

func logTx(tx *iotago.Transaction, chainID *iscp.ChainID) {
	panic("TODO implement")
	// var reqs []iscp.RequestID
	// if chainID != nil {
	// 	reqs = request.RequestsInTransaction(chainID, tx)
	// }
	// if len(reqs) == 0 {
	// 	log.Printf("Posted on-ledger transaction %s\n", tx.ID().Base58())
	// } else {
	// 	plural := ""
	// 	if len(reqs) != 1 {
	// 		plural = "s"
	// 	}
	// 	log.Printf("Posted on-ledger transaction %s containing %d request%s:\n", tx.ID().Base58(), len(reqs), plural)
	// 	for i, reqID := range reqs {
	// 		log.Printf("  - #%d (check result with: %s chain request %s)\n", i, os.Args[0], reqID.Base58())
	// 	}
	// }
}
