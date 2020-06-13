package wasptest

import (
	"fmt"
	nodeapi "github.com/iotaledger/goshimmer/packages/waspconn/apilib"
	"github.com/iotaledger/goshimmer/packages/waspconn/utxodb"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/tools/cluster"
	"time"
)

func SendRequests(clu *cluster.Cluster, sc *cluster.SmartContractFinalConfig, n int, wait time.Duration) error {
	for i := 0; i < n; i++ {
		tx, err := createRequestTx(clu.Config.Goshimmer.BindAddress, sc)
		if err != nil {
			return err
		}
		err = nodeapi.PostTransaction(clu.Config.Goshimmer.BindAddress, tx.Transaction)
		if err != nil {
			return err
		}
		fmt.Printf("[cluster] posted request txid %s", tx.ID().String())
		time.Sleep(wait)
	}
	return nil
}

func createRequestTx(node string, sc *cluster.SmartContractFinalConfig) (*sctransaction.Transaction, error) {
	sigScheme := utxodb.GetSigScheme(utxodb.GetAddress(sc.OwnerIndexUtxodb))
	reqBlock := &waspapi.RequestBlockJson{
		Address: sc.Address,
	}
	return waspapi.CreateRequestTransaction(node, sigScheme, []*waspapi.RequestBlockJson{reqBlock})
}
