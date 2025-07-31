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

		return []log.TreeItem{
			{K: "Param 1", V: arg1.String()},
		}
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

	var argsTree interface{} = "(empty)"
	if len(args) > 0 {
		argsTree = args
	}

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

	tree := []log.TreeItem{
		{K: "Kind", V: kind},
		{K: "Sender", V: req.SenderAccount},
		{K: "Contract Hname", V: contractStr},
		{K: "Function Hname", V: funcStr},
		{K: "Arguments", V: argsTree},
	}

	tree = append(tree, decodeKnownContractCall(req)...)

	coinsString := ""

	for _, coin := range receipt.Request.Assets.Coins {
		coinsString += fmt.Sprintf("	%s (%s)\n", coin.CoinType, coin.Balance)
	}

	treeRest := []log.TreeItem{
		{K: "Error", V: errMsg},
		{K: "Gas budget", V: receipt.GasBudget},
		{K: "Gas burned", V: receipt.GasBurned},
		{K: "Gas fee charged", V: receipt.GasFeeCharged},
		{K: "Storage deposit charged", V: receipt.StorageDepositCharged},
		{K: "Assets", V: coinsString},
	}

	tree = append(tree, treeRest...)

	if len(index) > 0 {
		log.Printf("Request #%d (%s):\n", index[0], req.RequestId)
	} else {
		log.Printf("Request %s:\n", req.RequestId)
	}
	log.PrintTree(tree, 2, 2)
}
