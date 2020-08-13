package goshimmer

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/client"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/util"
)

func NewGoshimmerClient(goshimmerHost string) nodeclient.NodeClient {
	return &goshimmerClient{client.NewGoShimmerAPI("http://" + goshimmerHost)}
}

type goshimmerClient struct {
	goshimmerClient *client.GoShimmerAPI
}

func (api *goshimmerClient) RequestFunds(targetAddress *address.Address) error {
	initialBalance, err := api.balanceIOTA(targetAddress)
	if err != nil {
		return fmt.Errorf("balanceIOTA: %s", err)
	}
	_, err = api.goshimmerClient.SendFaucetRequest(targetAddress.String())
	if err != nil {
		return fmt.Errorf("SendFaucetRequest: %s", err)
	}
	for attempts := 10; attempts > 0; attempts-- {
		time.Sleep(1 * time.Second)
		balance, err := api.balanceIOTA(targetAddress)
		if err != nil {
			return fmt.Errorf("balanceIOTA: %s", err)
		}
		if balance > initialBalance {
			return nil
		}
	}
	return fmt.Errorf("Faucet request seems to have failed")
}

func (api *goshimmerClient) balanceIOTA(targetAddress *address.Address) (int64, error) {
	outs, err := api.GetAccountOutputs(targetAddress)
	if err != nil {
		return 0, fmt.Errorf("GetAccountOutputs: %s", err)
	}
	bals, _ := util.OutputBalancesByColor(outs)
	return bals[balance.ColorIOTA], nil
}

func (api *goshimmerClient) GetAccountOutputs(address *address.Address) (map[transaction.OutputID][]*balance.Balance, error) {
	r, err := api.goshimmerClient.GetUnspentOutputs([]string{address.String()})
	if err != nil {
		return nil, fmt.Errorf("GetUnspentOutputs: %s", err)
	}
	if r.Error != "" {
		return nil, fmt.Errorf("%s", r.Error)
	}
	ret := make(map[transaction.OutputID][]*balance.Balance)
	for _, out := range r.UnspentOutputs {
		for _, outid := range out.OutputIDs {
			id, err := transaction.OutputIDFromBase58(outid.ID)
			if err != nil {
				return nil, fmt.Errorf("OutputIDFromBase58: %s", err)
			}
			balances := make([]*balance.Balance, 0)
			for _, b := range outid.Balances {
				color, err := util.ColorFromString(b.Color)
				if err != nil {
					return nil, fmt.Errorf("ColorFromString: %s", err)
				}
				balances = append(balances, &balance.Balance{Value: b.Value, Color: color})
			}
			ret[id] = balances
		}
	}
	return ret, nil
}

func (api *goshimmerClient) PostTransaction(tx *transaction.Transaction) error {
	_, err := api.goshimmerClient.SendTransaction(tx.Bytes())
	return err
}

func (api *goshimmerClient) PostAndWaitForConfirmation(tx *transaction.Transaction) error {
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
