package util

import (
	"context"
	"os"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func WithOffLedgerRequest(chainID isc.ChainID, f func() (isc.OffLedgerRequest, error)) {
	req, err := f()
	log.Check(err)
	log.Printf("Posted off-ledger request (check result with: %s chain request %s)\n", os.Args[0], req.ID().String())
	if config.WaitForCompletion {
		_, _, err = cliclients.WaspClient().RequestsApi.
			WaitForRequest(context.Background(), chainID.String(), req.ID().String()).
			TimeoutSeconds(60).
			Execute()

		log.Check(err)
	}
	// TODO print receipt?
}

func WithSCTransaction(chainID isc.ChainID, f func() (*iotago.Transaction, error), forceWait ...bool) *iotago.Transaction {
	tx, err := f()
	log.Check(err)
	logTx(chainID, tx)

	if config.WaitForCompletion || len(forceWait) > 0 {
		log.Printf("Waiting for tx requests to be processed...\n")
		client := cliclients.WaspClient()
		_, err := apiextensions.APIWaitUntilAllRequestsProcessed(client, chainID, tx, 1*time.Minute)
		log.Check(err)
	}

	return tx
}

func logTx(chainID isc.ChainID, tx *iotago.Transaction) {
	allReqs, err := isc.RequestsInTransaction(tx)
	log.Check(err)
	txid, err := tx.ID()
	log.Check(err)
	reqs := allReqs[chainID]
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
