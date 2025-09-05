package chain

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/util"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/waspcmd"
)

func initBlockCmd() *cobra.Command {
	var node string
	var chain string
	cmd := &cobra.Command{
		Use:   "block [index]",
		Short: "Get information about a block given its index, or latest block if missing",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			node, err = waspcmd.DefaultWaspNodeFallback(node)
			if err != nil {
				return err
			}
			chain = defaultChainFallback(chain)
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			bi := fetchBlockInfo(ctx, client, args)
			log.Printf("Block index: %d\n", bi.BlockIndex)
			log.Printf("Timestamp: %s\n", bi.Timestamp.UTC().Format(time.RFC3339))
			log.Printf("Total requests: %d\n", bi.TotalRequests)
			log.Printf("Successful requests: %d\n", bi.NumSuccessfulRequests)
			log.Printf("Off-ledger requests: %d\n", bi.NumOffLedgerRequests)
			log.Printf("\n")
			logRequestsInBlock(ctx, client, bi.BlockIndex)
			log.Printf("\n")
			logEventsInBlock(ctx, client, bi.BlockIndex)
			return nil
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	return cmd
}

func fetchBlockInfo(ctx context.Context, client *apiclient.APIClient, args []string) *apiclient.BlockInfoResponse {
	if len(args) == 0 {
		blockInfo, _, err := client.
			CorecontractsAPI.
			BlocklogGetLatestBlockInfo(ctx).
			Execute() //nolint:bodyclose // false positive

		log.Check(err)
		return blockInfo
	}

	blockIndexStr := args[0]
	index, err := strconv.ParseUint(blockIndexStr, 10, 32)
	log.Check(err)

	blockInfo, _, err := client.
		CorecontractsAPI.
		BlocklogGetBlockInfo(ctx, uint32(index)).
		Block(blockIndexStr).
		Execute() //nolint:bodyclose // false positive

	log.Check(err)
	return blockInfo
}

func logRequestsInBlock(ctx context.Context, client *apiclient.APIClient, index uint32) {
	receipts, _, err := client.CorecontractsAPI.
		BlocklogGetRequestReceiptsOfBlock(ctx, index).
		Block(fmt.Sprintf("%d", index)).
		Execute() //nolint:bodyclose // false positive

	log.Check(err)

	for i, receipt := range receipts {
		r := receipt
		util.LogReceipt(r, i)
	}
}

func logEventsInBlock(ctx context.Context, client *apiclient.APIClient, index uint32) {
	events, _, err := client.CorecontractsAPI.
		BlocklogGetEventsOfBlock(ctx, index).
		Block(fmt.Sprintf("%d", index)).
		Execute() //nolint:bodyclose // false positive

	log.Check(err)
	logEvents(events)
}

func hexLenFromByteLen(length int) int {
	return (length * 2) + 2
}

func reqIDFromString(s string) isc.RequestID {
	switch len(s) {
	case hexLenFromByteLen(iotago.AddressLen):
		// isc ReqID
		reqID, err := isc.RequestIDFromString(s)
		log.Check(err)
		return reqID
	case hexLenFromByteLen(common.HashLength):
		bytes, err := cryptolib.DecodeHex(s)
		log.Check(err)
		var txHash common.Hash
		copy(txHash[:], bytes)
		return isc.RequestIDFromEVMTxHash(txHash)
	default:
		log.Fatalf("invalid requestID length: %d", len(s))
	}
	panic("unreachable")
}

func initRequestCmd() *cobra.Command {
	var node string
	var chain string
	cmd := &cobra.Command{
		Use:   "request <request-id>",
		Short: "Get information about a request given its ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			node, err = waspcmd.DefaultWaspNodeFallback(node)
			if err != nil {
				return err
			}
			chain = defaultChainFallback(chain)
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			reqID := reqIDFromString(args[0])

			// TODO add optional block param?
			receipt, _, err := client.ChainsAPI.
				GetReceipt(ctx, reqID.String()).
				Execute() //nolint:bodyclose // false positive

			if err != nil {
				return err
			}

			log.Printf("Request found in block %d\n\n", receipt.BlockIndex)
			util.LogReceipt(*receipt)

			log.Printf("\n")
			logEventsInRequest(ctx, client, reqID)
			log.Printf("\n")
			return nil
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	return cmd
}

func logEventsInRequest(ctx context.Context, client *apiclient.APIClient, reqID isc.RequestID) {
	events, _, err := client.CorecontractsAPI.
		BlocklogGetEventsOfRequest(ctx, reqID.String()).
		Execute() //nolint:bodyclose // false positive

	log.Check(err)
	logEvents(events)
}

func logEvents(ret *apiclient.EventsResponse) {
	header := []string{"event"}
	rows := make([][]string, len(ret.Events))

	for i, event := range ret.Events {
		rows[i] = []string{event.Topic}
	}

	log.Printf("Total %d events\n", len(ret.Events))
	log.PrintTable(header, rows)
}
