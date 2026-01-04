package iotagraphql

import (
	api "github.com/iotaledger/wasp/v2/clients/iota-go/client"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
)

// Re-export request/utility types so callers can stay within the iotagraphql package.
type (
	GetDynamicFieldObjectRequest      = api.GetDynamicFieldObjectRequest
	GetDynamicFieldsRequest           = api.GetDynamicFieldsRequest
	GetOwnedObjectsRequest            = api.GetOwnedObjectsRequest
	QueryEventsRequest                = api.QueryEventsRequest
	QueryTransactionBlocksRequest     = api.QueryTransactionBlocksRequest
	ResolveNameServiceNamesRequest    = api.ResolveNameServiceNamesRequest
	DevInspectTransactionBlockRequest = api.DevInspectTransactionBlockRequest
	DryRunTransactionRequest          = api.DryRunTransactionRequest
	ExecuteTransactionBlockRequest    = api.ExecuteTransactionBlockRequest
	BatchTransactionRequest           = api.BatchTransactionRequest
	MergeCoinsRequest                 = api.MergeCoinsRequest
	MoveCallRequest                   = api.MoveCallRequest
	PayRequest                        = api.PayRequest
	PayAllIotaRequest                 = api.PayAllIotaRequest
	PayIotaRequest                    = api.PayIotaRequest
	PublishRequest                    = api.PublishRequest
	RequestAddStakeRequest            = api.RequestAddStakeRequest
	RequestWithdrawStakeRequest       = api.RequestWithdrawStakeRequest
	SplitCoinRequest                  = api.SplitCoinRequest
	SplitCoinEqualRequest             = api.SplitCoinEqualRequest
	TransferObjectRequest             = api.TransferObjectRequest
	TransferIotaRequest               = api.TransferIotaRequest
	GetAllCoinsRequest                = api.GetAllCoinsRequest
	GetBalanceRequest                 = api.GetBalanceRequest
	GetCoinsRequest                   = api.GetCoinsRequest
	GetCheckpointsRequest             = api.GetCheckpointsRequest
	GetObjectRequest                  = api.GetObjectRequest
	GetTransactionBlockRequest        = api.GetTransactionBlockRequest
	MultiGetObjectsRequest            = api.MultiGetObjectsRequest
	MultiGetTransactionBlocksRequest  = api.MultiGetTransactionBlocksRequest
	TryGetPastObjectRequest           = api.TryGetPastObjectRequest
	TryMultiGetPastObjectsRequest     = api.TryMultiGetPastObjectsRequest
	SignAndExecuteTransactionRequest  = api.SignAndExecuteTransactionRequest
	WaitParams                        = api.WaitParams
	RetryCondition[T any]             = api.RetryCondition[T]
)

// Re-export iotajsonrpc types for backward compatibility with test files
type (
	IotaTransactionBlockResponseOptions = iotajsonrpc.IotaTransactionBlockResponseOptions
	IotaTransactionBlockResponseQuery   = iotajsonrpc.IotaTransactionBlockResponseQuery
	IotaTransactionBlockResponse        = iotajsonrpc.IotaTransactionBlockResponse
	IotaObjectDataOptions               = iotajsonrpc.IotaObjectDataOptions
	IotaObjectResponseQuery             = iotajsonrpc.IotaObjectResponseQuery
	IotaObjectDataFilter                = iotajsonrpc.IotaObjectDataFilter
	IotaObjectResponse                  = iotajsonrpc.IotaObjectResponse
	IotaPastObjectResponse              = iotajsonrpc.IotaPastObjectResponse
	IotaSystemStateSummary              = iotajsonrpc.IotaSystemStateSummary
	IotaCoinMetadata                    = iotajsonrpc.IotaCoinMetadata
	BigInt                              = iotajsonrpc.BigInt
	CoinType                            = iotajsonrpc.CoinType
	Coins                               = iotajsonrpc.Coins
	CoinPage                            = iotajsonrpc.CoinPage
	Balance                        = iotajsonrpc.Balance
	Supply                         = iotajsonrpc.Supply
	DevInspectResults              = iotajsonrpc.DevInspectResults
	DynamicFieldPage               = iotajsonrpc.DynamicFieldPage
	ObjectsPage                    = iotajsonrpc.ObjectsPage
	TransactionBlocksPage          = iotajsonrpc.TransactionBlocksPage
	TransactionBytes               = iotajsonrpc.TransactionBytes
	EventId                                   = iotajsonrpc.EventId
	ExecuteTransactionRequestType             = iotajsonrpc.ExecuteTransactionRequestType
	CoinValue                                 = iotajsonrpc.CoinValue
	ProgrammableTransactionBlockPureInput     = iotajsonrpc.ProgrammableTransactionBlockPureInput
	IotaTransactionBlockData                  = iotajsonrpc.IotaTransactionBlockData
	IotaTransactionBlockEffects               = iotajsonrpc.IotaTransactionBlockEffects
	IotaEvent                                 = iotajsonrpc.IotaEvent
	ObjectChange                              = iotajsonrpc.ObjectChange
	BalanceChange                             = iotajsonrpc.BalanceChange
	IotaObjectData                            = iotajsonrpc.IotaObjectData
	Coin                                      = iotajsonrpc.Coin
	TransactionFilter                         = iotajsonrpc.TransactionFilter
	// Note: DryRunTransactionBlockResponse from generated.go is different from iotajsonrpc
	// Use DryRunResult alias for the iotajsonrpc version to avoid conflicts
	DryRunResult                              = iotajsonrpc.DryRunTransactionBlockResponse
	GasCostSummary                            = iotajsonrpc.GasCostSummary
	PickedCoins                               = iotajsonrpc.PickedCoins
)

const (
	DefaultGasBudget = api.DefaultGasBudget
	DefaultGasPrice  = api.DefaultGasPrice
	MinGasBudget     = api.MinGasBudget
	MaxGasBudget     = api.MaxGasBudget
)

// Re-export iotajsonrpc coin picking method constants
const (
	PickMethodSmaller = iotajsonrpc.PickMethodSmaller
	PickMethodBigger  = iotajsonrpc.PickMethodBigger
	PickMethodByOrder = iotajsonrpc.PickMethodByOrder
)

var (
	WaitForEffectsDisabled = api.WaitForEffectsDisabled
	WaitForEffectsEnabled  = api.WaitForEffectsEnabled
)

// Re-export iotajsonrpc constants
var (
	IotaCoinType                        = iotajsonrpc.IotaCoinType
	TxnRequestTypeWaitForLocalExecution = iotajsonrpc.TxnRequestTypeWaitForLocalExecution
)

// Re-export iotajsonrpc functions
var (
	NewBigInt                = iotajsonrpc.NewBigInt
	NewBigIntInt64           = iotajsonrpc.NewBigIntInt64
	PickupCoins              = iotajsonrpc.PickupCoins
	PickupCoinsWithFilter    = iotajsonrpc.PickupCoinsWithFilter
	PickupCoinWithFilter     = iotajsonrpc.PickupCoinWithFilter
	PickupCoinsWithCointype  = iotajsonrpc.PickupCoinsWithCointype
	MustCoinTypeFromString   = iotajsonrpc.MustCoinTypeFromString
	CoinTypeFromString       = iotajsonrpc.CoinTypeFromString
)

