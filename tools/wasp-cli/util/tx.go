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
	config.L1Client().PostTx(tx)
}

func WithOffLedgerRequest(chainID *iscp.ChainID, f func() (iscp.OffLedgerRequest, error)) {
	req, err := f()
	log.Check(err)
	log.Printf("Posted off-ledger request (check result with: %s chain request %s)\n", os.Args[0], req.ID().String())
	if config.WaitForCompletion {
		_, err = config.WaspClient().WaitUntilRequestProcessed(chainID, req.ID(), 1*time.Minute)
		log.Check(err)
	}
	// TODO print receipt?
}

func WithSCTransaction(chainID *iscp.ChainID, f func() (*iotago.Transaction, error), forceWait ...bool) *iotago.Transaction {
	tx, err := f()
	log.Check(err)
	logTx(chainID, tx)

	if config.WaitForCompletion || len(forceWait) > 0 {
		log.Printf("Waiting for tx requests to be processed...\n")
		_, err := config.WaspClient().WaitUntilAllRequestsProcessed(chainID, tx, 1*time.Minute)
		log.Check(err)
	}

	return tx
}

func logTx(chainID *iscp.ChainID, tx *iotago.Transaction) {
	allReqs, err := iscp.RequestsInTransaction(tx)
	log.Check(err)
	txid, err := tx.ID()
	log.Check(err)
	reqs := allReqs[*chainID]
	if len(reqs) == 0 {
		log.Printf("Posted on-ledger transaction %s\n", txid.ToHex())
	} else {
		plural := ""
		if len(reqs) != 1 {
			plural = "s"
		}
		log.Printf("Posted on-ledger transaction %s containing %d request%s:\n", txid.ToHex(), len(reqs), plural)
		for i, req := range reqs {
			log.Printf("  - #%d (check result with: %s chain request %s)\n", i, os.Args[0], req.ID().String())
		}
	}
}
