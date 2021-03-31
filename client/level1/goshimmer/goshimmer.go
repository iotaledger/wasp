package goshimmer

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/client"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
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
	_, err = api.goshimmerClient.SendFaucetRequest(targetAddress.Base58())
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
		bal, _ := out.Balances().Get(ledgerstate.ColorIOTA)
		total += uint64(bal)
	}
	return total, nil
}

func (api *goshimmerClient) GetConfirmedOutputs(address ledgerstate.Address) ([]ledgerstate.Output, error) {
	r, err := api.goshimmerClient.GetAddressUnspentOutputs(address.Base58())
	if err != nil {
		return nil, fmt.Errorf("GetUnspentOutputs: %w", err)
	}
	ret := make([]ledgerstate.Output, len(r.Outputs))
	for i, out := range r.Outputs {
		var err error
		ret[i], err = out.ToLedgerstateOutput()
		if err != nil {
			return nil, err
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
		tx, err := api.goshimmerClient.GetTransactionByID(txid.Base58())
		if err != nil {
			return err
		}
		if tx.InclusionState.Confirmed {
			break
		}
	}
	return nil
}
