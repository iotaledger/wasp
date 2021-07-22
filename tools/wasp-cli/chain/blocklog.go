package chain

import (
	"fmt"
	"strconv"
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

func blockCmd() *cobra.Command {
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
			logRequestsInBlock(bi.BlockIndex)
		},
	}
}

func fetchBlockInfo(args []string) *blocklog.BlockInfo {
	if len(args) == 0 {
		ret, err := SCClient(blocklog.Contract.Hname()).CallView(blocklog.FuncGetLatestBlockInfo.Name, nil)
		log.Check(err)
		index, _, err := codec.DecodeUint32(ret.MustGet(blocklog.ParamBlockIndex))
		log.Check(err)
		b, err := blocklog.BlockInfoFromBytes(index, ret.MustGet(blocklog.ParamBlockInfo))
		log.Check(err)
		return b
	}
	index, err := strconv.Atoi(args[0])
	log.Check(err)
	ret, err := SCClient(blocklog.Contract.Hname()).CallView(blocklog.FuncGetBlockInfo.Name, dict.Dict{
		blocklog.ParamBlockIndex: codec.EncodeUint32(uint32(index)),
	})
	log.Check(err)
	b, err := blocklog.BlockInfoFromBytes(uint32(index), ret.MustGet(blocklog.ParamBlockInfo))
	log.Check(err)
	return b
}

func logRequestsInBlock(index uint32) {
	ret, err := SCClient(blocklog.Contract.Hname()).CallView(blocklog.FuncGetRequestReceiptsForBlock.Name, dict.Dict{
		blocklog.ParamBlockIndex: codec.EncodeUint32(index),
	})
	log.Check(err)
	arr := collections.NewArray16ReadOnly(ret, blocklog.ParamRequestRecord)
	header := []string{"request ID", "kind", "log"}
	var rows [][]string
	for i := uint16(0); i < arr.MustLen(); i++ {
		req, err := blocklog.RequestReceiptFromBytes(arr.MustGetAt(i))
		log.Check(err)

		kind := "on-ledger"
		if req.OffLedger {
			kind = "off-ledger"
		}

		rows = append(rows, []string{
			req.RequestID.Base58(),
			kind,
			fmt.Sprintf("%q", string(req.Error)),
		})
	}
	log.Printf("Total %d requests\n", arr.MustLen())
	log.PrintTable(header, rows)
}

func requestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "request <request-id>",
		Short: "Get information about a request given its ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			reqID, err := iscp.RequestIDFromBase58(args[0])
			log.Check(err)
			ret, err := SCClient(blocklog.Contract.Hname()).CallView(blocklog.FuncGetRequestReceipt.Name, dict.Dict{
				blocklog.ParamRequestID: codec.EncodeRequestID(reqID),
			})
			log.Check(err)

			blockIndex, _, err := codec.DecodeUint32(ret.MustGet(blocklog.ParamBlockIndex))
			log.Check(err)
			req, err := blocklog.RequestReceiptFromBytes(ret.MustGet(blocklog.ParamRequestRecord))
			log.Check(err)

			kind := "on-ledger"
			if req.OffLedger {
				kind = "off-ledger"
			}

			log.Printf("%s request %s in block %d\n", kind, reqID.Base58(), blockIndex)
			log.Printf("Log: %q\n", string(req.Error))
		},
	}
}
