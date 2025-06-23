package util

import (
	"context"
	"os"
	"time"

	"fortio.org/safecast"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func WithOffLedgerRequest(ctx context.Context, client *apiclient.APIClient, f func() (isc.OffLedgerRequest, error)) {
	req, err := f()
	log.Check(err)
	log.Printf("Posted off-ledger request (check result with: %s chain request %s)\n", os.Args[0], req.ID().String())
	if config.WaitForCompletion != config.DefaultWaitForCompletion {
		timeout, err := time.ParseDuration(config.WaitForCompletion)
		log.Check(err)
		timeoutSeconds, err := safecast.Convert[int32](timeout / time.Second)
		log.Check(err)
		receipt, _, err := client.ChainsAPI.
			WaitForRequest(ctx, req.ID().String()).
			WaitForL1Confirmation(true).
			TimeoutSeconds(timeoutSeconds).
			Execute()

		log.Check(err)
		LogReceipt(*receipt)
	}
}

func WithSCTransaction(ctx context.Context, client *apiclient.APIClient, f func() (*iotajsonrpc.IotaTransactionBlockResponse, error), forceWait ...time.Duration) *iotajsonrpc.IotaTransactionBlockResponse {
	tx, err := f()
	log.Check(err)
	log.Printf("Posted on-ledger transaction %s\n", tx.Digest)

	ref, err := tx.GetCreatedObjectInfo(iscmove.RequestModuleName, iscmove.RequestObjectName)
	log.Check(err)
	log.Printf("Request ID: %s\n", ref.ObjectID.String())

	if len(forceWait) > 0 {
		log.Printf("Waiting for tx requests to be processed...\n")
		_, err2 := apiextensions.APIWaitUntilAllRequestsProcessed(ctx, client, tx, true, forceWait[0])
		log.Check(err2)
	} else if config.WaitForCompletion != config.DefaultWaitForCompletion {
		log.Printf("Waiting for tx requests to be processed...\n")
		timeout, err := time.ParseDuration(config.WaitForCompletion)
		log.Check(err)
		_, err2 := apiextensions.APIWaitUntilAllRequestsProcessed(ctx, client, tx, true, timeout)
		log.Check(err2)
	}

	return tx
}
