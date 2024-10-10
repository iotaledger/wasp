package suiclient

type SuiMethod string

func (s SuiMethod) String() string {
	return string(s)
}

type SuiXMethod string

func (s SuiXMethod) String() string {
	return string(s)
}

type UnsafeMethod string

func (u UnsafeMethod) String() string {
	return string(u)
}

const (
	// Coin Query API
	getAllBalances  SuiXMethod = "iotax_getAllBalances"
	getAllCoins     SuiXMethod = "iotax_getAllCoins"
	getBalance      SuiXMethod = "iotax_getBalance"
	getCoinMetadata SuiXMethod = "iotax_getCoinMetadata"
	getCoins        SuiXMethod = "iotax_getCoins"
	getTotalSupply  SuiXMethod = "iotax_getTotalSupply"

	// Extended API
	getDynamicFieldObject     SuiXMethod = "iotax_getDynamicFieldObject"
	getDynamicFields          SuiXMethod = "iotax_getDynamicFields"
	getOwnedObjects           SuiXMethod = "iotax_getOwnedObjects"
	queryEvents               SuiXMethod = "iotax_queryEvents"
	queryTransactionBlocks    SuiXMethod = "iotax_queryTransactionBlocks"
	resolveNameServiceAddress SuiXMethod = "iotax_resolveNameServiceAddress"
	resolveNameServiceNames   SuiXMethod = "iotax_resolveNameServiceNames"
	subscribeEvent            SuiXMethod = "iotax_subscribeEvent"
	subscribeTransaction      SuiXMethod = "iotax_subscribeTransaction"

	// Governance Read API
	getCommitteeInfo        SuiXMethod = "iotax_getCommitteeInfo" // TODO
	getLatestSuiSystemState SuiXMethod = "iotax_getLatestIotaSystemState"
	getReferenceGasPrice    SuiXMethod = "iotax_getReferenceGasPrice"
	getStakes               SuiXMethod = "iotax_getStakes"
	getStakesByIds          SuiXMethod = "iotax_getStakesByIds"
	getValidatorsApy        SuiXMethod = "iotax_getValidatorsApy"

	// Move Utils
	getMoveFunctionArgTypes           SuiMethod = "iota_getMoveFunctionArgTypes"           // TODO
	getNormalizedMoveFunction         SuiMethod = "iota_getNormalizedMoveFunction"         // TODO
	getNormalizedMoveModule           SuiMethod = "iota_getNormalizedMoveModule"           // TODO
	getNormalizedMoveModulesByPackage SuiMethod = "iota_getNormalizedMoveModulesByPackage" // TODO
	getNormalizedMoveStruct           SuiMethod = "iota_getNormalizedMoveStruct"           // TODO

	// Read API
	getChainIdentifier                SuiMethod = "iota_getChainIdentifier"
	getCheckpoint                     SuiMethod = "iota_getCheckpoint"
	getCheckpoints                    SuiMethod = "iota_getCheckpoints"
	getEvents                         SuiMethod = "iota_getEvents"
	getLatestCheckpointSequenceNumber SuiMethod = "iota_getLatestCheckpointSequenceNumber"
	getLoadedChildObjects             SuiMethod = "iota_getLoadedChildObjects" // TODO
	getObject                         SuiMethod = "iota_getObject"
	getProtocolConfig                 SuiMethod = "iota_getProtocolConfig" // TODO
	getTotalTransactionBlocks         SuiMethod = "iota_getTotalTransactionBlocks"
	getTransactionBlock               SuiMethod = "iota_getTransactionBlock"
	multiGetObjects                   SuiMethod = "iota_multiGetObjects"
	multiGetTransactionBlocks         SuiMethod = "iota_multiGetTransactionBlocks"
	tryGetPastObject                  SuiMethod = "iota_tryGetPastObject"
	tryMultiGetPastObjects            SuiMethod = "iota_tryMultiGetPastObjects"

	// Transaction Builder API
	batchTransaction     UnsafeMethod = "unsafe_batchTransaction"
	mergeCoins           UnsafeMethod = "unsafe_mergeCoins"
	moveCall             UnsafeMethod = "unsafe_moveCall"
	pay                  UnsafeMethod = "unsafe_pay"
	payAllSui            UnsafeMethod = "unsafe_payAllIota"
	paySui               UnsafeMethod = "unsafe_payIota"
	publish              UnsafeMethod = "unsafe_publish"
	requestAddStake      UnsafeMethod = "unsafe_requestAddStake"
	requestWithdrawStake UnsafeMethod = "unsafe_requestWithdrawStake"
	splitCoin            UnsafeMethod = "unsafe_splitCoin"
	splitCoinEqual       UnsafeMethod = "unsafe_splitCoinEqual"
	transferObject       UnsafeMethod = "unsafe_transferObject"
	transferSui          UnsafeMethod = "unsafe_transferIota"

	// Write API
	devInspectTransactionBlock SuiMethod = "iota_devInspectTransactionBlock"
	dryRunTransactionBlock     SuiMethod = "iota_dryRunTransactionBlock"
	executeTransactionBlock    SuiMethod = "iota_executeTransactionBlock"
)
