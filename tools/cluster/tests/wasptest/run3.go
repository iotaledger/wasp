package wasptest

import (
	"fmt"
	"github.com/iotaledger/wasp/tools/cluster"
)

const goshimmer = "127.0.0.1:8080"

func CreateOrigin3SC(clu *cluster.Cluster) error {
	fmt.Printf("------------------------------   Test 3: create origin of 3 SC\n")

	for _, sc := range clu.SmartContractConfig {
		tx1, _, err := cluster.CreateOriginData(goshimmer, &sc)
		if err != nil {
			return err
		}
		tx2, _, err := cluster.CreateOriginDataUtxodb(&sc)
		if err != nil {
			return err
		}
		if tx1.ID().String() != sc.Color {
			return fmt.Errorf("wrong color 1")
		}
		if tx2.ID().String() != sc.Color {
			return fmt.Errorf("wrong color 2")
		}
	}
	fmt.Printf("[cluster] all colors OK\n")

	// TODO not finished
	return nil
}
