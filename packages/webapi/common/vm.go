package common

import (
	"fmt"
	"strconv"
	"strings"

	iotago "github.com/iotaledger/iota.go/v3"
	chainpkg "github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chainutil"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

func ParseReceipt(chain chainpkg.Chain, receipt *blocklog.RequestReceipt) (*isc.Receipt, error) {
	resolvedReceiptErr, err := chainutil.ResolveError(chain, receipt.Error)
	if err != nil {
		return nil, err
	}

	iscReceipt := receipt.ToISCReceipt(resolvedReceiptErr)

	return iscReceipt, nil
}

func CallView(ch chainpkg.Chain, contractName, functionName isc.Hname, params dict.Dict, blockIndexOrHash string) (dict.Dict, error) {
	var chainState state.State
	var err error
	switch {
	case blockIndexOrHash == "":
		chainState, err = ch.LatestState(chainpkg.ActiveOrCommittedState)
		if err != nil {
			return nil, fmt.Errorf("error getting latest chain state: %w", err)
		}
	case strings.HasPrefix(blockIndexOrHash, "0x"):
		hashBytes, err := iotago.DecodeHex(blockIndexOrHash)
		if err != nil {
			return nil, fmt.Errorf("invalid block hash: %v", blockIndexOrHash)
		}
		trieRoot, err := trie.HashFromBytes(hashBytes)
		if err != nil {
			return nil, fmt.Errorf("invalid block hash: %v", blockIndexOrHash)
		}
		chainState, err = ch.Store().StateByTrieRoot(trieRoot)
		if err != nil {
			return nil, fmt.Errorf("error getting block by trie root: %w", err)
		}
	default:
		blockIndex, err := strconv.ParseUint(blockIndexOrHash, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid block number: %v", blockIndexOrHash)
		}
		chainState, err = ch.Store().StateByIndex(uint32(blockIndex))
		if err != nil {
			return nil, fmt.Errorf("error getting block by index: %w", err)
		}
	}
	return chainutil.CallView(chainState, ch, contractName, functionName, params)
}

func EstimateGas(ch chainpkg.Chain, req isc.Request) (*isc.Receipt, error) {
	rec, err := chainutil.SimulateRequest(ch, req, true)
	if err != nil {
		return nil, err
	}
	parsedRec, err := ParseReceipt(ch, rec)
	return parsedRec, err
}
