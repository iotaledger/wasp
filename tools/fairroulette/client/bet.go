package client

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	nodeapi "github.com/iotaledger/goshimmer/dapps/waspconn/packages/apilib"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
)

func PlaceBet(goshimmerApi string, scAddress address.Address, color int, amount int, sigScheme signaturescheme.SignatureScheme) error {
	req := &waspapi.RequestBlockJson{
		Address:     scAddress.String(),
		RequestCode: fairroulette.RequestPlaceBet,
		AmountIotas: int64(amount),
		Vars: map[string]interface{}{
			fairroulette.ReqVarColor: int64(color),
		},
	}

	tx, err := waspapi.CreateRequestTransaction(goshimmerApi, sigScheme, []*waspapi.RequestBlockJson{req})
	if err != nil {
		return err
	}

	return nodeapi.PostTransaction(goshimmerApi, tx.Transaction)
}
