package nodeclient

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/client"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
)

func New(goshimmerHost string) NodeClient {
	return &nodeclient{client.NewGoShimmerAPI("http://" + goshimmerHost)}
}

type nodeclient struct {
	goshimmerClient *client.GoShimmerAPI
}

func (api *nodeclient) RequestFunds(targetAddress address.Address) error {
	_, err := api.goshimmerClient.SendFaucetRequest(targetAddress.String())
	return err
}

func (api *nodeclient) GetAccountOutputs(address *address.Address) (map[transaction.OutputID][]*balance.Balance, error) {
	r, err := api.goshimmerClient.GetUnspentOutputs([]string{address.String()})
	if err != nil {
		return nil, err
	}
	if r.Error != "" {
		return nil, fmt.Errorf("%s", r.Error)
	}
	ret := make(map[transaction.OutputID][]*balance.Balance)
	for _, out := range r.UnspentOutputs {
		for _, outid := range out.OutputIDs {
			id, err := transaction.OutputIDFromBase58(outid.ID)
			if err != nil {
				return nil, err
			}
			balances := make([]*balance.Balance, 0)
			for _, b := range outid.Balances {
				color, err := util.ColorFromString(b.Color)
				if err != nil {
					return nil, err
				}
				balances = append(balances, &balance.Balance{Value: b.Value, Color: color})
			}
			ret[id] = balances
		}
	}
	return ret, nil
}

func (api *nodeclient) PostAndWaitForConfirmation(tx *transaction.Transaction) error {
	txid, err := api.goshimmerClient.SendTransaction(tx.Bytes())
	if err != nil {
		return err
	}
	for {
		time.Sleep(1 * time.Second)
		tx, err := api.goshimmerClient.GetTransactionByID(txid)
		if err != nil {
			return err
		}
		if tx.InclusionState.Confirmed {
			break
		}
	}
	return nil
}
