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
			chain, err = defaultChainFallback(chain)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			bi, err := fetchBlockInfo(ctx, client, args)
			if err != nil {
				return err
			}
			log.Printf("Block index: %d\n", bi.BlockIndex)
			log.Printf("Timestamp: %s\n", bi.Timestamp.UTC().Format(time.RFC3339))
			log.Printf("Total requests: %d\n", bi.TotalRequests)
			log.Printf("Successful requests: %d\n", bi.NumSuccessfulRequests)
			log.Printf("Off-ledger requests: %d\n", bi.NumOffLedgerRequests)
			log.Printf("\n")
			if err := logRequestsInBlock(ctx, client, bi.BlockIndex); err != nil {
				return err
			}
			log.Printf("\n")
			if err := logEventsInBlock(ctx, client, bi.BlockIndex); err != nil {
				return err
			}
			return nil
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	return cmd
}

func fetchBlockInfo(ctx context.Context, client *apiclient.APIClient, args []string) (*apiclient.BlockInfoResponse, error) {
	if len(args) == 0 {
		blockInfo, _, err := client.
			CorecontractsAPI.
			BlocklogGetLatestBlockInfo(ctx).
			Execute() //nolint:bodyclose // false positive
		if err != nil {
			return nil, err
		}
		return blockInfo, nil
	}

	blockIndexStr := args[0]
	index, err := strconv.ParseUint(blockIndexStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid block index '%s': %w", blockIndexStr, err)
	}

	blockInfo, _, err := client.
		CorecontractsAPI.
		BlocklogGetBlockInfo(ctx, uint32(index)).
		Block(blockIndexStr).
		Execute() //nolint:bodyclose // false positive
	if err != nil {
		return nil, err
	}
	return blockInfo, nil
}

func logRequestsInBlock(ctx context.Context, client *apiclient.APIClient, index uint32) error {
	receipts, _, err := client.CorecontractsAPI.
		BlocklogGetRequestReceiptsOfBlock(ctx, index).
		Block(fmt.Sprintf("%d", index)).
		Execute() //nolint:bodyclose // false positive
	if err != nil {
		return err
	}

	for i, receipt := range receipts {
		r := receipt
		util.LogReceipt(r, i) //nolint:contextcheck
	}
	return nil
}

func logEventsInBlock(ctx context.Context, client *apiclient.APIClient, index uint32) error {
	events, _, err := client.CorecontractsAPI.
		BlocklogGetEventsOfBlock(ctx, index).
		Block(fmt.Sprintf("%d", index)).
		Execute() //nolint:bodyclose // false positive
	if err != nil {
		return err
	}
	logEvents(events)
	return nil
}

func hexLenFromByteLen(length int) int {
	return (length * 2) + 2
}

func reqIDFromString(s string) (isc.RequestID, error) {
	switch len(s) {
	case hexLenFromByteLen(iotago.AddressLen):
		// isc ReqID
		reqID, err := isc.RequestIDFromString(s)
		if err != nil {
			return isc.RequestID{}, fmt.Errorf("invalid isc requestID: %w", err)
		}
		return reqID, nil
	case hexLenFromByteLen(common.HashLength):
		bytes, err := cryptolib.DecodeHex(s)
		if err != nil {
			return isc.RequestID{}, fmt.Errorf("invalid evm tx hash: %w", err)
		}
		var txHash common.Hash
		copy(txHash[:], bytes)
		return isc.RequestIDFromEVMTxHash(txHash), nil
	default:
		return isc.RequestID{}, fmt.Errorf("invalid requestID length: %d", len(s))
	}
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
			chain, err = defaultChainFallback(chain)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			reqID, err := reqIDFromString(args[0])
			if err != nil {
				return err
			}

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
			if err := logEventsInRequest(ctx, client, reqID); err != nil {
				return err
			}
			log.Printf("\n")
			return nil
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	return cmd
}

func logEventsInRequest(ctx context.Context, client *apiclient.APIClient, reqID isc.RequestID) error {
	events, _, err := client.CorecontractsAPI.
		BlocklogGetEventsOfRequest(ctx, reqID.String()).
		Execute() //nolint:bodyclose // false positive
	if err != nil {
		return err
	}
	logEvents(events)
	return nil
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
