package testutil

import (
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	nodeapi "github.com/iotaledger/goshimmer/dapps/waspconn/packages/apilib"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
	"github.com/iotaledger/wasp/packages/nodeclient"
)

const RequestFundsAmount = utxodb.RequestFundsAmount

type utxodbclient struct {
	goshimmerHost string
}

func NewGoshimmerUtxodbClient(host string) nodeclient.NodeClient {
	return &utxodbclient{host}
}

func (api *utxodbclient) RequestFunds(targetAddress *address.Address) error {
	return nodeapi.RequestFunds(api.goshimmerHost, targetAddress)
}

func (api *utxodbclient) GetConfirmedAccountOutputs(address *address.Address) (map[transaction.OutputID][]*balance.Balance, error) {
	return nodeapi.GetAccountOutputs(api.goshimmerHost, address)
}

func (api *utxodbclient) PostTransaction(tx *transaction.Transaction) error {
	return nodeapi.PostTransaction(api.goshimmerHost, tx)
}

func (api *utxodbclient) PostAndWaitForConfirmation(tx *transaction.Transaction) error {
	err := nodeapi.PostTransaction(api.goshimmerHost, tx)
	if err != nil {
		return err
	}
	return api.WaitForConfirmation(tx.ID())
}

func (api *utxodbclient) WaitForConfirmation(txid transaction.ID) error {
	for {
		conf, err := nodeapi.IsConfirmed(api.goshimmerHost, &txid)
		if err != nil {
			return err
		}
		if conf {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}
