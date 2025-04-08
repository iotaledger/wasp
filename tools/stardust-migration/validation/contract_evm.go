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
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmimpl"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"

	old_iotago "github.com/iotaledger/iota.go/v3"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_evm "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm"
	old_evmimpl "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm/evmimpl"
)

// ISCMagicPrivileged - ignored (bytes just copied)
// ISCMagicERC20ExternalNativeTokens - ignored (bytes just copied)
// ISCMagicAllowance - ignored (migration is very simple and is done by prefix - does not make much sense it check it)
// TODO: Review if ignors here are still correct (maybe some those have changed?)

func OldEVMContractContentToStr(chainState old_kv.KVStoreReader) string {
	cli.DebugLogf("Retrieving old EVM contract content...\n")
	contractState := old_evm.ContractPartitionR(chainState)
	var allowanceStr string

	GoAllAndWait(func() {
		allowanceStr = oldISCMagicAllowanceContentToStr(contractState)
		cli.DebugLogf("Old ISC magic allowance preview:\n%v\n", utils.MultilinePreview(allowanceStr))
	})

	return allowanceStr
}

func oldISCMagicAllowanceContentToStr(contractState old_kv.KVStoreReader) string {
	cli.DebugLogf("Retrieving old ISCMagicAllowance content...\n")
	iscMagicState := old_evm.ISCMagicSubrealmR(contractState)

	var res strings.Builder
	printProgress, done := NewProgressPrinter("old_evm", "allowance", "entries", 0)
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

func NewEVMContractContentToStr(chainState kv.KVStoreReader) string {
	cli.DebugLogf("Retrieving new EVM contract content...\n")
	contractState := evm.ContractPartitionR(chainState)
	var allowanceStr string

	GoAllAndWait(func() {
		allowanceStr = newISCMagicAllowanceContentToStr(contractState)
		cli.DebugLogf("New ISC magic allowance preview:\n%v\n", utils.MultilinePreview(allowanceStr))
	})

	return allowanceStr
}

func newISCMagicAllowanceContentToStr(contractState kv.KVStoreReader) string {
	cli.DebugLogf("Retrieving new ISCMagicAllowance content...\n")
	iscMagicState := evm.ISCMagicSubrealmR(contractState)

	var res strings.Builder
	printProgress, done := NewProgressPrinter("evm", "allowance", "entries", 0)
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
