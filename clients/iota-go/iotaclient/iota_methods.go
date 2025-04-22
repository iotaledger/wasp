package iotaclient

type IotaMethod string

func (s IotaMethod) String() string {
	return string(s)
}

type IotaXMethod string

func (s IotaXMethod) String() string {
	return string(s)
}

type UnsafeMethod string

func (u UnsafeMethod) String() string {
	return string(u)
}

const (
	// Coin Query API
	getAllBalances  IotaXMethod = "iotax_getAllBalances"
	getAllCoins     IotaXMethod = "iotax_getAllCoins"
	getBalance      IotaXMethod = "iotax_getBalance"
	getCoinMetadata IotaXMethod = "iotax_getCoinMetadata"
	getCoins        IotaXMethod = "iotax_getCoins"
	getTotalSupply  IotaXMethod = "iotax_getTotalSupply"

	// Extended API
	getDynamicFieldObject     IotaXMethod = "iotax_getDynamicFieldObject"
	getDynamicFields          IotaXMethod = "iotax_getDynamicFields"
	getOwnedObjects           IotaXMethod = "iotax_getOwnedObjects"
	queryEvents               IotaXMethod = "iotax_queryEvents"
	queryTransactionBlocks    IotaXMethod = "iotax_queryTransactionBlocks"
	resolveNameServiceAddress IotaXMethod = "iotax_resolveNameServiceAddress"
	resolveNameServiceNames   IotaXMethod = "iotax_resolveNameServiceNames"
	subscribeEvent            IotaXMethod = "iotax_subscribeEvent"
	subscribeTransaction      IotaXMethod = "iotax_subscribeTransaction"

	// Governance Read API
	getCommitteeInfo         IotaXMethod = "iotax_getCommitteeInfo"
	getLatestIotaSystemState IotaXMethod = "iotax_getLatestIotaSystemState"
	getReferenceGasPrice     IotaXMethod = "iotax_getReferenceGasPrice"
	getStakes                IotaXMethod = "iotax_getStakes"
	getStakesByIDs           IotaXMethod = "iotax_getStakesByIds"
	getValidatorsApy         IotaXMethod = "iotax_getValidatorsApy"

	// Move Utils
	getMoveFunctionArgTypes           IotaMethod = "iota_getMoveFunctionArgTypes"           // TODO
	getNormalizedMoveFunction         IotaMethod = "iota_getNormalizedMoveFunction"         // TODO
	getNormalizedMoveModule           IotaMethod = "iota_getNormalizedMoveModule"           // TODO
	getNormalizedMoveModulesByPackage IotaMethod = "iota_getNormalizedMoveModulesByPackage" // TODO
	getNormalizedMoveStruct           IotaMethod = "iota_getNormalizedMoveStruct"           // TODO

	// Read API
	getChainIdentifier                IotaMethod = "iota_getChainIdentifier"
	getCheckpoint                     IotaMethod = "iota_getCheckpoint"
	getCheckpoints                    IotaMethod = "iota_getCheckpoints"
	getEvents                         IotaMethod = "iota_getEvents"
	getLatestCheckpointSequenceNumber IotaMethod = "iota_getLatestCheckpointSequenceNumber"
	getLoadedChildObjects             IotaMethod = "iota_getLoadedChildObjects"
	getObject                         IotaMethod = "iota_getObject"
	getProtocolConfig                 IotaMethod = "iota_getProtocolConfig"
	getTotalTransactionBlocks         IotaMethod = "iota_getTotalTransactionBlocks"
	getTransactionBlock               IotaMethod = "iota_getTransactionBlock"
	multiGetObjects                   IotaMethod = "iota_multiGetObjects"
	multiGetTransactionBlocks         IotaMethod = "iota_multiGetTransactionBlocks"
	tryGetPastObject                  IotaMethod = "iota_tryGetPastObject"
	tryMultiGetPastObjects            IotaMethod = "iota_tryMultiGetPastObjects"

	// Transaction Builder API
	batchTransaction     UnsafeMethod = "unsafe_batchTransaction"
	mergeCoins           UnsafeMethod = "unsafe_mergeCoins"
	moveCall             UnsafeMethod = "unsafe_moveCall"
	pay                  UnsafeMethod = "unsafe_pay"
	payAllIota           UnsafeMethod = "unsafe_payAllIota"
	payIota              UnsafeMethod = "unsafe_payIota"
	publish              UnsafeMethod = "unsafe_publish"
	requestAddStake      UnsafeMethod = "unsafe_requestAddStake"
	requestWithdrawStake UnsafeMethod = "unsafe_requestWithdrawStake"
	splitCoin            UnsafeMethod = "unsafe_splitCoin"
	splitCoinEqual       UnsafeMethod = "unsafe_splitCoinEqual"
	transferObject       UnsafeMethod = "unsafe_transferObject"
	transferIota         UnsafeMethod = "unsafe_transferIota"

	// Write API
	devInspectTransactionBlock IotaMethod = "iota_devInspectTransactionBlock"
	dryRunTransactionBlock     IotaMethod = "iota_dryRunTransactionBlock"
	executeTransactionBlock    IotaMethod = "iota_executeTransactionBlock"
)
