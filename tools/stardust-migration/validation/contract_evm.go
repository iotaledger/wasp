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

func OldEVMContractContentToStr(chainState old_kv.KVStoreReader, fromBlockIndex, toBlockIndex uint32) string {
	cli.DebugLogf("Retrieving old EVM contract content...\n")
	contractState := old_evm.ContractPartitionR(chainState)
	var allowanceStr, txByBlockStr string

	GoAllAndWait(func() {
		allowanceStr = oldISCMagicAllowanceContentToStr(contractState)
		cli.DebugLogf("Old ISC magic allowance preview:\n%v\n", utils.MultilinePreview(allowanceStr))
	}, func() {
		txByBlockStr = oldTransactionsByBlockNumberToStr(contractState, fromBlockIndex, toBlockIndex)
		cli.DebugLogf("Old transactions by block number preview:\n%v\n", utils.MultilinePreview(txByBlockStr))
	})

	return allowanceStr + txByBlockStr
}

func oldISCMagicAllowanceContentToStr(contractState old_kv.KVStoreReader) string {
	// TODO: This validation does not find any records, needs to be fixed
	cli.DebugLogf("Retrieving old ISCMagicAllowance content...\n")
	iscMagicState := old_evm.ISCMagicSubrealmR(contractState)

	var res strings.Builder
	printProgress, done := NewProgressPrinter("old_evm", "iscmagic allowance", "entries", 0)
	defer done()
	count := 0

	iscMagicState.Iterate(old_evmimpl.PrefixAllowance, func(k old_kv.Key, v []byte) bool {
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

func NewEVMContractContentToStr(chainState kv.KVStoreReader, fromBlockIndex, toBlockIndex uint32) string {
	cli.DebugLogf("Retrieving new EVM contract content...\n")
	contractState := evm.ContractPartitionR(chainState)
	var allowanceStr, txByBlockStr string

	GoAllAndWait(func() {
		allowanceStr = newISCMagicAllowanceContentToStr(contractState)
		cli.DebugLogf("New ISC magic allowance preview:\n%v\n", utils.MultilinePreview(allowanceStr))
	}, func() {
		txByBlockStr = newTransactionsByBlockNumberToStr(contractState, fromBlockIndex, toBlockIndex)
		cli.DebugLogf("New transactions by block number preview:\n%v\n", utils.MultilinePreview(txByBlockStr))
	})

	return allowanceStr + txByBlockStr
}

func newISCMagicAllowanceContentToStr(contractState kv.KVStoreReader) string {
	cli.DebugLogf("Retrieving new ISCMagicAllowance content...\n")
	iscMagicState := evm.ISCMagicSubrealmR(contractState)

	var res strings.Builder
	printProgress, done := NewProgressPrinter("evm", "iscmagic allowance", "entries", 0)
	defer done()
	count := 0

	iscMagicState.Iterate(evmimpl.PrefixAllowance, func(k kv.Key, v []byte) bool {
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
	cli.DebugLogf("Retrieving old transactions by block number...\n")
	var res strings.Builder

	totalKeysCount := 0
	totalTransactionsCount := 0

	GoAllAndWait(func() {
		bc := old_emulator.NewBlockchainDB(OldReadOnlyKVStore(old_evm.EmulatorStateSubrealmR(contractState)), 0, 0)
		firstAvailBlockIndex := max(getFirstAvailableBlockIndex(fromBlockIndex, toBlockIndex), 1)
		printProgress, done := NewProgressPrinter("old_evm", "transactions in block (tx)", "transactions", 0)
		defer done()
		cli.DebugLogf("Retrieving old transactions by block number in range [%v, %v]...\n", firstAvailBlockIndex, toBlockIndex)

		for blockIndex := firstAvailBlockIndex; blockIndex <= toBlockIndex; blockIndex++ {
			txs := bc.GetTransactionsByBlockNumber(uint64(blockIndex))
			totalTransactionsCount += len(txs)

			for _, tx := range txs {
				printProgress()
				oldTxBytes := old_evm_types.EncodeTransaction(tx)
				bytesHash := hashing.HashData(oldTxBytes)
				res.WriteString(fmt.Sprintf("Tx in block: %v, txHash: %v, txBytesHash: %v\n", blockIndex, tx.Hash().Hex(), bytesHash.Hex()))
			}
		}

		cli.DebugLogf("Found %v old transactions by block number", totalTransactionsCount)
	}, func() {
		cli.DebugLogf("Retrieving all old keys of transactions by block number...\n")
		bcState := old_emulator.BlockchainDBSubrealmR(old_evm.EmulatorStateSubrealmR(contractState))
		printProgress, done := NewProgressPrinter("old_evm", "transactions in block (keys)", "keys", 0)
		defer done()

		bcState.IterateSorted(old_emulator.KeyTransactionsByBlockNumber, func(k old_kv.Key, v []byte) bool {
			printProgress()
			totalKeysCount++
			return true
		})

		cli.DebugLogf("Found %v old keys for transactions by block number", totalKeysCount)
	})

	if totalTransactionsCount == 0 {
		panic("no transactions found") // should never happen
	}

	res.WriteString(fmt.Sprintf("Tx in block txs count: %v\n", totalTransactionsCount))
	res.WriteString(fmt.Sprintf("Tx in block keys count: %v\n", totalKeysCount))

	return res.String()
}

func newTransactionsByBlockNumberToStr(contractState kv.KVStoreReader, fromBlockIndex, toBlockIndex uint32) string {
	var res strings.Builder

	totalKeysCount := 0
	totalTransactionsCount := 0

	GoAllAndWait(func() {
		bc := emulator.NewBlockchainDB(NewReadOnlyKVStore(evm.EmulatorStateSubrealmR(contractState)), 0, 0)
		firstAvailBlockIndex := max(getFirstAvailableBlockIndex(fromBlockIndex, toBlockIndex), 1)
		printProgress, done := NewProgressPrinter("evm", "transactions in block (tx)", "transactions", 0)
		defer done()
		cli.DebugLogf("Retrieving new transactions by block number in range [%v, %v]...\n", firstAvailBlockIndex, toBlockIndex)

		for blockIndex := firstAvailBlockIndex; blockIndex <= toBlockIndex; blockIndex++ {
			txs := bc.GetTransactionsByBlockNumber(uint64(blockIndex))
			totalTransactionsCount += len(txs)

			for _, tx := range txs {
				printProgress()
				oldTxBytes := old_evm_types.EncodeTransaction(tx)
				bytesHash := hashing.HashData(oldTxBytes)
				res.WriteString(fmt.Sprintf("Tx in block: %v, txHash: %v, txBytesHash: %v\n", blockIndex, tx.Hash().Hex(), bytesHash.Hex()))
			}
		}

		cli.DebugLogf("Found %v new transactions by block number", totalTransactionsCount)
	}, func() {
		cli.DebugLogf("Retrieving all new keys of transactions by block number...\n")
		bcState := emulator.BlockchainDBSubrealmR(evm.EmulatorStateSubrealmR(contractState))
		printProgress, done := NewProgressPrinter("evm", "transactions in block (keys)", "keys", 0)
		defer done()

		bcState.IterateSorted(emulator.KeyTransactionsByBlockNumber, func(k kv.Key, v []byte) bool {
			printProgress()
			totalKeysCount++
			return true
		})

		cli.DebugLogf("Found %v new keys for transactions by block number", totalKeysCount)
	})

	res.WriteString(fmt.Sprintf("Tx in block txs count: %v\n", totalTransactionsCount))
	res.WriteString(fmt.Sprintf("Tx in block keys count: %v\n", totalKeysCount))

	return res.String()
}
