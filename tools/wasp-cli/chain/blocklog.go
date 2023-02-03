package chain

import (
	"context"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func initBlockCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "block [index]",
		Short: "Get information about a block given its index, or latest block if missing",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			bi := fetchBlockInfo(args)
			log.Printf("Block index: %d\n", bi.BlockIndex)
			log.Printf("Timestamp: %s\n", bi.Timestamp.UTC().Format(time.RFC3339))
			log.Printf("Total requests: %d\n", bi.TotalRequests)
			log.Printf("Successful requests: %d\n", bi.NumSuccessfulRequests)
			log.Printf("Off-ledger requests: %d\n", bi.NumOffLedgerRequests)
			log.Printf("\n")
			logRequestsInBlock(bi.BlockIndex)
			log.Printf("\n")
			logEventsInBlock(bi.BlockIndex)
		},
	}
}

func fetchBlockInfo(args []string) *apiclient.BlockInfoResponse {
	client := cliclients.WaspClientForIndex()

	if len(args) == 0 {
		blockInfo, _, err := client.
			CorecontractsApi.
			BlocklogGetLatestBlockInfo(context.Background(), config.GetCurrentChainID().String()).
			Execute()

		log.Check(err)
		return blockInfo
	}

	index, err := strconv.Atoi(args[0])
	log.Check(err)

	blockInfo, _, err := client.
		CorecontractsApi.
		BlocklogGetBlockInfo(context.Background(), config.GetCurrentChainID().String(), uint32(index)).
		Execute()

	log.Check(err)

	return blockInfo
}

func logRequestsInBlock(index uint32) {
	client := cliclients.WaspClientForIndex()
	// TODO: It should be uint32 here. Fix generator!
	receipts, _, err := client.CorecontractsApi.
		BlocklogGetRequestReceiptsOfBlock(context.Background(), config.GetCurrentChainID().String(), index).
		Execute()

	log.Check(err)

	for i, receipt := range receipts.Receipts {
		logReceipt(&receipt, i)
	}
}

func logReceipt(receipt *apiclient.RequestReceiptResponse, index ...int) {
	req := receipt.Request

	kind := "on-ledger"
	if req.IsOffLedger {
		kind = "off-ledger"
	}

	args := req.Params.Items

	var argsTree interface{} = "(empty)"
	if len(args) > 0 {
		argsTree = args
	}

	errMsg := "(empty)"
	if receipt.Error != nil {
		errMsg = receipt.Error.ErrorMessage
	}

	tree := []log.TreeItem{
		{K: "Kind", V: kind},
		{K: "Sender", V: req.SenderAccount},
		{K: "Contract Hname", V: isc.Hname(req.CallTarget.Contract).String()},
		{K: "Entry point", V: isc.Hname(req.CallTarget.EntryPoint).String()},
		{K: "Arguments", V: argsTree},
		{K: "Error", V: errMsg},
	}
	if len(index) > 0 {
		log.Printf("Request #%d (%s):\n", index[0], req.RequestId)
	} else {
		log.Printf("Request %s:\n", req.RequestId)
	}
	log.PrintTree(tree, 2, 2)
}

func logEventsInBlock(index uint32) {
	client := cliclients.WaspClientForIndex()
	// TODO: It should be uint32 here. Fix generator!
	events, _, err := client.CorecontractsApi.
		BlocklogGetEventsOfBlock(context.Background(), config.GetCurrentChainID().String(), index).
		Execute()

	log.Check(err)
	logEvents(events)
}

func initRequestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "request <request-id>",
		Short: "Get information about a request given its ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			reqID, err := isc.RequestIDFromString(args[0])
			log.Check(err)

			client := cliclients.WaspClientForIndex()
			receipt, _, err := client.CorecontractsApi.
				BlocklogGetRequestReceipt(context.Background(), config.GetCurrentChainID().String(), reqID.String()).
				Execute()

			log.Check(err)

			log.Printf("Request found in block %d\n\n", receipt.BlockIndex)
			logReceipt(receipt)

			log.Printf("\n")
			logEventsInRequest(reqID)
			log.Printf("\n")
		},
	}
}

func logEventsInRequest(reqID isc.RequestID) {
	client := cliclients.WaspClientForIndex()
	// TODO: It should be uint32 here. Fix generator!
	events, _, err := client.CorecontractsApi.
		BlocklogGetEventsOfRequest(context.Background(), config.GetCurrentChainID().String(), reqID.String()).
		Execute()

	log.Check(err)
	logEvents(events)
}

func logEvents(ret *apiclient.EventsResponse) {
	header := []string{"event"}
	rows := make([][]string, len(ret.Events))

	for i, event := range ret.Events {
		rows[i] = []string{event}
	}

	log.Printf("Total %d events\n", len(ret.Events))
	log.PrintTable(header, rows)
}
