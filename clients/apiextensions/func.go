package apiextensions

import (
	"fmt"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/isc"
)

func CallArgsToAPIArgs(args isc.CallArguments) []string {
	converted := make([]string, 0, len(args))

	for _, entry := range args {
		converted = append(converted, iotago.EncodeHex(entry))
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
		convertedEntry, err := iotago.DecodeHex(entry)
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

func APIWaitUntilAllRequestsProcessed(client *apiclient.APIClient, chainID isc.ChainID, tx iotajsonrpc.ParsedTransactionResponse, waitForL1Confirmation bool, timeout time.Duration) ([]*apiclient.ReceiptResponse, error) {
	panic("refactor me: APIWaitUntilAllRequestsProcessed")
	/*reqs, err := isc.Requests(tx)
	if err != nil {
		return nil, err
	}
	ret := make([]*apiclient.ReceiptResponse, len(reqs))
	for i, req := range reqs[chainID] {
		receipt, _, err := client.ChainsApi.
			WaitForRequest(context.Background(), chainID.String(), req.ID().String()).
			TimeoutSeconds(int32(timeout.Seconds())).
			WaitForL1Confirmation(waitForL1Confirmation).
			Execute()
		if err != nil {
			return nil, fmt.Errorf("Error in WaitForRequest, reqID=%v: %w", req.ID(), err)
		}

		ret[i] = receipt
	}
	return ret, nil*/
}
