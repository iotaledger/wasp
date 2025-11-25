package util

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/clients/apiextensions"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/format"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
)

var (
	knownContractHnames  = map[string]string{}
	knownFunctionsHnames = map[string]string{}
)

func init() {
	// fill known hnames for contracts/functions so we can print them as humanly readable
	for hn, contract := range corecontracts.All {
		knownContractHnames[hn.String()] = contract.Name
	}
	for _, proc := range coreprocessors.All {
		for hn, handler := range proc.Entrypoints() {
			knownFunctionsHnames[hn.String()] = handler.Name()
		}
	}
}

func decodeKnownContractCall(req apiclient.RequestJSON) []log.TreeItem {
	contract, _ := isc.HnameFromString(req.CallTarget.ContractHName)
	entrypoint, _ := isc.HnameFromString(req.CallTarget.FunctionHName)

	// This just tries to decode the most common contract calls.
	// This is not a complete solution, but makes current development easier.

	if contract == accounts.Contract.Hname() && entrypoint == accounts.FuncTransferAllowanceTo.Hname() {
		params, err := hexutil.Decode(req.Params[0])
		if err != nil {
			fmt.Println("failed to decode params for FuncTransferAllowanceTo")
			return []log.TreeItem{}
		}

		arg1, err := accounts.FuncTransferAllowanceTo.Input1.Decode(params)
		if err != nil {
			fmt.Println("failed to decode params for FuncTransferAllowanceTo")
			return []log.TreeItem{}
		}

		return []log.TreeItem{{K: "Param 1", V: arg1.String()}}
	}

	return []log.TreeItem{}
}

func LogReceipt(receipt apiclient.ReceiptResponse, index ...int) {
	req := receipt.Request

	kind := "on-ledger"
	if req.IsOffLedger {
		kind = "off-ledger"
	}

	args, err := apiextensions.APIResultToCallArgs(req.Params)
	log.Check(err)

	errMsg := "(empty)"
	if receipt.ErrorMessage != nil {
		errMsg = *receipt.ErrorMessage
	}

	contractStr := req.CallTarget.ContractHName
	if contractName, ok := knownContractHnames[contractStr]; ok {
		contractStr = fmt.Sprintf("%s (%s)", contractStr, contractName)
	}

	funcStr := req.CallTarget.FunctionHName
	if funcName, ok := knownFunctionsHnames[funcStr]; ok {
		funcStr = fmt.Sprintf("%s (%s)", funcStr, funcName)
	}

	assets := make([]format.ChainReceiptAsset, 0, len(receipt.Request.Assets.Coins))
	for _, coin := range receipt.Request.Assets.Coins {
		assets = append(assets, format.ChainReceiptAsset{
			CoinType: coin.CoinType,
			Balance:  coin.Balance,
		})
	}

	output := format.ChainReceiptOutput{
		RequestID:          req.RequestId,
		Kind:               kind,
		Sender:             req.SenderAccount,
		ContractHName:      contractStr,
		FunctionHName:      funcStr,
		ParamsHex:          req.Params,
		ArgumentsRaw:       args,
		DecodedKnownParams: decodeKnownContractCall(req),
		Error:              errMsg,
		GasBudget:          receipt.GasBudget,
		GasBurned:          receipt.GasBurned,
		GasFeeCharged:      receipt.GasFeeCharged,
		StorageDeposit:     receipt.StorageDepositCharged,
		Assets:             assets,
	}
	if len(index) > 0 {
		idx := index[0]
		output.Index = &idx
	}
	log.Check(format.FormatSuccess("chain_receipt", output.ToMap()))
}
