package accounts

import (
	"io"
	"math/big"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/sui-go/sui"
)

// TreasuryCapRecord represents a L1 TreasuryCap<T> object owned by the chain
type TreasuryCapRecord struct {
	ID          sui.ObjectID
	CoinType    coin.Type // transient
	TotalSupply *big.Int
}

func (rec *TreasuryCapRecord) Bytes() []byte {
	return rwutil.WriteToBytes(rec)
}

func TreasuryCapRecordFromBytes(data []byte, coinType coin.Type) (*TreasuryCapRecord, error) {
	return rwutil.ReadFromBytes(data, &TreasuryCapRecord{CoinType: coinType})
}

func (rec *TreasuryCapRecord) Read(r io.Reader) error {
	if rec.CoinType == "" {
		panic("unknown TreasuryCap CoinType")
	}
	rr := rwutil.NewReader(r)
	rr.ReadN(rec.ID[:])
	rec.TotalSupply = rr.ReadBigUint()
	return rr.Err
}

func (rec *TreasuryCapRecord) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteN(rec.ID[:])
	ww.WriteBigUint(rec.TotalSupply)
	return ww.Err
}
