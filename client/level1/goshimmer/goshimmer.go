package goshimmer

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/client"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/plugins/webapi/value"
	"github.com/iotaledger/wasp/client/level1"
	"github.com/iotaledger/wasp/packages/parameters"
)

// NewGoshimmerClient returns a Level1Client that sends the requests to a Goshimmer node
func NewGoshimmerClient(goshimmerHost string) level1.Level1Client {
	return &goshimmerClient{client.NewGoShimmerAPI("http://" + goshimmerHost)}
}

type goshimmerClient struct {
	goshimmerClient *client.GoShimmerAPI
}

func (api *goshimmerClient) RequestFunds(targetAddress ledgerstate.Address) error {
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

func (api *goshimmerClient) balanceIOTA(targetAddress ledgerstate.Address) (uint64, error) {
	outs, err := api.GetConfirmedOutputs(targetAddress)
	if err != nil {
		return 0, fmt.Errorf("GetConfirmedOutputs: %s", err)
	}
	total := uint64(0)
	for _, out := range outs {
		for _, bal := range out.Balances {
			if bal.Color == ledgerstate.ColorIOTA.String() {
				total += uint64(bal.Value)
			}
		}
	}
	return total, nil
}

func (api *goshimmerClient) GetConfirmedOutputs(address ledgerstate.Address) ([]value.Output, error) {
	r, err := api.goshimmerClient.GetUnspentOutputs([]string{address.String()})
	if err != nil {
		return nil, fmt.Errorf("GetUnspentOutputs: %w", err)
	}
	if r.Error != "" {
		return nil, fmt.Errorf("%s", r.Error)
	}
	ret := make([]value.Output, 0)
	for _, out := range r.UnspentOutputs {
		for _, outid := range out.OutputIDs {
			if !outid.InclusionState.Confirmed {
				continue
			}
			id, err := ledgerstate.OutputIDFromBase58(outid.ID)
			if err != nil {
				return nil, fmt.Errorf("OutputIDFromBase58: %w", err)
			}
			txres, err := api.goshimmerClient.GetTransactionByID(id.TransactionID().Base58())
			if err != nil {
				return nil, fmt.Errorf("GetTransactionByID: %w", err)
			}
			if txres.InclusionState.Confirmed {
				continue
			}
			i := id.OutputIndex()
			if int(i) > len(txres.Transaction.Outputs)-1 {
				return nil, fmt.Errorf("can't find output with index %d", i)
			}
			ret = append(ret, txres.Transaction.Outputs[i])
		}
	}
	return ret, nil
}

func (api *goshimmerClient) sendTx(tx *ledgerstate.Transaction) error {
	data := tx.Bytes()
	if len(data) > parameters.MaxSerializedTransactionToGoshimmer {
		return fmt.Errorf("goshimmerClient: size of serialized transation %d bytes > max of %d bytes: %s",
			len(data), parameters.MaxSerializedTransactionToGoshimmer, tx.ID())
	}
	_, err := api.goshimmerClient.SendTransaction(data)
	return err
}

func (api *goshimmerClient) PostTransaction(tx *ledgerstate.Transaction) error {
	return api.sendTx(tx)
}

func (api *goshimmerClient) PostAndWaitForConfirmation(tx *ledgerstate.Transaction) error {
	if err := api.sendTx(tx); err != nil {
		return err
	}
	return api.WaitForConfirmation(tx.ID())
}

func (api *goshimmerClient) WaitForConfirmation(txid ledgerstate.TransactionID) error {
	for {
		time.Sleep(1 * time.Second)
		tx, err := api.goshimmerClient.GetTransactionByID(txid.String())
		if err != nil {
			return err
		}
		if tx.InclusionState.Confirmed {
			break
		}
	}
	return nil
}
