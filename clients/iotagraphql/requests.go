package iotagraphql

import (
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql/graphqltypes"
)

type GetDynamicFieldObjectRequest struct {
	ParentObjectID iotago.ObjectID
	Name           iotago.DynamicFieldName
}

type GetDynamicFieldsRequest struct {
	ParentObjectID iotago.ObjectID
	Cursor         *string // optional, opaque GraphQL cursor
	Limit          *int    // optional
}

type GetOwnedObjectsRequest struct {
	Address iotago.Address
	Filter  *graphqltypes.ObjectFilter // optional
	Cursor  *string                    // optional, opaque GraphQL cursor
	Limit   *int                       // optional
}

type PayAllIotaRequest struct {
	Signer     iotago.Address
	Recipient  iotago.Address
	InputCoins []iotago.ObjectID
	GasBudget  *graphqltypes.BigInt
}

type PayIotaRequest struct {
	Signer     iotago.Address
	InputCoins []iotago.ObjectID
	Recipients []*iotago.Address
	Amount     []*graphqltypes.BigInt
	GasBudget  *graphqltypes.BigInt
}

type PublishRequest struct {
	Sender          iotago.Address
	CompiledModules []*iotago.Base64Data
	Dependencies    []*iotago.ObjectID
	Gas             *iotago.ObjectID // optional
	GasBudget       *graphqltypes.BigInt
}

type TransferObjectRequest struct {
	Signer    iotago.Address
	ObjectID  iotago.ObjectID
	Gas       *iotago.ObjectID // optional
	GasBudget *graphqltypes.BigInt
	Recipient iotago.Address
}

type GetBalanceRequest struct {
	Owner    iotago.Address
	CoinType graphqltypes.CoinType // optional
}

type GetCoinsRequest struct {
	Owner    iotago.Address
	CoinType *graphqltypes.CoinType // optional
	Cursor   *string                // optional
	Limit    int                    // optional
}
