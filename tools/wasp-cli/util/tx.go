package util

import (
	"context"
	"fmt"
	"os"
	"time"

	"fortio.org/safecast"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/clients/apiextensions"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/format"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
)

func WithOffLedgerRequest(ctx context.Context, client *apiclient.APIClient, f func() (isc.OffLedgerRequest, error)) {
	req, err := f()
	log.Check(err)
	reqID := req.ID().String()
	waitForCompletion := config.WaitForCompletion != config.DefaultWaitForCompletion

	data := map[string]interface{}{
		"request_id":          reqID,
		"wait_for_completion": waitForCompletion,
		"check_receipt_hint":  fmt.Sprintf("%s chain request %s", os.Args[0], reqID),
	}
	if waitForCompletion {
		data["wait_timeout"] = config.WaitForCompletion
	}
	log.Check(format.FormatSuccess("off_ledger_request", data)) //nolint:contextcheck
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
		LogReceipt(*receipt) //nolint:contextcheck
	}
}

func WithSCTransaction(ctx context.Context, client *apiclient.APIClient, f func() (*iotagraphql.ExecuteTransactionBlockResponse, error), forceWait ...time.Duration) *iotagraphql.ExecuteTransactionBlockResponse {
	tx, err := f()
	log.Check(err)
	ref, err := tx.ExecuteTransactionBlock.Effects.GetCreatedObjectByName(iscmove.RequestModuleName, iscmove.RequestObjectName)
	log.Check(err)
	reqID := ref.ObjectID.String()
	waitRequested := len(forceWait) > 0 || config.WaitForCompletion != config.DefaultWaitForCompletion
	waitDescription := ""
	if len(forceWait) > 0 {
		waitDescription = forceWait[0].String()
	} else if config.WaitForCompletion != config.DefaultWaitForCompletion {
		waitDescription = config.WaitForCompletion
	}

	data := map[string]interface{}{
		"transaction_digest":  tx.ExecuteTransactionBlock.Effects.TransactionBlock.Digest,
		"request_id":          reqID,
		"wait_for_completion": waitRequested,
	}
	if waitDescription != "" {
		data["wait_timeout"] = waitDescription
	}
	log.Check(format.FormatSuccess("on_ledger_transaction", data)) //nolint:contextcheck

	if len(forceWait) > 0 {
		_, err2 := apiextensions.APIWaitUntilAllRequestsProcessed(ctx, client, tx, true, forceWait[0])
		log.Check(err2)
	} else if config.WaitForCompletion != config.DefaultWaitForCompletion {
		timeout, err := time.ParseDuration(config.WaitForCompletion)
		log.Check(err)
		_, err2 := apiextensions.APIWaitUntilAllRequestsProcessed(ctx, client, tx, true, timeout)
		log.Check(err2)
	}

	return tx
}
