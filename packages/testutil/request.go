package testutil

import (
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	nodeapi "github.com/iotaledger/goshimmer/dapps/waspconn/packages/apilib"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
)

func RequestFunds(goshimmerHost string, targetAddress address.Address) error {
	// TODO: allow using "real" goshimmer API to request funds
	tx, err := makeUtxodbTransferTx(goshimmerHost, targetAddress)
	if err != nil {
		return err
	}

	return nodeapi.PostTransaction(goshimmerHost, tx)
}

func makeUtxodbTransferTx(goshimmerHost string, target address.Address) (*transaction.Transaction, error) {
	source := utxodb.GetAddress(1)
	sourceOutputs, err := nodeapi.GetAccountOutputs(goshimmerHost, &source)
	if err != nil {
		return nil, err
	}

	amount := int64(1337) // same as Faucet

	oids := make([]transaction.OutputID, 0)
	sum := int64(0)
	for oid, bals := range sourceOutputs {
		containsIotas := false
		for _, b := range bals {
			if b.Color == balance.ColorIOTA {
				sum += b.Value
				containsIotas = true
			}
		}
		if containsIotas {
			oids = append(oids, oid)
		}
		if sum >= amount {
			break
		}
	}

	if sum < amount {
		return nil, fmt.Errorf("not enough input balance")
	}

	inputs := transaction.NewInputs(oids...)

	out := make(map[address.Address][]*balance.Balance)
	out[target] = []*balance.Balance{balance.New(balance.ColorIOTA, amount)}
	if sum > amount {
		out[source] = []*balance.Balance{balance.New(balance.ColorIOTA, sum-amount)}
	}
	outputs := transaction.NewOutputs(out)

	tx := transaction.New(inputs, outputs)
	tx.Sign(utxodb.GetSigScheme(source))
	return tx, nil
}
