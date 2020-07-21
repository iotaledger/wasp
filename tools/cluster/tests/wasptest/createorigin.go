package wasptest

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
	"github.com/iotaledger/wasp/tools/cluster"
)

const goshimmerDirectly = true

func CreateOrigin1SC(clu *cluster.Cluster, sc *cluster.SmartContractFinalConfig) error {
	//fmt.Printf("------------------------------   Test 3: create origin of 1 SC\n")

	var bindAddress string
	if goshimmerDirectly {
		bindAddress = clu.Config.GoshimmerApiHost()
	} else {
		bindAddress = clu.ApiHosts()[0]
	}

	//fmt.Printf("++++++++++ create origin bind address: %s\n", bindAddress)

	tx, err := cluster.CreateOrigin(bindAddress, sc)
	if err != nil {
		return err
	}

	fmt.Printf("++++++++++ created origin tx:\n%s\n", tx.String())

	ownerAddr := utxodb.GetAddress(sc.OwnerIndexUtxodb)
	sh := tx.MustState().StateHash()
	fmt.Printf("[cluster] new origin tx: id: %s, state hash: %v, addr: %s owner: %s\n",
		tx.ID().String(), sh.String(), sc.Address, ownerAddr.String())

	if tx.ID().String() != sc.Color {
		return fmt.Errorf("mismatch: origin tx id %s should be equal to SC color %s", tx.ID().String(), sc.Color)
	}

	// in real situation we have to wait for confirmation
	err = clu.PostAndWaitForConfirmation(tx.Transaction)
	if err != nil {
		return err
	}
	//fmt.Printf("[cluster] posted node origin tx to Goshimmer: addr: %s, txid: %s\n", sc.Address, tx.ID().String())

	//outs, err := nodeapi.GetAccountOutputs(clu.Config.Goshimmer.BindAddress, &ownerAddr)
	//if err != nil{
	//	fmt.Printf("nodeapi.GetAccountOutputs after post origin %s: %v\n", ownerAddr.String(), err)
	//} else {
	//	fmt.Printf("nodeapi.GetAccountOutputs after post origin %s: \n%+v\n", ownerAddr.String(), outs)
	//}
	return nil
}
