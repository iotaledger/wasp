package util

import (
	"context"
	"os"
	"time"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func WithOffLedgerRequest(ctx context.Context, client *apiclient.APIClient, chainID isc.ChainID, f func() (isc.OffLedgerRequest, error)) {
	req, err := f()
	log.Check(err)
	log.Printf("Posted off-ledger request (check result with: %s chain request %s)\n", os.Args[0], req.ID().String())
	if config.WaitForCompletion {
		receipt, _, err := client.ChainsAPI.
			WaitForRequest(ctx, chainID.String(), req.ID().String()).
			WaitForL1Confirmation(true).
			TimeoutSeconds(60).
			Execute()

		log.Check(err)
		LogReceipt(*receipt)
	}
}

func WithSCTransaction(ctx context.Context, client *apiclient.APIClient, chainID isc.ChainID, f func() (*iotajsonrpc.IotaTransactionBlockResponse, error), forceWait ...bool) *iotajsonrpc.IotaTransactionBlockResponse {
	tx, err := f()
	log.Check(err)
	log.Printf("Posted on-ledger transaction %s\n", tx.Digest)

	ref, err := tx.GetCreatedObjectInfo(iscmove.RequestModuleName, iscmove.RequestObjectName)
	log.Check(err)
	log.Printf("Request ID: %s\n", ref.ObjectID.String())

	if config.WaitForCompletion || len(forceWait) > 0 {
		log.Printf("Waiting for tx requests to be processed...\n")
		_, err2 := apiextensions.APIWaitUntilAllRequestsProcessed(ctx, client, chainID, tx, true, 1*time.Minute)
		log.Check(err2)
	}

	return tx
}
