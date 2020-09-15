package goshimmer

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/iotaledger/goshimmer/client"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/util"
)

func NewGoshimmerClient(goshimmerHost string) nodeclient.NodeClient {
	fmt.Printf("using Goshimmer host %s\n", goshimmerHost)
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

func (api *goshimmerClient) GetAccountOutputs(address *address.Address) (map[valuetransaction.OutputID][]*balance.Balance, error) {
	r, err := api.goshimmerClient.GetUnspentOutputs([]string{address.String()})
	if err != nil {
		return nil, fmt.Errorf("GetUnspentOutputs: %s", err)
	}
	if r.Error != "" {
		return nil, fmt.Errorf("%s", r.Error)
	}
	ret := make(map[valuetransaction.OutputID][]*balance.Balance)
	for _, out := range r.UnspentOutputs {
		for _, outid := range out.OutputIDs {
			if !outid.InclusionState.Confirmed {
				continue
			}
			id, err := valuetransaction.OutputIDFromBase58(outid.ID)
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

func (api *goshimmerClient) PostTransaction(tx *valuetransaction.Transaction) error {
	_, err := api.goshimmerClient.SendTransaction(tx.Bytes())
	return err
}

func (api *goshimmerClient) PostAndWaitForConfirmation(tx *valuetransaction.Transaction) error {
	_, err := api.goshimmerClient.SendTransaction(tx.Bytes())
	if err != nil {
		return err
	}
	return api.WaitForConfirmation(tx.ID())
}

func (api *goshimmerClient) WaitForConfirmation(txid valuetransaction.ID) error {
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

//goland:noinspection ALL
func (api *goshimmerClient) PrintTransactionById(txidBase58 string, outText ...io.Writer) {
	var out io.Writer
	out = os.Stdout
	if len(outText) != 0 {
		if outText[0] == nil {
			out = os.Stdout
		} else {
			out = outText[0]
		}
	}
	resp, err := api.goshimmerClient.GetTransactionByID(txidBase58)
	if err != nil {
		fmt.Fprintf(out, "error while querying transaction %s: %v", txidBase58, err)
		return
	}

	fmt.Fprintf(out, "-- Transaction: %s\n", os.Args[1])
	fmt.Fprintf(out, "-- Data payload: %d bytes\n", len(resp.Transaction.DataPayload))
	fmt.Fprintf(out, "-- Inputs:\n")
	for _, inp := range resp.Transaction.Inputs {
		fmt.Fprintf(out, "    %s\n", inp)
	}
	fmt.Fprintf(out, "-- Outputs:\n")
	for _, outp := range resp.Transaction.Outputs {
		fmt.Fprintf(out, "    Address: %s\n", outp.Address)
		for _, bal := range outp.Balances {
			fmt.Fprintf(out, "        %s: %d\n", bal.Color, bal.Value)
		}
	}
	fmt.Fprintf(out, "-- Inclusion state:\n    %+v\n", resp.InclusionState)
	fmt.Fprintf(out, "-- Error:\n%s\n", resp.Error)
}
