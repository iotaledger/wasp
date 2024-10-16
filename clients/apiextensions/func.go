package apiextensions

import (
	"fmt"
	"math"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/isc"
)

func CallArgsToAPIArgs(args isc.CallArguments) [][]int32 {
	converted := make([][]int32, 0, len(args))

	for _, entry := range args {
		convertedEntry := make([]int32, 0, len(entry))

		for _, b := range entry {
			convertedEntry = append(convertedEntry, int32(b))
		}

		converted = append(converted, convertedEntry)
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

func APIArgsToCallArgs(res [][]int32) (isc.CallArguments, error) {
	converted := make([][]byte, 0, len(res))

	for i, entry := range res {
		convertedEntry := make([]byte, 0, len(entry))

		for j, b := range entry {
			if b < 0 || b > math.MaxUint8 {
				return nil, fmt.Errorf("invalid value of byte %v of API call data entry %v: %d", j, i, b)
			}

			convertedEntry = append(convertedEntry, byte(b))
		}

		converted = append(converted, convertedEntry)
	}

	return isc.CallArguments(converted), nil
}

func APIResultToCallArgs(res [][]int32) (isc.CallResults, error) {
	return APIArgsToCallArgs(res)
}

func APIWaitUntilAllRequestsProcessed(client *apiclient.APIClient, chainID isc.ChainID, tx *iotago.Transaction, waitForL1Confirmation bool, timeout time.Duration) ([]*apiclient.ReceiptResponse, error) {
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
