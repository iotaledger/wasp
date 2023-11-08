package util

import (
	"fmt"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
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

func LogReceipt(receipt apiclient.ReceiptResponse, index ...int) {
	req := receipt.Request

	kind := "on-ledger"
	if req.IsOffLedger {
		kind = "off-ledger"
	}

	args, err := apiextensions.APIJsonDictToDict(req.Params)
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
		{K: "Error", V: errMsg},
		{K: "Gas budget", V: receipt.GasBudget},
		{K: "Gas burned", V: receipt.GasBurned},
		{K: "Gas fee charged", V: receipt.GasFeeCharged},
		{K: "Storage deposit charged", V: receipt.StorageDepositCharged},
	}
	if len(index) > 0 {
		log.Printf("Request #%d (%s):\n", index[0], req.RequestId)
	} else {
		log.Printf("Request %s:\n", req.RequestId)
	}
	log.PrintTree(tree, 2, 2)
}
