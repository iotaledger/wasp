package apiextensions

import (
	"context"
	"fmt"
	"time"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
)

func CallArgsToAPIArgs(args isc.CallArguments) []string {
	converted := make([]string, 0, len(args))

	for _, entry := range args {
		converted = append(converted, cryptolib.EncodeHex(entry))
	}

	return converted
}

func CallViewReq(msg isc.Message) apiclient.ContractCallViewRequest {
	return apiclient.ContractCallViewRequest{
		ContractHName: msg.Target.Contract.String(),
		FunctionHName: msg.Target.EntryPoint.String(),
		Arguments:     CallArgsToAPIArgs(msg.Params),
	}
}

func APIArgsToCallArgs(res []string) (isc.CallArguments, error) {
	converted := make([][]byte, 0, len(res))

	for i, entry := range res {
		convertedEntry, err := cryptolib.DecodeHex(entry)
		if err != nil {
			return nil, fmt.Errorf("error decoding hex string of entry %d: %w", i, err)
		}

		converted = append(converted, convertedEntry)
	}

	return isc.CallArguments(converted), nil
}

func APIResultToCallArgs(res []string) (isc.CallResults, error) {
	return APIArgsToCallArgs(res)
}

func APIWaitUntilAllRequestsProcessed(ctx context.Context, client *apiclient.APIClient, tx *iotajsonrpc.IotaTransactionBlockResponse, waitForL1Confirmation bool, timeout time.Duration) ([]*apiclient.ReceiptResponse, error) {
	req, err := tx.GetCreatedObjectInfo(iscmove.RequestModuleName, iscmove.RequestObjectName)
	if err != nil {
		return nil, err
	}

	// TODO: In theory we can pass multiple requests into a PTB call, but we don't right now.
	// For now fake old behavior and just create an array with a length of 1

	reqs := []isc.RequestID{isc.RequestID(*req.ObjectID)}

	ret := make([]*apiclient.ReceiptResponse, len(reqs))
	for i, req := range reqs {
		receipt, _, err := client.ChainsAPI.
			WaitForRequest(ctx, req.String()).
			TimeoutSeconds(int32(timeout.Seconds())).
			WaitForL1Confirmation(waitForL1Confirmation).
			Execute()
		if err != nil {
			return nil, fmt.Errorf("error in WaitForRequest, reqID=%v: %w", req.String(), err)
		}

		ret[i] = receipt
	}
	return ret, nil
}
