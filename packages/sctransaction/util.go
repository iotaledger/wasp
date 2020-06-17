package sctransaction

import (
	"errors"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/util"
	"io"
)

func ReadRequestId(r io.Reader, reqid *RequestId) error {
	n, err := r.Read(reqid[:])
	if err != nil {
		return err
	}
	if n != RequestIdSize {
		return errors.New("error while reading request id")
	}
	return nil
}

func BalanceOfOutputToColor(tx *Transaction, addr address.Address, color balance.Color) int64 {
	bals, ok := tx.Outputs().Get(addr)
	if !ok {
		return 0
	}

	return util.BalanceOfColor(bals.([]*balance.Balance), color)
}
