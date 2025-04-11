package validation

import (
	"fmt"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmimpl"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"

	old_iotago "github.com/iotaledger/iota.go/v3"
	old_evm_types "github.com/nnikolash/wasp-types-exported/packages/evm/evmtypes"
	"github.com/nnikolash/wasp-types-exported/packages/hashing"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_codec "github.com/nnikolash/wasp-types-exported/packages/kv/codec"
	old_evm "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm"
	old_emulator "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm/emulator"
	old_evmimpl "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm/evmimpl"
)

// ISCMagicPrivileged - ignored (bytes just copied)
// ISCMagicERC20ExternalNativeTokens - ignored (bytes just copied)
// ISCMagicAllowance - ignored (migration is very simple and is done by prefix - does not make much sense it check it)
// emulator.KeyAccountNonce - ignored (bytes just copied)
// emulator.KeyAccountCode - ignored (bytes just copied)
// emulator.KeyAccountState - ignored (bytes just copied)
// emulator.KeyAccountSelfDestructed - ignored (bytes just copied)
// emulator.KeyBlockNumberByBlockHash - ignored (bytes just copied)
// emulator.KeyBlockIndexByTxHash - ignored (bytes just copied)
// emulator.KeyBlockNumberByTxHash - ignored (bytes just copied)
// TODO: Review if ignores here are still correct (maybe some those have changed?)
// TODO: All of these ignores should at least have some intergration tess written for them

func oldEVMContractContentToStr(chainState old_kv.KVStoreReader, fromBlockIndex, toBlockIndex uint32) string {
	cli.DebugLogf("Retrieving old EVM contract content...")
	contractState := old_evm.ContractPartitionR(chainState)
	var allowanceStr, txByBlockStr, blockHeaderStr string

	GoAllAndWait(func() {
		allowanceStr = oldISCMagicAllowanceToStr(contractState)
		cli.DebugLogf("Old ISC magic allowance preview:\n%v", utils.MultilinePreview(allowanceStr))
	}, func() {
		txByBlockStr = oldTransactionsByBlockNumberToStr(contractState, fromBlockIndex, toBlockIndex)
		cli.DebugLogf("Old transactions by block number preview:\n%v", utils.MultilinePreview(txByBlockStr))
	}, func() {
		blockHeaderStr = oldBlockHeaderByBlockNumberToStr(contractState)
		cli.DebugLogf("Old block header by block number preview:\n%v", utils.MultilinePreview(blockHeaderStr))
	})

	return allowanceStr + txByBlockStr + blockHeaderStr
}

func newEVMContractContentToStr(chainState kv.KVStoreReader, fromBlockIndex, toBlockIndex uint32) string {
	cli.DebugLogf("Retrieving new EVM contract content...")
	contractState := evm.ContractPartitionR(chainState)
	var allowanceStr, txByBlockStr, blockHeaderStr string

	GoAllAndWait(func() {
		allowanceStr = newISCMagicAllowanceToStr(contractState)
		cli.DebugLogf("New ISC magic allowance preview:\n%v", utils.MultilinePreview(allowanceStr))
	}, func() {
		txByBlockStr = newTransactionsByBlockNumberToStr(contractState, fromBlockIndex, toBlockIndex)
		cli.DebugLogf("New transactions by block number preview:\n%v", utils.MultilinePreview(txByBlockStr))
	}, func() {
		blockHeaderStr = newBlockHeaderByBlockNumberToStr(contractState)
		cli.DebugLogf("New block header by block number preview:\n%v", utils.MultilinePreview(blockHeaderStr))
	})

	return allowanceStr + txByBlockStr + blockHeaderStr
}

func oldISCMagicAllowanceToStr(contractState old_kv.KVStoreReader) string {
	// TODO: This validation does not find any records. Assuming the reason is that allowance is added and then substracted
	// 	     in the same transaction. Need to double-check it.
	cli.DebugLogf("Retrieving old ISCMagicAllowance entries...")
	iscMagicState := old_evm.ISCMagicSubrealmR(contractState)

	var res strings.Builder
	printProgress, done := NewProgressPrinter("evm_old", "iscmagic allowance", "entries", 0)
	defer done()
	count := 0

	iscMagicState.IterateSorted(old_evmimpl.PrefixAllowance, func(k old_kv.Key, v []byte) bool {
		printProgress()
		count++

		k = utils.MustRemovePrefix(k, old_evmimpl.PrefixAllowance)
		if len(k) != 2*common.AddressLength {
			log.Panicf("unexpected key length: %v", len(k))
		}

		from := common.BytesToAddress([]byte(k[:common.AddressLength]))
		to := common.BytesToAddress([]byte(k[common.AddressLength:]))
		allowance := old_isc.MustAssetsFromBytes(v)

		var allowanceStr strings.Builder
		allowanceStr.WriteString(fmt.Sprintf("base=%v", allowance.BaseTokens))
		for _, nt := range allowance.NativeTokens {
			allowanceStr.WriteString(fmt.Sprintf(", nt=(%v: %v)", nt.ID.ToHex(), nt.Amount))
		}
		for _, nft := range allowance.NFTs {
			allowanceStr.WriteString(fmt.Sprintf(", nft=%v", nft.ToHex()))
		}

		res.WriteString("Magic allowance: ")
		res.WriteString(from.Hex())
		res.WriteString(" -> ")
		res.WriteString(to.Hex())
		res.WriteString(" = ")
		res.WriteString(allowanceStr.String())
		res.WriteString("\n")

		return true
	})

	cli.DebugLogf("Found %v old ISC magic allowance entries", count)
	return res.String()
}

func newISCMagicAllowanceToStr(contractState kv.KVStoreReader) string {
	cli.DebugLogf("Retrieving new ISCMagicAllowance entries...")
	iscMagicState := evm.ISCMagicSubrealmR(contractState)

	var res strings.Builder
	printProgress, done := NewProgressPrinter("evm_new", "iscmagic allowance", "entries", 0)
	defer done()
	count := 0

	iscMagicState.IterateSorted(evmimpl.PrefixAllowance, func(k kv.Key, v []byte) bool {
		printProgress()
		count++

		k = utils.MustRemovePrefix(k, evmimpl.PrefixAllowance)
		if len(k) != 2*common.AddressLength {
			log.Panicf("unexpected key length: %v", len(k))
		}

		from := common.BytesToAddress([]byte(k[:common.AddressLength]))
		to := common.BytesToAddress([]byte(k[common.AddressLength:]))
		allowance := lo.Must(isc.AssetsFromBytes(v))

		var allowanceStr strings.Builder
		allowanceStr.WriteString(fmt.Sprintf("base=%v", allowance.BaseTokens()))
		allowance.Coins.IterateSorted(func(coinType coin.Type, amount coin.Value) bool {
			if coinType == coin.BaseTokenType {
				return true
			}
			ntID := coinTypeToOldNTID(coinType)
			allowanceStr.WriteString(fmt.Sprintf(", nt=(%v: %v)", ntID.ToHex(), amount))
			return true
		})
		allowance.Objects.IterateSorted(func(o isc.IotaObject) bool {
			var nftIDFromObjID = old_iotago.NFTID(o.ID[:])
			// TODO: uncomment after NFT validation is ready
			// if nftIDFromObjType := oldNtfIDFromNewObjectType(o.Type); !nftIDFromObjType.Matches(nftIDFromObjID) {
			// 	panic("failed to convert object to nft ID: %v != %v, oID = %v, oType = %v",
			// 		nftIDFromObjType.ToHex(), nftIDFromObjID.ToHex(),
			// 		o.ID, o.Type)
			// }

			allowanceStr.WriteString(fmt.Sprintf(", nft=%v", nftIDFromObjID.ToHex()))
			return true
		})

		res.WriteString("Magic allowance: ")
		res.WriteString(from.Hex())
		res.WriteString(" -> ")
		res.WriteString(to.Hex())
		res.WriteString(" = ")
		res.WriteString(allowanceStr.String())
		res.WriteString("\n")

		return true
	})

	cli.DebugLogf("Found %v new ISC magic allowance entries", count)
	return res.String()
}

func oldTransactionsByBlockNumberToStr(contractState old_kv.KVStoreReader, fromBlockIndex, toBlockIndex uint32) string {
	cli.DebugLogf("Retrieving old transactions by block number...")
	var res strings.Builder
	keysCount := 0
	firstAvailBlockIndex := max(getFirstAvailableBlockIndex(fromBlockIndex, toBlockIndex), 1)

	GoAllAndWait(func() {
		bc := old_emulator.NewBlockchainDB(OldReadOnlyKVStore(old_evm.EmulatorStateSubrealmR(contractState)), 0, 0)
		printProgress, done := NewProgressPrinter("evm_old", "transactions in block (tx)", "transactions", 0)
		defer done()
		txsCount := 0
		cli.DebugLogf("Retrieving old transactions by block number in range [%v, %v]...", firstAvailBlockIndex, toBlockIndex)

		for blockIndex := firstAvailBlockIndex; blockIndex <= toBlockIndex; blockIndex++ {
			txs := bc.GetTransactionsByBlockNumber(uint64(blockIndex))
			txsCount += len(txs)

			for _, tx := range txs {
				printProgress()
				oldTxBytes := old_evm_types.EncodeTransaction(tx)
				bytesHash := hashing.HashData(oldTxBytes)
				res.WriteString(fmt.Sprintf("Tx in block: %v, txHash: %v, txBytesHash: %v\n", blockIndex, tx.Hash().Hex(), bytesHash.Hex()))
			}
		}

		cli.DebugLogf("Found %v old transactions by block number", txsCount)

		if uint32(txsCount) < (toBlockIndex - firstAvailBlockIndex) {
			panic(fmt.Sprintf("Not enough transactions found in range [%v, %v]: %v", firstAvailBlockIndex, toBlockIndex, txsCount))
		}
	}, func() {
		cli.DebugLogf("Retrieving all old keys of transactions by block number...")
		bcState := old_emulator.BlockchainDBSubrealmR(old_evm.EmulatorStateSubrealmR(contractState))
		printProgress, done := NewProgressPrinter("evm_old", "transactions in block (keys)", "keys", 0)
		defer done()

		bcState.IterateSorted(old_emulator.KeyTransactionsByBlockNumber, func(k old_kv.Key, v []byte) bool {
			printProgress()
			keysCount++
			return true
		})

		cli.DebugLogf("Found %v old keys for transactions by block number", keysCount)

		if uint32(keysCount) < (toBlockIndex - firstAvailBlockIndex) {
			panic(fmt.Sprintf("Not enough transaction keys found in range [%v, %v]: %v", firstAvailBlockIndex, toBlockIndex, keysCount))
		}
	})

	res.WriteString(fmt.Sprintf("Tx in block keys count: %v\n", keysCount))

	return res.String()
}

func newTransactionsByBlockNumberToStr(contractState kv.KVStoreReader, fromBlockIndex, toBlockIndex uint32) string {
	var res strings.Builder
	keysCount := 0

	GoAllAndWait(func() {
		bc := emulator.NewBlockchainDB(NewReadOnlyKVStore(evm.EmulatorStateSubrealmR(contractState)), 0, 0)
		firstAvailBlockIndex := max(getFirstAvailableBlockIndex(fromBlockIndex, toBlockIndex), 1)
		printProgress, done := NewProgressPrinter("evm_new", "transactions in block (tx)", "transactions", 0)
		defer done()
		txsCount := 0
		cli.DebugLogf("Retrieving new transactions by block number in range [%v, %v]...", firstAvailBlockIndex, toBlockIndex)

		for blockIndex := firstAvailBlockIndex; blockIndex <= toBlockIndex; blockIndex++ {
			txs := bc.GetTransactionsByBlockNumber(uint64(blockIndex))
			txsCount += len(txs)

			for _, tx := range txs {
				printProgress()
				oldTxBytes := old_evm_types.EncodeTransaction(tx)
				bytesHash := hashing.HashData(oldTxBytes)
				res.WriteString(fmt.Sprintf("Tx in block: %v, txHash: %v, txBytesHash: %v\n", blockIndex, tx.Hash().Hex(), bytesHash.Hex()))
			}
		}

		cli.DebugLogf("Found %v new transactions by block number", txsCount)
	}, func() {
		cli.DebugLogf("Retrieving all new keys of transactions by block number...")
		bcState := emulator.BlockchainDBSubrealmR(evm.EmulatorStateSubrealmR(contractState))
		printProgress, done := NewProgressPrinter("evm_new", "transactions in block (keys)", "keys", 0)
		defer done()

		bcState.IterateSorted(emulator.KeyTransactionsByBlockNumber, func(k kv.Key, v []byte) bool {
			printProgress()
			keysCount++
			return true
		})

		cli.DebugLogf("Found %v new keys for transactions by block number", keysCount)
	})

	res.WriteString(fmt.Sprintf("Tx in block keys count: %v\n", keysCount))

	return res.String()
}

func oldBlockHeaderByBlockNumberToStr(contractState old_kv.KVStoreReader) string {
	cli.DebugLogf("Retrieving old BlockHeaderByBlockNumber...")
	iscMagicState := old_emulator.BlockchainDBSubrealmR(old_evm.EmulatorStateSubrealmR(contractState))

	var res strings.Builder
	printProgress, done := NewProgressPrinter("evm_old", "block headers by block", "headers", 0)
	defer done()
	count := 0

	iscMagicState.IterateSorted(old_emulator.KeyBlockHeaderByBlockNumber, func(k old_kv.Key, v []byte) bool {
		keyWithoutPrefix := utils.MustRemovePrefix(k, old_emulator.KeyBlockHeaderByBlockNumber)

		blockNumber := old_codec.MustDecodeUint64([]byte(keyWithoutPrefix))
		header := lo.Must(old_emulator.HeaderFromBytes(v))

		// String is very big - hashing it. Might not be needed if later we will hash all strings anyway (which is likely)
		headerStrHash := hashing.HashData([]byte(oldBlockHeaderToStr(header)))
		res.WriteString(fmt.Sprintf("Block header: %v: %v\n", blockNumber, headerStrHash))

		printProgress()
		count++
		return true
	})

	if count == 0 {
		panic("no old block headers found")
	}

	cli.DebugLogf("Found %v old block headers by block number", count)
	return res.String()
}

func newBlockHeaderByBlockNumberToStr(contractState kv.KVStoreReader) string {
	cli.DebugLogf("Retrieving new BlockHeaderByBlockNumber...")
	iscMagicState := emulator.BlockchainDBSubrealmR(evm.EmulatorStateSubrealmR(contractState))

	var res strings.Builder
	printProgress, done := NewProgressPrinter("evm_new", "block headers by block", "headers", 0)
	defer done()
	count := 0

	iscMagicState.IterateSorted(emulator.KeyBlockHeaderByBlockNumber, func(k kv.Key, v []byte) bool {
		keyWithoutPrefix := utils.MustRemovePrefix(k, emulator.KeyBlockHeaderByBlockNumber)

		blockNumber := codec.MustDecode[uint64]([]byte(keyWithoutPrefix))
		newHeader := emulator.MustHeaderFromBytes(v)

		// Reverse conversion
		var oldHeader old_emulator.Header = old_emulator.Header(*newHeader)

		// String is very big - hashing it. Might not be needed if later we will hash all strings anyway (which is likely)
		headerStrHash := hashing.HashData([]byte(oldBlockHeaderToStr(&oldHeader)))
		res.WriteString(fmt.Sprintf("Block header: %v: %v\n", blockNumber, headerStrHash))

		printProgress()
		count++
		return true
	})

	cli.DebugLogf("Found %v new block headers by block number", count)
	return res.String()
}
