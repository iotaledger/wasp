package models

import "github.com/iotaledger/wasp/sui-go/sui_types"

type GetAllCoinsRequest struct {
	Owner  *sui_types.SuiAddress
	Cursor *sui_types.ObjectID // optional
	Limit  uint                // optional
}

type GetBalanceRequest struct {
	Owner    *sui_types.SuiAddress
	CoinType sui_types.ObjectType // optional
}

type GetCoinsRequest struct {
	Owner    *sui_types.SuiAddress
	CoinType *sui_types.ObjectType // optional
	Cursor   *sui_types.ObjectID   // optional
	Limit    uint                  // optional
}
