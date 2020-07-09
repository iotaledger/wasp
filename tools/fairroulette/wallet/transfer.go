package wallet

import (
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	nodeapi "github.com/iotaledger/goshimmer/dapps/waspconn/packages/apilib"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
	"github.com/iotaledger/wasp/tools/fairroulette/config"
)

func transfer(utxodbIndex int, amount int) {
	walletAddress := Load().Address()
	check(nodeapi.PostTransaction(
		config.GoshimmerApi(),
		makeTransferTx(walletAddress, utxodbIndex, int64(amount)),
	))
}

func makeTransferTx(target address.Address, utxodbIndex int, amount int64) *transaction.Transaction {
	source := utxodb.GetAddress(utxodbIndex)
	sourceOutputs, err := nodeapi.GetAccountOutputs(config.GoshimmerApi(), &source)
	check(err)

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
		panic(fmt.Errorf("not enough input balance"))
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
	return tx
}
