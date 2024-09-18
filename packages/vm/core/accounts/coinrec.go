package accounts

import (
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/sui-go/sui"
)

// CoinRecord represents a L1 Coin<T> object owned by the chain
type CoinRecord struct {
	ID       sui.ObjectID
	CoinType coin.Type `bcs:"-"` // transient
	Amount   coin.Value
}

func CoinRecordFromBytes(data []byte, coinType coin.Type) (*CoinRecord, error) {
	if coinType == "" {
		panic("unknown CoinType for CoinRecord")
	}

	rec, err := bcs.Unmarshal[CoinRecord](data)
	if err != nil {
		return nil, err
	}

	rec.CoinType = coinType

	return &rec, nil
}

func (rec *CoinRecord) Bytes() []byte {
	return bcs.MustMarshal(rec)
}
