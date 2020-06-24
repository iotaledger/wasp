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

func SendRequests(clu *cluster.Cluster, sc *cluster.SmartContractFinalConfig, n int, code sctransaction.RequestCode, args map[string]string, wait time.Duration) error {
	for i := 0; i < n; i++ {
		// in real situation one must wait until the previous request is confirmed
		// (because of access to the same owner address)
		tx, err := createRequestTx(clu.Config.Goshimmer.BindAddress, sc, code, args)
		if err != nil {
			return err
		}
		fmt.Printf("[cluster] created request tx: %s\n", tx.String())

		err = nodeapi.PostTransaction(clu.Config.Goshimmer.BindAddress, tx.Transaction)
		if err != nil {
			return err
		}
		//fmt.Printf("[cluster] posted request txid %s\n", tx.ID().String())
		time.Sleep(wait)
	}
	return nil
}

func createRequestTx(node string, sc *cluster.SmartContractFinalConfig, code sctransaction.RequestCode, args map[string]string) (*sctransaction.Transaction, error) {
	sigScheme := utxodb.GetSigScheme(utxodb.GetAddress(sc.OwnerIndexUtxodb))
	reqBlock := &waspapi.RequestBlockJson{
		RequestCode: uint16(code),
		Address:     sc.Address,
		Vars:        args,
	}
	return waspapi.CreateRequestTransaction(node, sigScheme, []*waspapi.RequestBlockJson{reqBlock})
}
