package wasptest

import (
	"fmt"
	nodeapi "github.com/iotaledger/goshimmer/packages/waspconn/apilib"
	"github.com/iotaledger/goshimmer/packages/waspconn/utxodb"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/tools/cluster"
)

func Send1Request(clu *cluster.Cluster, sc *cluster.SmartContractFinalConfig) error {
	fmt.Printf("---------------------   Test: create origin of 1 smart contract and send 1 request to it\n")

	tx, err := createRequestTx(clu.Config.Goshimmer.BindAddress, sc)
	if err != nil {
		return err
	}
	err = nodeapi.PostTransaction(clu.Config.Goshimmer.BindAddress, tx.Transaction)
	if err != nil {
		return err
	}
	fmt.Printf("[cluster] posted request txid %s", tx.ID().String())
	return nil
}

func createRequestTx(node string, sc *cluster.SmartContractFinalConfig) (*sctransaction.Transaction, error) {
	sigScheme := utxodb.GetSigScheme(utxodb.GetAddress(sc.OwnerIndexUtxodb))
	reqBlock := &apilib.RequestBlockJson{
		Address: sc.Address,
	}
	return apilib.CreateRequestTransaction(node, sigScheme, []*apilib.RequestBlockJson{reqBlock})
}
