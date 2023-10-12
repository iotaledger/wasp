package util

import (
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

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

	tree := []log.TreeItem{
		{K: "Kind", V: kind},
		{K: "Sender", V: req.SenderAccount},
		{K: "Contract Hname", V: req.CallTarget.ContractHName},
		{K: "Function Hname", V: req.CallTarget.FunctionHName},
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
