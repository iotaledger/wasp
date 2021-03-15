package testutil

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxodb"
	"github.com/iotaledger/wasp/packages/parameters"
	"time"

	"github.com/iotaledger/wasp/client/level1"
)

const RequestFundsAmount = utxodb.RequestFundsAmount

type utxodbclient struct {
	goshimmerHost string
}

// NewGoshimmerUtxodbClient returns a Level1Client that uses the UTXODB interface.
//
// It requires a Goshimmer node configuerd in UTXODB mode in order to work.
func NewGoshimmerUtxodbClient(host string) level1.Level1Client {
	return &utxodbclient{host}
}

func (api *utxodbclient) RequestFunds(targetAddress ledgerstate.Address) error {
	return nodeapi.RequestFunds(api.goshimmerHost, targetAddress)
}

func (api *utxodbclient) GetConfirmedAccountOutputs(address ledgerstate.Address) (map[transaction.OutputID][]*balance.Balance, error) {
	return nodeapi.GetAccountOutputs(api.goshimmerHost, address)
}

func checkTxSize(tx *ledgerstate.Transaction) error {
	data := tx.Bytes()
	if len(data) > parameters.MaxSerializedTransactionToGoshimmer {
		return fmt.Errorf("utxodbclient: size of serialized transaction %d bytes > max of %d bytes: %s",
			len(data), parameters.MaxSerializedTransactionToGoshimmer, tx.ID())
	}
	return nil
}
func (api *utxodbclient) PostTransaction(tx *ledgerstate.Transaction) error {
	if err := checkTxSize(tx); err != nil {
		return err
	}
	return nodeapi.PostTransaction(api.goshimmerHost, tx)
}

func (api *utxodbclient) PostAndWaitForConfirmation(tx *ledgerstate.Transaction) error {
	if err := checkTxSize(tx); err != nil {
		return err
	}
	err := nodeapi.PostTransaction(api.goshimmerHost, tx)
	if err != nil {
		return err
	}
	return api.WaitForConfirmation(tx.ID())
}

func (api *utxodbclient) WaitForConfirmation(txid ledgerstate.TransactionID) error {
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
