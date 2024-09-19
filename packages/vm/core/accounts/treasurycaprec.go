package accounts

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/sui-go/sui"
)

// TreasuryCapRecord represents a L1 TreasuryCap<T> object owned by the chain
type TreasuryCapRecord struct {
	ID          sui.ObjectID
	CoinType    coin.Type `bcs:"-"` // transient
	TotalSupply *big.Int
}

func (rec *TreasuryCapRecord) Bytes() []byte {
	return bcs.MustMarshal(rec)
}

func TreasuryCapRecordFromBytes(data []byte, coinType coin.Type) (*TreasuryCapRecord, error) {
	return bcs.UnmarshalInto(data, &TreasuryCapRecord{CoinType: coinType})
}
