package main

import (
	"bytes"
	"fmt"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_parameters "github.com/nnikolash/wasp-types-exported/packages/parameters"
	old_state "github.com/nnikolash/wasp-types-exported/packages/state"
	old_indexedstore "github.com/nnikolash/wasp-types-exported/packages/state/indexedstore"
	old_accounts "github.com/nnikolash/wasp-types-exported/packages/vm/core/accounts"
	old_blocklog "github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"
	old_evm "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	old_iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/tools/stardust-migration/migrations"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/newstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/oldstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/db"
)

type Cont struct {
	Key   []byte
	Value []byte
}

func OpenDBs(stateIndex uint32) (old_state.State, state.State) {
	old_parameters.InitL1(&old_parameters.L1Params{
		Protocol: &old_iotago.ProtocolParameters{
			Bech32HRP: old_iotago.PrefixMainnet,
		},
		BaseToken: &old_parameters.BaseToken{
			Decimals: 9, // TODO: 9? 6?
		},
	})

	srcChainDBDir := "/mnt/isc/wasp_stardust_mainnet/chains/data/iota1pzt3mstq6khgc3tl0mwuzk3eqddkryqnpdxmk4nr25re2466uxwm28qqxu5"
	destChainDBDir := "/tmp/isc-migration"

	srcKVS := db.ConnectOld(srcChainDBDir)
	srcStore := old_indexedstore.New(old_state.NewStoreWithUniqueWriteMutex(srcKVS))

	destKVS := db.ConnectNew(destChainDBDir)
	destStore := indexedstore.New(state.NewStoreWithUniqueWriteMutex(destKVS))

	srcState := lo.Must(srcStore.StateByIndex(stateIndex))
	dstState := lo.Must(destStore.StateByIndex(1000))

	return srcState, dstState
}

func TestDumpAccountKeys(t *testing.T) {
	srcState, _ := OpenDBs(5200000)

	accountState := oldstate.GetContactStateReader(srcState, old_accounts.Contract.Hname())

	old_accounts.AllAccountsMapR(accountState).Iterate(func(elemKey []byte, value []byte) bool {
		a, err := old_accounts.AgentIDFromKey(old_kv.Key(elemKey), old_isc.ChainID{})
		require.NoError(t, err)

		fmt.Println(a)
		return true
	})
}

func TestDumpTries(t *testing.T) {
	dbPath := "/mnt/dev/Coding/iota/isc-rebased-migration/waspdb/chains/data/0xecadd251cfd00e65b1b742822e5b8c3ce1b7a4e427316a49797955753e1f3b20"
	destKVS := db.ConnectNew(dbPath)
	destStore := indexedstore.New(state.NewStoreWithUniqueWriteMutex(destKVS))
	s, _ := destStore.LatestState()
	blockIdx, _ := destStore.LatestBlockIndex()

	fmt.Printf("Latest block index: %d\n", blockIdx)
	fmt.Printf("Block: %d, TrieRoot: %s\n", s.BlockIndex(), s.TrieRoot())

	count := 0
	for {
		blockIdx--
		childs, _ := destStore.StateByIndex(blockIdx)
		fmt.Printf("Block: %d, TrieRoot: %s\n", childs.BlockIndex(), childs.TrieRoot())

		count++

		if count > 1000 {
			break
		}
	}

}

func TestGetAccountBalance(t *testing.T) {
	const stateIndex = 3
	srcState, dstState := OpenDBs(5)

	oldBlocklogState := oldstate.GetContactStateReader(srcState, old_blocklog.Contract.Hname())
	blockInfo, requests, _ := old_blocklog.GetRequestsInBlock(oldBlocklogState, 3)

	fmt.Print(blockInfo.String())
	fmt.Printf("Allowance:%v\nAssets:%v\nCallTarget:%v\nEVM CallMsg:%v\n", requests[0].Allowance(), requests[0].Assets(), requests[0].CallTarget(), string(lo.Must(requests[0].Params().MarshalJSON())))

	targetAddress := requests[0].SenderAccount()

	newBlockLogState := newstate.GetContactStateReader(dstState, blocklog.Contract.Hname())
	blocklogReader := blocklog.NewStateReader(newBlockLogState)
	newBlockInfo, _ := blocklogReader.GetBlockInfo(3)
	_, newRequests, _ := blocklogReader.GetRequestsInBlock(3)

	fmt.Print(newBlockInfo.String())
	a, _ := accounts.FuncTransferAllowanceTo.Input1.Decode(newRequests[0].Message().Params.MustAt(0))
	fmt.Printf("Allowance:%v\nAssets:%v\nCallTarget:%v\nTarget:%s\n", lo.Must(newRequests[0].Allowance()), newRequests[0].Assets(), newRequests[0].Message(), a.String())

	newAccountsState := newstate.GetContactStateReader(dstState, accounts.Contract.Hname())
	baseTokenNew, remainder := accounts.NewStateReader(5, newAccountsState).GetBaseTokensBalance(migrations.OldAgentIDtoNewAgentID(targetAddress, old_isc.ChainID{}))
	fmt.Printf("%d / %d\n", baseTokenNew, remainder)

}

/*
0x03971dc160d5ae8c457f7eddc15a39035b6190130b4dbb5663550795575ae19db5c0a46d280bdbf7c0b6184bd8d58f093d81c7ba5a
0x032bc9ef026dfd9536880aace330f0f2c4bd5c7f37bef4b4483ab9ec611f013efbc0a46d280bdbf7c0b6184bd8d58f093d81c7ba5a


transferAllowanceTo:
From:
0x78357316239040e19fC823372cC179ca75e64b81@iota1pzt3mstq6khgc3tl0mwuzk3eqddkryqnpdxmk4nr25re2466uxwm28qqxu5
To:
0x78357316239040e19fC823372cC179ca75e64b81@0x218321f6adeb3e975fba2c7f51f425e468ab213d775696c151c887abe0bb82a6

transferAllowanceTo:
From:
0xC0A46D280BdbF7c0B6184bd8d58F093D81C7BA5A@iota1pzt3mstq6khgc3tl0mwuzk3eqddkryqnpdxmk4nr25re2466uxwm28qqxu5
To:0xC0A46D280BdbF7c0B6184bd8d58F093D81C7BA5A@0x218321f6adeb3e975fba2c7f51f425e468ab213d775696c151c887abe0bb82a6

transferAllowanceTo:
From:
0xdeD212B8BAb662B98f49e757CbB409BB7808dc10@iota1pzt3mstq6khgc3tl0mwuzk3eqddkryqnpdxmk4nr25re2466uxwm28qqxu5
To:0xdeD212B8BAb662B98f49e757CbB409BB7808dc10@0x218321f6adeb3e975fba2c7f51f425e468ab213d775696c151c887abe0bb82a6

transferAllowanceTo:
From:
0xf7Ee22cee0d9cfe97b0c84c15ee03458D41C5bBf@iota1pzt3mstq6khgc3tl0mwuzk3eqddkryqnpdxmk4nr25re2466uxwm28qqxu5
To:0xf7Ee22cee0d9cfe97b0c84c15ee03458D41C5bBf@0x218321f6adeb3e975fba2c7f51f425e468ab213d775696c151c887abe0bb82a6
*/

func TestDumpEVM(t *testing.T) {
	const stateIndex = 2

	srcState, dstState := OpenDBs(stateIndex)

	oldEVM := old_evm.ContractPartitionR(srcState)
	newEVM := evm.ContractPartitionR(dstState)

	oldEVMKeys := 0
	oldMap := map[string][]byte{}
	oldData := make([]Cont, 0)
	oldEVM.Iterate("", func(key old_kv.Key, value []byte) bool {
		fmt.Printf("key: %s\n", key[:])
		oldEVMKeys += 1
		oldMap[string(key[:])] = value
		oldData = append(oldData, Cont{Key: []byte(key[:]), Value: (value)})

		return true
	})

	fmt.Printf("oldEVMKeys: %d\n", oldEVMKeys)

	fmt.Print("\nNEW EVM\n")

	newEVMKeys := 0
	newMap := map[string][]byte{}
	newData := make([]Cont, 0)
	newEVM.Iterate("", func(key kv.Key, value []byte) bool {
		fmt.Printf("key: %s\n", key[:])
		newEVMKeys += 1
		newMap[string(key[:])] = value
		newData = append(newData, Cont{Key: []byte(key[:]), Value: (value)})
		return true
	})

	oldKeys := lo.Keys(oldMap)
	newKeys := lo.Keys(newMap)

	slices.Sort(oldKeys)
	slices.Sort(newKeys)

	fmt.Printf("Keys Diff: %s\n", cmp.Diff(oldKeys, newKeys))
	fmt.Printf("newEVMKeys: %d\n", newEVMKeys)

	oldValues := lo.Values(oldMap)
	newValues := lo.Values(newMap)

	slices.SortFunc(oldValues, func(a, b []byte) int {
		return bytes.Compare(a, b)
	})
	slices.SortFunc(newValues, func(a, b []byte) int {
		return bytes.Compare(a, b)
	})
}
