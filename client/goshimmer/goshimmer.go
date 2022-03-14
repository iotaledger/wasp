package goshimmer

import (
	"fmt"
	"strings"
	"time"

	"github.com/iotaledger/goshimmer/client"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/transaction"
)

// Client is a wrapper for the official Goshimmer client, providing a cleaner interface
// for commonly used Goshimmer webapi endpoints in wasp.
type Client struct {
	api             *client.GoShimmerAPI
	faucetPowTarget int
}

// NewClient returns a new Goshimmer client
func NewClient(goshimmerHost string, faucetPowTarget int) *Client {
	if !strings.HasPrefix(goshimmerHost, "http") {
		goshimmerHost = "http://" + goshimmerHost
	}
	return &Client{
		api:             client.NewGoShimmerAPI(goshimmerHost),
		faucetPowTarget: faucetPowTarget,
	}
}

func (c *Client) RequestFunds(targetAddress ledgerstate.Address) error {
	initialBalance, err := c.BalanceIOTA(targetAddress)
	if err != nil {
		return fmt.Errorf("BalanceIOTA: %s", err)
	}
	_, err = c.api.SendFaucetRequest(targetAddress.Base58(), c.faucetPowTarget)
	if err != nil {
		return fmt.Errorf("SendFaucetRequest: %s", err)
	}
	for attempts := 10; attempts > 0; attempts-- {
		time.Sleep(1 * time.Second)
		balance, err := c.BalanceIOTA(targetAddress)
		if err != nil {
			return fmt.Errorf("BalanceIOTA: %s", err)
		}
		if balance > initialBalance {
			return nil
		}
	}
	return fmt.Errorf("Faucet request seems to have failed")
}

func (c *Client) BalanceIOTA(targetAddress ledgerstate.Address) (uint64, error) {
	outs, err := c.GetConfirmedOutputs(targetAddress)
	if err != nil {
		return 0, fmt.Errorf("GetConfirmedOutputs: %s", err)
	}
	total := uint64(0)
	for _, out := range outs {
		bal, _ := out.Balances().Get(ledgerstate.ColorIOTA)
		total += bal
	}
	return total, nil
}

func (c *Client) GetConfirmedOutputs(address ledgerstate.Address) ([]ledgerstate.Output, error) {
	r, err := c.api.GetAddressUnspentOutputs(address.Base58())
	if err != nil {
		return nil, fmt.Errorf("GetUnspentOutputs: %w", err)
	}

	// prevent calling c.IsTransactionConfirmed() twice for the same tx
	confirmedCache := map[string]bool{}
	isConfirmed := func(txID string) (bool, error) {
		confirmed, ok := confirmedCache[txID]
		if ok {
			return confirmed, nil
		}
		confirmed, err = c.IsTransactionConfirmed(txID)
		if err != nil {
			return false, err
		}
		confirmedCache[txID] = confirmed
		return confirmed, nil
	}

	var ret []ledgerstate.Output
	for _, out := range r.Outputs {
		var err error
		confirmed, err := isConfirmed(out.OutputID.TransactionID)
		if err != nil {
			return nil, err
		}
		if !confirmed {
			continue
		}
		output, err := out.ToLedgerstateOutput()
		if err != nil {
			return nil, err
		}
		ret = append(ret, output)
	}
	return ret, nil
}

func (c *Client) IsTransactionConfirmed(txID string) (bool, error) {
	r, err := c.api.GetTransactionInclusionState(txID)
	if err != nil {
		return false, fmt.Errorf("IsTransactionConfirmed: %w", err)
	}
	return r.Confirmed && !r.Rejected, nil
}

func (c *Client) postTx(tx *ledgerstate.Transaction) error {
	data := tx.Bytes()
	if len(data) > parameters.MaxSerializedTransactionToGoshimmer {
		return fmt.Errorf("size of serialized transaction %d bytes > max of %d bytes: %s",
			len(data), parameters.MaxSerializedTransactionToGoshimmer, tx.ID())
	}
	_, err := c.api.PostTransaction(data)
	return err
}

func (c *Client) PostTransaction(tx *ledgerstate.Transaction) error {
	return c.postTx(tx)
}

func (c *Client) PostAndWaitForConfirmation(tx *ledgerstate.Transaction) error {
	if err := c.postTx(tx); err != nil {
		return err
	}
	return c.WaitForConfirmation(tx.ID())
}

func (c *Client) WaitForConfirmation(txid ledgerstate.TransactionID) error {
	for {
		time.Sleep(1 * time.Second)
		state, err := c.api.GetTransactionInclusionState(txid.Base58())
		if err != nil {
			return err
		}
		if state.Confirmed {
			break
		}
	}
	return nil
}

func (c *Client) PostRequestTransaction(par transaction.NewRequestTransactionParams) (*ledgerstate.Transaction, error) {
	var err error

	if len(par.UnspentOutputs) == 0 {
		addr := ledgerstate.NewED25519Address(par.SenderKeyPair.PublicKey)
		par.UnspentOutputs, err = c.GetConfirmedOutputs(addr)
		if err != nil {
			return nil, fmt.Errorf("can't get outputs from the node: %v", err)
		}
	}

	tx, err := transaction.NewRequestTransaction(par)
	if err != nil {
		return nil, err
	}

	if err := c.PostTransaction(tx); err != nil {
		return nil, err
	}
	return tx, nil
}
