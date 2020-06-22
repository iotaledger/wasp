package wasptest

import (
	"fmt"
	nodeapi "github.com/iotaledger/goshimmer/packages/waspconn/apilib"
	"github.com/iotaledger/wasp/tools/cluster"
)

func CreateOrigin1SC(clu *cluster.Cluster) error {
	//fmt.Printf("------------------------------   Test 3: create origin of 1 SC\n")

	sc := clu.SmartContractConfig[0]
	tx, err := cluster.CreateOrigin(clu.Config.Goshimmer.BindAddress, &sc)
	if err != nil {
		return err
	}

	fmt.Printf("++++++++++ created origin tx:\n%s\n", tx.String())
	//fmt.Printf("++++++++++ created origin batch:\n%s\n", batch.String())

	sh := tx.MustState().VariableStateHash()
	fmt.Printf("[cluster] new origin tx: id: %s, state hash: %v, addr: %s\n",
		tx.ID().String(), sh.String(), sc.Address)

	err = nodeapi.PostTransaction(clu.Config.Goshimmer.BindAddress, tx.Transaction)
	if err != nil {
		return err
	}
	//fmt.Printf("[cluster] posted node origin tx to Goshimmer: addr: %s, txid: %s\n", sc.Address, tx.ID().String())

	return nil
}
