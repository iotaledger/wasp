package evmimpl

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/evm/solidity"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
)

func Nonce(evmPartition kv.KVStoreReader, addr common.Address) uint64 {
	emuState := evm.EmulatorStateSubrealmR(evmPartition)
	stateDBStore := emulator.StateDBSubrealmR(emuState)
	return emulator.GetNonce(stateDBStore, addr)
}

func RegisterERC721NFTCollectionByNFTId(store kv.KVStore, nft *isc.NFT) {
	state := emulator.NewStateDBFromKVStore(store)
	addr := iscmagic.ERC721NFTCollectionAddress(nft.ID)

	if state.Exist(addr) {
		panic(errEVMAccountAlreadyExists)
	}

	metadata, err := isc.IRC27NFTMetadataFromBytes(nft.Metadata)
	if err != nil {
		panic(errEVMCanNotDecodeERC27Metadata)
	}

	state.CreateAccount(addr)
	state.SetCode(addr, iscmagic.ERC721NFTCollectionRuntimeBytecode)
	// see ERC721NFTCollection_storage.json
	state.SetState(addr, solidity.StorageSlot(2), solidity.StorageEncodeBytes32(nft.ID[:]))
	for k, v := range solidity.StorageEncodeString(3, metadata.Name) {
		state.SetState(addr, k, v)
	}

	addToPrivileged(store, addr)
}
