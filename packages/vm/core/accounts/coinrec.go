package accounts

import (
	"io"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/sui-go/sui"
)

// CoinRecord represents a L1 Coin<T> object owned by the chain
type CoinRecord struct {
	ID       sui.ObjectID
	CoinType coin.Type // transient
	Amount   coin.Value
}

func CoinRecordFromBytes(data []byte, coinType coin.Type) (*CoinRecord, error) {
	return rwutil.ReadFromBytes(data, &CoinRecord{CoinType: coinType})
}

func (rec *CoinRecord) Bytes() []byte {
	return rwutil.WriteToBytes(rec)
}

func (rec *CoinRecord) Read(r io.Reader) error {
	if rec.CoinType == "" {
		panic("unknown CoinType for CoinRecord")
	}
	rr := rwutil.NewReader(r)
	rr.ReadN(rec.ID[:])
	rr.Read(&rec.Amount)
	return rr.Err
}

func (rec *CoinRecord) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteN(rec.ID[:])
	ww.Write(rec.Amount)
	return ww.Err
}
