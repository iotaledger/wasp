package accounts

import (
	"encoding/binary"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

var Contract = coreutil.NewContract(coreutil.CoreContractAccounts, "Chain account ledger contract")

var (
	FuncViewBalance        = coreutil.ViewFunc("balance")
	FuncViewTotalAssets    = coreutil.ViewFunc("totalAssets")
	FuncViewAccounts       = coreutil.ViewFunc("accounts")
	FuncDeposit            = coreutil.Func("deposit")
	FuncSendTo             = coreutil.Func("sendTo")
	FuncWithdraw           = coreutil.Func("withdraw")
	FuncHarvest            = coreutil.Func("harvest")
	FuncGetAccountNonce    = coreutil.ViewFunc("getAccountNonce")
	FuncGetNativeTokensIDs = coreutil.ViewFunc("getNativeTokenIDs")
)

const (
	// prefix for a name of a particular account
	prefixAccount = string(byte(iota) + 'A')
	// map with all accounts listed
	prefixAllAccounts
	// map of account with all on-chain totals listed
	prefixTotalAssetsAccount
	// prefix for the map of nonces
	prefixMaxAssumedNonceKey
	// prefix for all foundries owned by the account
	prefixAccountFoundries
	// prefixNativeTokenOutputMap
	prefixNativeTokenOutputMap
	// prefixFoundryOutputRecords
	prefixFoundryOutputRecords

	ParamAgentID         = "a"
	ParamWithdrawAssetID = "c"
	ParamWithdrawAmount  = "m"
	ParamAccountNonce    = "n"
)

func SerialNumFromNativeTokenID(tokenID *iotago.NativeTokenID) uint32 {
	slice := tokenID[iotago.AliasAddressSerializedBytesSize : iotago.AliasAddressSerializedBytesSize+serializer.UInt32ByteSize]
	return binary.LittleEndian.Uint32(slice)
}

// DecodeBalances TODO move to iscp package
func DecodeBalances(balances dict.Dict) (*iscp.Assets, error) {
	return iscp.NewAssetsFromDict(balances)
}
