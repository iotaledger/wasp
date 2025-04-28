package validation

import (
	"fmt"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/coin"
	evm_types "github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmimpl"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"

	old_evm_types "github.com/nnikolash/wasp-types-exported/packages/evm/evmtypes"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_codec "github.com/nnikolash/wasp-types-exported/packages/kv/codec"
	old_collections "github.com/nnikolash/wasp-types-exported/packages/kv/collections"
	old_evm "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm"
	old_emulator "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm/emulator"
	old_evmimpl "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm/evmimpl"

	old_iotago "github.com/iotaledger/iota.go/v3"
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
	var allowanceStr, txByBlockStr, blockHeaderStr, receiptsStr string

	GoAllAndWait(func() {
		allowanceStr = oldISCMagicAllowanceToStr(contractState)
		cli.DebugLogf("Old ISC magic allowance preview:\n%v", utils.MultilinePreview(allowanceStr))
	}, func() {
		txByBlockStr = oldTransactionsByBlockNumberToStr(contractState, fromBlockIndex, toBlockIndex)
		cli.DebugLogf("Old transactions by block number preview:\n%v", utils.MultilinePreview(txByBlockStr))
	}, func() {
		blockHeaderStr = oldBlockHeaderByBlockNumberToStr(contractState)
		cli.DebugLogf("Old block header by block number preview:\n%v", utils.MultilinePreview(blockHeaderStr))
	}, func() {
		receiptsStr = oldReceiptsByBlockNumberToStr(contractState, fromBlockIndex, toBlockIndex)
		cli.DebugLogf("Old receipts by block number preview:\n%v", utils.MultilinePreview(receiptsStr))
	})

	return allowanceStr + "\n" + txByBlockStr + "\n" + blockHeaderStr + "\n" + receiptsStr
}

func newEVMContractContentToStr(chainState kv.KVStoreReader, fromBlockIndex, toBlockIndex uint32) string {
	cli.DebugLogf("Retrieving new EVM contract content...")
	contractState := evm.ContractPartitionR(chainState)
	var allowanceStr, txByBlockStr, blockHeaderStr, receiptsStr string

	GoAllAndWait(func() {
		allowanceStr = newISCMagicAllowanceToStr(contractState)
		cli.DebugLogf("New ISC magic allowance preview:\n%v", utils.MultilinePreview(allowanceStr))
	}, func() {
		txByBlockStr = newTransactionsByBlockNumberToStr(contractState, fromBlockIndex, toBlockIndex)
		cli.DebugLogf("New transactions by block number preview:\n%v", utils.MultilinePreview(txByBlockStr))
	}, func() {
		blockHeaderStr = newBlockHeaderByBlockNumberToStr(contractState, fromBlockIndex)
		cli.DebugLogf("New block header by block number preview:\n%v", utils.MultilinePreview(blockHeaderStr))
	}, func() {
		receiptsStr = newReceiptsByBlockNumberToStr(contractState, fromBlockIndex, toBlockIndex)
		cli.DebugLogf("New receipts by block number preview:\n%v", utils.MultilinePreview(receiptsStr))
	})

	return allowanceStr + "\n" + txByBlockStr + "\n" + blockHeaderStr + "\n" + receiptsStr
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

		for coinType, amount := range allowance.Coins.Iterate() {
			if coinType == coin.BaseTokenType {
				continue
			}
			ntID := CoinTypeToOldNTID(coinType)
			allowanceStr.WriteString(fmt.Sprintf(", nt=(%v: %v)", ntID.ToHex(), amount))
			continue
		}

		for o := range allowance.Objects.Iterate() {
			var nftIDFromObjID = old_iotago.NFTID(o.ID[:])
			// TODO: uncomment after NFT validation is ready
			// if nftIDFromObjType := oldNtfIDFromNewObjectType(o.Type); !nftIDFromObjType.Matches(nftIDFromObjID) {
			// 	panic("failed to convert object to nft ID: %v != %v, oID = %v, oType = %v",
			// 		nftIDFromObjType.ToHex(), nftIDFromObjID.ToHex(),
			// 		o.ID, o.Type)
			// }

			allowanceStr.WriteString(fmt.Sprintf(", nft=%v", nftIDFromObjID.ToHex()))
			continue
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
				bytesHash := hashValue(oldTxBytes)
				res.WriteString(fmt.Sprintf("Tx in block: %v, txHash: %v, txBytesHash: %v\n", blockIndex, tx.Hash().Hex(), bytesHash))
			}
		}

		cli.DebugLogf("Found %v old transactions by block number", txsCount)

		// Check disabled. There 9997 txs found instead of 9999 in blocks range [13163, 23162]
		// Per Lukas, the reason for that is:
		// Plain onLedger requests don't add TX into the EVM Transaction chain
		// Only if you deposit funds via an OnLedger request, then we create these FakeTXs
		// So it can very much be, that evm blocks have 0 tx inside
		//
		// if uint32(txsCount) < max(toBlockIndex-firstAvailBlockIndex, 1)-1 && toBlockIndex != 0 {
		// 	panic(fmt.Sprintf("Not enough transactions found in range [%v, %v]: %v", firstAvailBlockIndex, toBlockIndex, txsCount))
		// }
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

		if uint32(keysCount) < (toBlockIndex-firstAvailBlockIndex) && toBlockIndex != 0 {
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
				bytesHash := hashValue(oldTxBytes)
				res.WriteString(fmt.Sprintf("Tx in block: %v, txHash: %v, txBytesHash: %v\n", blockIndex, tx.Hash().Hex(), bytesHash))
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
	bcsState := old_emulator.BlockchainDBSubrealmR(old_evm.EmulatorStateSubrealmR(contractState))

	var res strings.Builder
	printProgress, done := NewProgressPrinter("evm_old", "block headers by block", "headers", 0)
	defer done()
	count := 0

	bcsState.IterateSorted(old_emulator.KeyBlockHeaderByBlockNumber, func(k old_kv.Key, v []byte) bool {
		keyWithoutPrefix := utils.MustRemovePrefix(k, old_emulator.KeyBlockHeaderByBlockNumber)

		blockNumber := old_codec.MustDecodeUint64([]byte(keyWithoutPrefix))
		header := lo.Must(old_emulator.HeaderFromBytes(v))

		// String is very big - hashing it. Might not be needed if later we will hash all strings anyway (which is likely)
		headerStrHash := hashValue([]byte(oldEVMBlockHeaderToStr(header)))
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

func newBlockHeaderByBlockNumberToStr(contractState kv.KVStoreReader, fromISCBlockIndex uint32) string {
	cli.DebugLogf("Retrieving new BlockHeaderByBlockNumber...")
	bcState := emulator.BlockchainDBSubrealmR(evm.EmulatorStateSubrealmR(contractState))

	var res strings.Builder
	printProgress, done := NewProgressPrinter("evm_new", "block headers by block", "headers", 0)
	defer done()
	count := 0

	bcState.IterateSorted(emulator.KeyBlockHeaderByBlockNumber, func(k kv.Key, v []byte) bool {
		keyWithoutPrefix := utils.MustRemovePrefix(k, emulator.KeyBlockHeaderByBlockNumber)

		blockNumber := codec.MustDecode[uint64]([]byte(keyWithoutPrefix))
		if blockNumber == 0 && fromISCBlockIndex > 0 {
			// Block 0 is not deleted, when -i option is used. So we just ignore it here.
			return true
		}

		newHeader := emulator.MustHeaderFromBytes(v)

		// Reverse conversion
		var oldHeader old_emulator.Header = old_emulator.Header(*newHeader)

		// String is very big - hashing it. Might not be needed if later we will hash all strings anyway (which is likely)
		headerStrHash := hashValue([]byte(oldEVMBlockHeaderToStr(&oldHeader)))
		res.WriteString(fmt.Sprintf("Block header: %v: %v\n", blockNumber, headerStrHash))

		printProgress()
		count++
		return true
	})

	cli.DebugLogf("Found %v new block headers by block number", count)
	return res.String()
}

func oldReceiptsByBlockNumberToStr(contractState old_kv.KVStoreReader, fromIndex, toIndex uint32) string {
	cli.DebugLogf("Retrieving old receipts by block number...")
	bcState := old_emulator.BlockchainDBSubrealmR(old_evm.EmulatorStateSubrealmR(contractState))
	fromIndex = max(getFirstAvailableBlockIndex(fromIndex, toIndex), 2) - 1
	var receiptsStr strings.Builder
	keysCount := 0

	GoAllAndWait(func() {
		printProgress, done := NewProgressPrinter("evm_old", "receipts by block (receipts)", "receipts", toIndex-fromIndex)
		defer done()
		recCount := 0

		for blockIndex := fromIndex; blockIndex < toIndex; blockIndex++ {
			printProgress()
			receipts := old_collections.NewArrayReadOnly(bcState, string(old_emulator.MakeReceiptsByBlockNumberKey(uint64(blockIndex))))

			for reqIdx := uint32(0); reqIdx < receipts.Len(); reqIdx++ {
				recBytes := receipts.GetAt(reqIdx)
				if recBytes == nil {
					receiptsStr.WriteString(fmt.Sprintf("Block: %v, receipt %v: MISSING\n", blockIndex, reqIdx))
					continue
				}

				rec := lo.Must(old_evm_types.DecodeReceipt(recBytes))
				// String is very big - hashing it. Might not be needed if later we will hash all strings anyway (which is likely)
				recStrHash := hashValue([]byte(evmReceiptsToStr(rec)))
				receiptsStr.WriteString(fmt.Sprintf("Block: %v, receipt %v: %v\n", blockIndex, reqIdx, recStrHash))
				recCount++
			}
		}

		cli.DebugLogf("Found %v old receipts by block number", recCount)

		// Check disabled. There 9997 txs found instead of 9999 in blocks range [13163, 23162]
		// Per Lukas, the reason for that is:
		// Plain onLedger requests don't add TX into the EVM Transaction chain
		// Only if you deposit funds via an OnLedger request, then we create these FakeTXs
		// So it can very much be, that evm blocks have 0 tx inside
		//
		// if uint32(recCount) < max(toIndex-fromIndex, 1)-1 && toIndex != 0 {
		// 	panic(fmt.Sprintf("Not enough receipts found in range [%v, %v]: %v", fromIndex, toIndex, recCount))
		// }
	}, func() {
		printProgress, done := NewProgressPrinter("evm_old", "receipts by block (keys)", "keys", 0)
		defer done()

		bcState.IterateSorted(old_emulator.KeyReceiptsByBlockNumber, func(k old_kv.Key, v []byte) bool {
			keyWithoutPrefix := utils.MustRemovePrefix(k, old_emulator.KeyReceiptsByBlockNumber)
			if len(keyWithoutPrefix) < 8 {
				// cannot even contain blockIndex
				return true
			}
			keyWithoutPrefix = keyWithoutPrefix[8:]
			if keyWithoutPrefix != "" && keyWithoutPrefix[0] != '#' {
				// not length and not element
				return true
			}

			keysCount++
			printProgress()
			return true
		})

		cli.DebugLogf("Found %v keys of old receipts by block number", keysCount)

		if uint32(keysCount) < (toIndex-fromIndex) && toIndex != 0 {
			panic(fmt.Sprintf("Not enough keys found in range [%v, %v]: %v", fromIndex, toIndex, keysCount))
		}
	})

	receiptsStr.WriteString(fmt.Sprintf("Receipts by block keys count: %v\n", keysCount))

	return receiptsStr.String()
}

func newReceiptsByBlockNumberToStr(contractState kv.KVStoreReader, fromIndex, toIndex uint32) string {
	cli.DebugLogf("Retrieving new receipts by block number...")
	bcState := emulator.BlockchainDBSubrealmR(evm.EmulatorStateSubrealmR(contractState))
	fromIndex = max(getFirstAvailableBlockIndex(fromIndex, toIndex), 2) - 1
	var receiptsStr strings.Builder
	keysCount := 0

	GoAllAndWait(func() {
		printProgress, done := NewProgressPrinter("evm_new", "receipts by block (receipts)", "receipts", toIndex-fromIndex)
		defer done()
		recCount := 0

		for blockIndex := fromIndex; blockIndex < toIndex; blockIndex++ {
			printProgress()
			receipts := collections.NewArrayReadOnly(bcState, string(emulator.MakeReceiptsByBlockNumberKey(uint64(blockIndex))))

			for reqIdx := uint32(0); reqIdx < receipts.Len(); reqIdx++ {
				recBytes := receipts.GetAt(reqIdx)
				if recBytes == nil {
					receiptsStr.WriteString(fmt.Sprintf("Block: %v, receipt %v: MISSING\n", blockIndex, reqIdx))
					continue
				}

				rec := lo.Must(evm_types.DecodeReceipt(recBytes))
				// String is very big - hashing it. Might not be needed if later we will hash all strings anyway (which is likely)
				recStrHash := hashValue([]byte(evmReceiptsToStr(rec)))
				receiptsStr.WriteString(fmt.Sprintf("Block: %v, receipt %v: %v\n", blockIndex, reqIdx, recStrHash))
				recCount++
			}
		}

		cli.DebugLogf("Found %v new receipts by block number", recCount)
	}, func() {
		printProgress, done := NewProgressPrinter("evm_new", "receipts by block (keys)", "keys", 0)
		defer done()

		bcState.IterateSorted(emulator.KeyReceiptsByBlockNumber, func(k kv.Key, v []byte) bool {
			keyWithoutPrefix := utils.MustRemovePrefix(k, emulator.KeyReceiptsByBlockNumber)
			if len(keyWithoutPrefix) < 8 {
				// cannot even contain blockIndex
				return true
			}
			keyWithoutPrefix = keyWithoutPrefix[8:]
			if keyWithoutPrefix != "" && keyWithoutPrefix[0] != '#' {
				// not length and not element
				return true
			}

			keysCount++
			printProgress()
			return true
		})

		cli.DebugLogf("Found %v keys of new receipts by block number", keysCount)
	})

	receiptsStr.WriteString(fmt.Sprintf("Receipts by block keys count: %v\n", keysCount))

	return receiptsStr.String()
}
