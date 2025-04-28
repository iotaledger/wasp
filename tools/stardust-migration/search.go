package main

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/samber/lo"
	cmd "github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/tools/stardust-migration/blockindex"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/oldstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/db"

	old_iotago "github.com/iotaledger/iota.go/v3"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_collections "github.com/nnikolash/wasp-types-exported/packages/kv/collections"
	old_dict "github.com/nnikolash/wasp-types-exported/packages/kv/dict"
	old_state "github.com/nnikolash/wasp-types-exported/packages/state"
	old_indexedstore "github.com/nnikolash/wasp-types-exported/packages/state/indexedstore"
	old_vm "github.com/nnikolash/wasp-types-exported/packages/vm"
	old_accounts "github.com/nnikolash/wasp-types-exported/packages/vm/core/accounts"
	old_blocklog "github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"
	old_evm "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm"
	old_evmimpl "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm/evmimpl"
	old_governance "github.com/nnikolash/wasp-types-exported/packages/vm/core/governance"
)

type SearchFunc func(state old_kv.KVStoreReader, onFound func(k old_kv.Key, v []byte) bool)
type SearchFuncConstructor func(store old_indexedstore.IndexedStore, args []string) SearchFunc

type SearchOptions struct {
	IncludeDeletions bool
	ArgsUsage        string
}

type SearchOption func(*SearchOptions)

func IncludeDeletions() SearchOption {
	return func(opts *SearchOptions) {
		opts.IncludeDeletions = true
	}
}

func ArgsUsage(usage string) SearchOption {
	return func(opts *SearchOptions) {
		opts.ArgsUsage = usage
	}
}

func search(name string, f SearchFuncConstructor, options SearchOptions) func(c *cmd.Context) error {
	return func(c *cmd.Context) error {
		chainDBDir := c.Args().Get(0)
		fromIndex := uint32(c.Uint64("from-block"))
		toIndex := uint32(c.Uint64("to-block"))
		findAll := c.Bool("all")
		threadsCount := uint32(c.Uint("parallel"))
		args := c.Args().Slice()[1:]

		kvs := db.ConnectOld(chainDBDir)
		store := old_indexedstore.New(old_state.NewStoreWithUniqueWriteMutex(kvs))

		if toIndex == 0 {
			cli.Logf("Using latest new state index")
			toIndex = lo.Must(store.LatestBlockIndex())
		}

		cli.Logf("Searching for %v in blocks [%d ; %d]", name, fromIndex, toIndex)
		searchLinear(c.Context, name, store, fromIndex, toIndex, findAll, threadsCount, f, args, &options)

		return nil
	}
}

func searchLinear(
	ctx context.Context,
	name string,
	store old_indexedstore.IndexedStore,
	fromIndex, toIndex uint32, findAll bool, threadsCount uint32,
	createHandler SearchFuncConstructor,
	args []string,
	opts *SearchOptions,
) {

	indexer := blockindex.New(store)
	e := errgroup.Group{}

	found := atomic.Bool{}
	firstContainingBlockIndexes := lo.RepeatBy(int(threadsCount), func(i int) uint32 { return math.MaxUint32 })
	firstFoundKeys := make([]old_kv.Key, threadsCount)
	firstFoundValues := make([][]byte, threadsCount)
	lastCheckedBlockIndexes := make([]uint32, threadsCount)
	totalBlocks := toIndex - fromIndex + 1
	printProgress, done := cli.NewProgressPrinter("blocks", totalBlocks)

	cli.Logf("Starting %v search threads", threadsCount)
	f := createHandler(store, args)

	for i := uint32(0); i < threadsCount; i++ {
		i := i

		e.Go(func() error {
			for blockIndex := fromIndex + i; blockIndex <= toIndex; blockIndex += threadsCount {
				block, _ := indexer.BlockByIndex(blockIndex)

				var state old_kv.KVStoreReader
				if opts.IncludeDeletions {
					state = utils.DictKvFromMuts(block.Mutations())
				} else {
					state = old_dict.Dict(block.Mutations().Sets)
				}

				f(state, func(k old_kv.Key, v []byte) bool {
					found.Store(true)

					if firstFoundKeys[i] == "" {
						firstContainingBlockIndexes[i] = blockIndex
						firstFoundKeys[i] = k
						firstFoundValues[i] = v
					}

					if findAll {
						oper := lo.Ternary(v != nil, "SET", "DEL")
						cli.Logf("%v: Block %v: %v %x", name, blockIndex, oper, []byte(k))
					}

					return findAll
				})

				lastCheckedBlockIndexes[i] = blockIndex

				if i == 0 && (blockIndex/threadsCount)%100 == 0 {
					printProgress(threadsCount * 100)
				}

				if found.Load() && !findAll {
					return nil
				}
				if ctx.Err() != nil {
					return nil
				}
			}

			return nil
		})
	}

	e.Wait()
	done()

	if ctx.Err() != nil {
		cli.Logf("Interrupted. Last checked block: %d", lo.Min(lastCheckedBlockIndexes))
		os.Exit(1)
	}

	_, earliestThreadIdx := lo.MinIndex(firstContainingBlockIndexes)
	if firstFoundKeys[earliestThreadIdx] == "" {
		cli.Logf("No %v found in blocks [%d; %d]\n", name, fromIndex, toIndex)
		return
	}

	if !findAll {
		earliestBlockIndex := firstContainingBlockIndexes[earliestThreadIdx]
		earliestKey := firstFoundKeys[earliestThreadIdx]
		earliestValue := firstFoundValues[earliestThreadIdx]
		oper := lo.Ternary(earliestValue != nil, "SET", "DEL")
		vStr := lo.Ternary(earliestValue != nil, fmt.Sprintf("\nValue: %x", earliestValue), "")
		cli.Logf("Found FIRST occurrence:\n%v: Block %v: %v %x%v", name, earliestBlockIndex, oper, []byte(earliestKey), vStr)
	}
}

// Although works, it does not make much sense - the object could be created on one block and immediatelly deleted on next.
// So it's impossible to find it by binary search...
// func searchBinary(ctx context.Context, name string, store old_indexedstore.IndexedStore, fromIndex, toIndex uint32, f StateContainsTargetCheckFunc) error {
// 	index := blockindex.New(store)

// 	lastFoundKey, lastFoundValue := old_kv.Key(""), []byte{}
// 	cli.Logf("Searching range: [%v ; %v]", fromIndex, toIndex)

// 	for {
// 		middleIndex := (fromIndex + toIndex) / 2
// 		cli.Logf("Searching at index %v", middleIndex)

// 		_, trieRoot := index.BlockByIndex(middleIndex)
// 		state := lo.Must(store.StateByTrieRoot(trieRoot))

// 		found := false
// 		f(state, func(k old_kv.Key, v []byte) bool {
// 			found = true
// 			lastFoundKey = k
// 			lastFoundValue = v
// 			return false
// 		})

// 		if fromIndex == toIndex {
// 			break
// 		}
// 		if found {
// 			toIndex = middleIndex
// 			cli.Logf("Found %v in block %d - new range = [%d; %d]", name, middleIndex, fromIndex, toIndex)
// 		} else {
// 			fromIndex = middleIndex + 1
// 			cli.Logf("Not found %v in block %d - new range = [%d; %d]", name, middleIndex, fromIndex, toIndex)
// 		}
// 	}

// 	if lastFoundValue != nil {
// 		cli.Logf("Found %v in block %d:\nKey = %x\nValue = %x", name, fromIndex, lastFoundKey, lastFoundValue)
// 	} else {
// 		cli.Logf("No %v found in blocks [%d; %d]\n", name, fromIndex, toIndex)
// 	}

// 	return nil
// }

func searchISCMagicAllowance(store old_indexedstore.IndexedStore, args []string) SearchFunc {
	return func(chainState old_kv.KVStoreReader, onFound func(k old_kv.Key, v []byte) bool) {
		contractState := old_evm.ISCMagicSubrealmR(old_evm.ContractPartitionR(chainState))
		contractState.Iterate(old_evmimpl.PrefixAllowance, onFound)
	}
}

func searchNFT(store old_indexedstore.IndexedStore, args []string) SearchFunc {
	return func(chainState old_kv.KVStoreReader, onFound func(k old_kv.Key, v []byte) bool) {
		contractState := oldstate.GetContactStateReader(chainState, old_accounts.Contract.Hname())
		nfts := old_collections.NewMapReadOnly(contractState, old_accounts.KeyNFTOwner)
		nfts.Iterate(func(k, v []byte) bool { onFound(old_kv.Key(k), v); return false })
	}
}

func searchBlockKeepAmountNot10000(store old_indexedstore.IndexedStore, args []string) SearchFunc {
	return func(chainState old_kv.KVStoreReader, onFound func(k old_kv.Key, v []byte) bool) {
		contractState := oldstate.GetContactStateReader(chainState, old_governance.Contract.Hname())
		blockKeepAmount := old_governance.NewStateAccess(contractState).GetBlockKeepAmount()
		if blockKeepAmount != 10000 {
			onFound(old_kv.Key(""), []byte{0})
		}
	}
}

func searchFoundies(store old_indexedstore.IndexedStore, args []string) SearchFunc {
	return func(chainState old_kv.KVStoreReader, onFound func(k old_kv.Key, v []byte) bool) {
		contractState := oldstate.GetContactStateReader(chainState, old_accounts.Contract.Hname())
		contractState.Iterate(old_accounts.PrefixFoundries, onFound)
	}
}

func searchNativeTokens(store old_indexedstore.IndexedStore, args []string) SearchFunc {
	return func(chainState old_kv.KVStoreReader, onFound func(k old_kv.Key, v []byte) bool) {
		contractState := oldstate.GetContactStateReader(chainState, old_accounts.Contract.Hname())
		contractState.Iterate(old_accounts.PrefixNativeTokens, onFound)
	}
}

func searchStrangeNativeTokenRecords(store old_indexedstore.IndexedStore, args []string) SearchFunc {
	return func(chainState old_kv.KVStoreReader, onFound func(k old_kv.Key, v []byte) bool) {
		// Some of the records, which start with PrefixNativeToken, had previously invalid account key.
		// Also, seemingly they always delete value, although there is nothing set.
		var IsValidOldAccountKeyBytesLen = func(n int) bool {
			return n == isc.HnameLength || n == common.AddressLength || n == iotago.AddressLen
		}

		contractState := oldstate.GetContactStateReader(chainState, old_accounts.Contract.Hname())
		contractState.Iterate(old_accounts.PrefixNativeTokens, func(k old_kv.Key, v []byte) bool {
			oldAccKey, oldNtIDBytes := utils.MustSplitMapKey(k, -old_iotago.FoundryIDLength-1, old_accounts.PrefixNativeTokens)
			if oldNtIDBytes == "" {
				// not a map entry
				return true
			}

			if oldAccKey == old_accounts.L2TotalsAccount {
				return true
			}

			if !IsValidOldAccountKeyBytesLen(len(oldAccKey)) {
				return onFound(k, v)
			}

			return true
		})
	}
}

func searchKey(store old_indexedstore.IndexedStore, args []string) SearchFunc {
	targetKey := args[0]
	if !strings.HasPrefix(targetKey, "0x") {
		targetKey = "0x" + targetKey
	}
	targetKey = string(lo.Must(cryptolib.DecodeHex(targetKey)))

	return func(chainState old_kv.KVStoreReader, onFound func(k old_kv.Key, v []byte) bool) {
		chainState.Iterate(old_kv.Key(targetKey), func(key old_kv.Key, value []byte) bool {
			if key == old_kv.Key(targetKey) {
				return onFound(key, value)
			}
			return true
		})
	}
}

func searchGasFeePolicyChange(store old_indexedstore.IndexedStore, args []string) SearchFunc {
	return func(chainState old_kv.KVStoreReader, onFound func(k old_kv.Key, v []byte) bool) {
		oldContractState := oldstate.GetContactStateReader(chainState, old_governance.Contract.Hname())
		policyBytes := oldContractState.Get(old_governance.VarGasFeePolicyBytes)
		if policyBytes != nil {
			onFound(old_governance.VarGasFeePolicyBytes, policyBytes)
		}
	}
}

func searchGasBudgetExceeded(store old_indexedstore.IndexedStore, args []string) SearchFunc {
	return func(chainState old_kv.KVStoreReader, onFound func(k old_kv.Key, v []byte) bool) {
		oldContractState := oldstate.GetContactStateReader(chainState, old_blocklog.Contract.Hname())
		oldRequests := old_collections.NewMapReadOnly(oldContractState, old_blocklog.PrefixRequestReceipts)

		oldRequests.Iterate(func(k, v []byte) bool {
			rec := lo.Must(old_blocklog.RequestReceiptFromBytes(v, 0, 0))
			if rec.Error != nil && rec.Error.ErrorCode == old_vm.ErrGasBudgetExceeded.Code() {
				cli.Println("Request result error: ", rec.Error.AsGoError())
				return onFound(old_kv.Key(k), v)
			}
			return true
		})
	}
}

func searchNicoleCoin(store old_indexedstore.IndexedStore, args []string) SearchFunc {
	const nicoleCoinID = "0x08971dc160d5ae8c457f7eddc15a39035b6190130b4dbb5663550795575ae19db50100000000"
	nicoleCoinIDBytes := lo.Must(cryptolib.DecodeHex(nicoleCoinID))

	return func(chainState old_kv.KVStoreReader, onFound func(k old_kv.Key, v []byte) bool) {
		oldContractState := oldstate.GetContactStateReader(chainState, old_accounts.Contract.Hname())
		oldContractState.Iterate("", func(k old_kv.Key, v []byte) bool {
			if bytes.Contains([]byte(k), nicoleCoinIDBytes) {
				return onFound(k, v)
			}
			return true
		})
	}
}

func searchCrossChain(store old_indexedstore.IndexedStore, args []string) SearchFunc {
	ownChainID := old_isc.ChainID(GetAnchorOutput(lo.Must(store.LatestState())).AliasID)

	return func(chainState old_kv.KVStoreReader, onFound func(k old_kv.Key, v []byte) bool) {
		oldContractState := oldstate.GetContactStateReader(chainState, old_accounts.Contract.Hname())
		old_accounts.AllAccountsMapR(oldContractState).Iterate(func(accKey, v []byte) bool {
			accID := lo.Must(old_accounts.AgentIDFromKey(old_kv.Key(accKey), ownChainID))
			//if !accID.BelongsToChain(ownChainID) {
			if isCrossChainAgentID(accID, ownChainID) {
				return onFound(old_kv.Key(accKey), v)
			}
			return true
		})

		// oldContractState := oldstate.GetContactStateReader(chainState, old_blocklog.Contract.Hname())
		// oldRequests := old_collections.NewMapReadOnly(oldContractState, old_blocklog.PrefixRequestReceipts)

		// oldRequests.Iterate(func(k, v []byte) bool {
		// 	rec := lo.Must(old_blocklog.RequestReceiptFromBytes(v, 0, 0))

		// 	sender := rec.Request.SenderAccount()
		// 	if !sender.BelongsToChain(ownChainID) {
		// 		return onFound(old_kv.Key(k), v)
		// 	}

		// 	return true
		// })
	}
}

func isCrossChainAgentID(agentID old_isc.AgentID, ownChainID old_isc.ChainID) bool {
	switch agentID.Kind() {
	case old_isc.AgentIDKindContract:
		agentID := agentID.(*old_isc.ContractAgentID)
		oldAgentChainID := agentID.ChainID()
		return !bytes.Equal(ownChainID.Bytes(), oldAgentChainID.Bytes())
	case old_isc.AgentIDKindEthereumAddress:
		oldAgentID := agentID.(*old_isc.EthereumAddressAgentID)
		oldAgentChainID := oldAgentID.ChainID()
		return !oldAgentChainID.Equals(ownChainID)
	default:
		return false
	}
}
