package cluster

import (
	"encoding/json"
	"fmt"
	"github.com/iotaledger/wasp/packages/hashing"
	"io/ioutil"

	waspapi "github.com/iotaledger/wasp/packages/apilib"
)

func (cluster *Cluster) GenerateDKSets() error {
	keysFile := cluster.ConfigKeysPath()
	exists, err := fileExists(keysFile)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("dk sets already generated in keys.json")
	}

	keys := make([]SmartContractFinalConfig, 0)

	for i, sc := range cluster.Config.SmartContracts {
		committee, err := cluster.Committee(&sc)
		if err != nil {
			return err
		}
		addr, err := waspapi.GenerateNewDistributedKeySet(
			committee,
			uint16(len(committee)),
			uint16(sc.Quorum),
		)
		if err != nil {
			return err
		}

		fmt.Printf("Generated key set for SC with address %s\n", addr)

		dkShares := make([]string, 0)
		for _, host := range cluster.Hosts() {
			dks, err := waspapi.ExportDKShare(host, addr)
			if err != nil {
				return err
			}
			dkShares = append(dkShares, dks)
		}

		keys = append(keys, SmartContractFinalConfig{
			Address:          addr.String(),
			Description:      sc.Description,
			ProgramHash:      hashing.HashStrings(sc.Description).String(),
			Nodes:            sc.Nodes,
			OwnerIndexUtxodb: i + 1, // owner index from utxodb
			DKShares:         dkShares,
		})
	}
	buf, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(keysFile, buf, 0644)
}
