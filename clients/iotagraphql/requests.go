package iotagraphql

import (
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql/graphqltypes"
)

type GetDynamicFieldObjectRequest struct {
	ParentObjectID *iotago.ObjectID
	Name           *iotago.DynamicFieldName
}

type GetDynamicFieldsRequest struct {
	ParentObjectID *iotago.ObjectID
	Cursor         *iotago.ObjectID // optional
	Limit          *int             // optional
}

type GetOwnedObjectsRequest struct {
	// Address is the owner's Iota address
	Address *iotago.Address
	// [optional] Query is the objects query criteria.
	Query *graphqltypes.IotaObjectResponseQuery
	// [optional] Cursor is an optional paging cursor.
	// If provided, the query will start from the next item after the specified cursor.
	Cursor *iotago.ObjectID
	// [optional] Limit is the maximum number of items returned per page, defaults to [QUERY_MAX_RESULT_LIMIT_OBJECTS] if not
	// provided
	Limit *int
}

type QueryEventsRequest struct {
	Query           *graphqltypes.EventFilter
	Cursor          *graphqltypes.EventID // optional
	Limit           *int                  // optional
	DescendingOrder bool                  // optional
}

type QueryTransactionBlocksRequest struct {
	Query           *graphqltypes.IotaTransactionBlockResponseQuery
	Cursor          *iotago.TransactionDigest // optional
	Limit           *int                      // optional
	DescendingOrder bool                      // optional
}

type ResolveNameServiceNamesRequest struct {
	Owner  *iotago.Address
	Cursor *iotago.ObjectID // optional
	Limit  *int             // optional
}

type DryRunTransactionRequest struct {
	TxDataBytes iotago.Base64Data
	Options     *graphqltypes.IotaTransactionBlockResponseOptions // optional
}

type ExecuteTransactionBlockRequest struct {
	TxDataBytes       iotago.Base64Data
	Signatures        []*iotasigner.Signature
	Options           *graphqltypes.IotaTransactionBlockResponseOptions // optional
	ExecutionWaitMode graphqltypes.ExecutionWaitMode                    // optional
}

type BatchTransactionRequest struct {
	Signer    *iotago.Address
	TxnParams []map[string]interface{}
	Gas       *iotago.ObjectID // optional
	GasBudget uint64
}

type MergeCoinsRequest struct {
	Signer      *iotago.Address
	PrimaryCoin *iotago.ObjectID
	CoinToMerge *iotago.ObjectID
	Gas         *iotago.ObjectID // optional
	GasBudget   *graphqltypes.BigInt
}

type MoveCallRequest struct {
	Signer    *iotago.Address
	PackageID *iotago.PackageID
	Module    string
	Function  string
	TypeArgs  []string
	Arguments []any
	Gas       *iotago.ObjectID // optional
	GasBudget *graphqltypes.BigInt
}

type PayRequest struct {
	Signer     *iotago.Address
	InputCoins []*iotago.ObjectID
	Recipients []*iotago.Address
	Amount     []*graphqltypes.BigInt
	Gas        *iotago.ObjectID // optional
	GasBudget  *graphqltypes.BigInt
}

type PayAllIotaRequest struct {
	Signer     *iotago.Address
	Recipient  *iotago.Address
	InputCoins []*iotago.ObjectID
	GasBudget  *graphqltypes.BigInt
}

type PayIotaRequest struct {
	Signer     *iotago.Address
	InputCoins []*iotago.ObjectID
	Recipients []*iotago.Address
	Amount     []*graphqltypes.BigInt
	GasBudget  *graphqltypes.BigInt
}

type PublishRequest struct {
	Sender          *iotago.Address
	CompiledModules []*iotago.Base64Data
	Dependencies    []*iotago.ObjectID
	Gas             *iotago.ObjectID // optional
	GasBudget       *graphqltypes.BigInt
}

type RequestAddStakeRequest struct {
	Signer    *iotago.Address
	Coins     []*iotago.ObjectID
	Amount    *graphqltypes.BigInt // optional
	Validator *iotago.Address
	Gas       *iotago.ObjectID // optional
	GasBudget *graphqltypes.BigInt
}

type RequestWithdrawStakeRequest struct {
	Signer       *iotago.Address
	StakedIotaID *iotago.ObjectID
	Gas          *iotago.ObjectID // optional
	GasBudget    *graphqltypes.BigInt
}

type SplitCoinRequest struct {
	Signer       *iotago.Address
	Coin         *iotago.ObjectID
	SplitAmounts []*graphqltypes.BigInt
	Gas          *iotago.ObjectID // optional
	GasBudget    *graphqltypes.BigInt
}

type SplitCoinEqualRequest struct {
	Signer     *iotago.Address
	Coin       *iotago.ObjectID
	SplitCount *graphqltypes.BigInt
	Gas        *iotago.ObjectID // optional
	GasBudget  *graphqltypes.BigInt
}

type TransferObjectRequest struct {
	Signer    *iotago.Address
	ObjectID  *iotago.ObjectID
	Gas       *iotago.ObjectID // optional
	GasBudget *graphqltypes.BigInt
	Recipient *iotago.Address
}

type TransferIotaRequest struct {
	Signer    *iotago.Address
	ObjectID  *iotago.ObjectID
	GasBudget *graphqltypes.BigInt
	Recipient *iotago.Address
	Amount    *graphqltypes.BigInt // optional
}

type GetAllCoinsRequest struct {
	Owner  *iotago.Address
	Cursor *iotago.ObjectID // optional
	Limit  int              // optional
}

type GetBalanceRequest struct {
	Owner    *iotago.Address
	CoinType string // optional
}

type GetCoinsRequest struct {
	Owner    *iotago.Address
	CoinType *string // optional
	Cursor   *string // optional
	Limit    int     // optional
}

type GetCheckpointsRequest struct {
	Cursor          *graphqltypes.BigInt // optional
	Limit           *uint64              // optional
	DescendingOrder bool
}

type GetObjectRequest struct {
	ObjectID *iotago.ObjectID
	Options  *graphqltypes.IotaObjectDataOptions // optional
}

type GetTransactionBlockRequest struct {
	Digest  *iotago.TransactionDigest
	Options *graphqltypes.IotaTransactionBlockResponseOptions // optional
}

type MultiGetObjectsRequest struct {
	ObjectIDs []*iotago.ObjectID
	Options   *graphqltypes.IotaObjectDataOptions // optional
}

type MultiGetTransactionBlocksRequest struct {
	Digests []*iotago.Digest
	Options *graphqltypes.IotaTransactionBlockResponseOptions // optional
}

type TryGetPastObjectRequest struct {
	ObjectID *iotago.ObjectID
	Version  uint64
	Options  *graphqltypes.IotaObjectDataOptions // optional
}

type TryMultiGetPastObjectsRequest struct {
	PastObjects []*graphqltypes.IotaGetPastObjectRequest
	Options     *graphqltypes.IotaObjectDataOptions // optional
}

type SignAndExecuteTransactionRequest struct {
	TxDataBytes       iotago.Base64Data
	Signer            iotasigner.Signer
	Options           *graphqltypes.IotaTransactionBlockResponseOptions // optional
	ExecutionWaitMode graphqltypes.ExecutionWaitMode                    // optional, defaults to WaitForLocalExecution
}
