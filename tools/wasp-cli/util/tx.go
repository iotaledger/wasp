package util

import (
	"os"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func PostTransaction(tx *ledgerstate.Transaction) {
	WithTransaction(func() (*ledgerstate.Transaction, error) {
		return tx, config.GoshimmerClient().PostTransaction(tx)
	})
}

func WithTransaction(f func() (*ledgerstate.Transaction, error)) *ledgerstate.Transaction {
	tx, err := f()
	log.Check(err)
	logTx(tx, nil)

	if config.WaitForCompletion {
		log.Check(config.GoshimmerClient().WaitForConfirmation(tx.ID()))
	}

	return tx
}

func WithOffLedgerRequest(chainID *iscp.ChainID, f func() (*request.OffLedger, error)) {
	req, err := f()
	log.Check(err)
	log.Printf("Posted off-ledger request (check result with: %s chain request %s)\n", os.Args[0], req.ID().Base58())
	if config.WaitForCompletion {
		log.Check(config.WaspClient().WaitUntilRequestProcessed(chainID, req.ID(), 1*time.Minute))
	}
}

func WithSCTransaction(chainID *iscp.ChainID, f func() (*ledgerstate.Transaction, error), forceWait ...bool) *ledgerstate.Transaction {
	tx, err := f()
	log.Check(err)
	logTx(tx, chainID)

	if config.WaitForCompletion || len(forceWait) > 0 {
		log.Printf("Waiting for tx requests to be processed...\n")
		log.Check(config.WaspClient().WaitUntilAllRequestsProcessed(*chainID, tx, 1*time.Minute))
	}

	return tx
}

func logTx(tx *ledgerstate.Transaction, chainID *iscp.ChainID) {
	var reqs []iscp.RequestID
	if chainID != nil {
		reqs = request.RequestsInTransaction(chainID, tx)
	}
	if len(reqs) == 0 {
		log.Printf("Posted on-ledger transaction %s\n", tx.ID().Base58())
	} else {
		plural := ""
		if len(reqs) != 1 {
			plural = "s"
		}
		log.Printf("Posted on-ledger transaction %s containing %d request%s:\n", tx.ID().Base58(), len(reqs), plural)
		for i, reqID := range reqs {
			log.Printf("  - #%d (check result with: %s chain request %s)\n", i, os.Args[0], reqID.Base58())
		}
	}
}
