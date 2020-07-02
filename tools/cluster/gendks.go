package cluster

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/table"

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
		committee := cluster.WaspHosts(sc.CommitteeNodes, (*WaspNodeConfig).ApiHost)
		addr, err := waspapi.GenerateNewDistributedKeySet(
			committee,
			uint16(len(committee)),
			uint16(sc.Quorum),
		)
		if err != nil {
			return err
		}

		fmt.Printf("[cluster] Generated key set for SC with address %s\n", addr)

		dkShares := make([]string, 0)
		for _, host := range cluster.ApiHosts() {
			dks, err := waspapi.ExportDKShare(host, addr)
			if err != nil {
				return err
			}
			dkShares = append(dkShares, dks)
		}

		scdata := SmartContractFinalConfig{
			Address:          addr.String(),
			Description:      sc.Description,
			ProgramHash:      hashing.HashStrings(sc.Description).String(),
			CommitteeNodes:   sc.CommitteeNodes,
			OwnerIndexUtxodb: i + 1, // owner index from utxodb
			DKShares:         dkShares,
		}
		if err = calcColorUtxodb(&scdata); err != nil {
			return err
		}
		keys = append(keys, scdata)
	}
	buf, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(keysFile, buf, 0644)
}

func calcColorUtxodb(scdata *SmartContractFinalConfig) error {
	origTx, err := CreateOriginUtxodb(scdata)
	if err != nil {
		return err
	}
	scdata.Color = origTx.ID().String()
	return nil
}

func CreateOriginUtxodb(scdata *SmartContractFinalConfig) (*sctransaction.Transaction, error) {
	addr, err := address.FromBase58(scdata.Address)
	if err != nil {
		return nil, err
	}
	progHash, err := hashing.HashValueFromBase58(scdata.ProgramHash)
	if err != nil {
		return nil, err
	}
	// creating origin transaction just to determine color
	origTx, err := waspapi.CreateOriginUtxodb(waspapi.CreateOriginParams{
		Address:              addr,
		OwnerSignatureScheme: utxodb.GetSigScheme(utxodb.GetAddress(scdata.OwnerIndexUtxodb)),
		ProgramHash:          progHash,
		Variables: table.FromMap(map[table.Key][]byte{
			"description": []byte(scdata.Description),
		}),
	})
	if err != nil {
		return nil, err
	}
	return origTx, nil
}

func CreateOrigin(host string, scdata *SmartContractFinalConfig) (*sctransaction.Transaction, error) {
	addr, err := address.FromBase58(scdata.Address)
	if err != nil {
		return nil, err
	}
	progHash, err := hashing.HashValueFromBase58(scdata.ProgramHash)
	if err != nil {
		return nil, err
	}
	// creating origin transaction just to determine color
	origTx, err := waspapi.CreateOrigin(host, waspapi.CreateOriginParams{
		Address:              addr,
		OwnerSignatureScheme: utxodb.GetSigScheme(utxodb.GetAddress(scdata.OwnerIndexUtxodb)),
		ProgramHash:          progHash,
		Variables: table.FromMap(map[table.Key][]byte{
			"description": []byte(scdata.Description),
		}),
	})
	if err != nil {
		return nil, err
	}
	fmt.Printf("[cluster] created origin data: addr : %s descr: %s program hash: %s\n",
		addr.String(), scdata.Description, scdata.ProgramHash)
	return origTx, nil
}
