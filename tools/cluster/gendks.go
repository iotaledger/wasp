package cluster

import (
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/packages/waspconn/utxodb"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/origin"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/variables"
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
	origTx, _, err := CreateOriginDataUtxodb(scdata)
	if err != nil {
		return err
	}
	scdata.Color = origTx.ID().String()
	return nil
}

func CreateOriginDataUtxodb(scdata *SmartContractFinalConfig) (*sctransaction.Transaction, state.Batch, error) {
	vars := variables.New(nil)
	vars.Set("description", scdata.Description)

	addr, err := address.FromBase58(scdata.Address)
	if err != nil {
		return nil, nil, err
	}
	progHash, err := hashing.HashValueFromString(scdata.ProgramHash)
	if err != nil {
		return nil, nil, err
	}
	// creating origin transaction just to determine color
	origTx, batch, err := waspapi.CreateOriginDataUtxodb(origin.NewOriginParams{
		Address:              addr,
		OwnerSignatureScheme: utxodb.GetSigScheme(utxodb.GetAddress(scdata.OwnerIndexUtxodb)),
		ProgramHash:          progHash,
		Variables:            vars,
	})
	if err != nil {
		return nil, nil, err
	}
	return origTx, batch, nil
}

func CreateOriginData(host string, scdata *SmartContractFinalConfig) (*sctransaction.Transaction, state.Batch, error) {
	vars := variables.New(nil)
	vars.Set("description", scdata.Description)

	addr, err := address.FromBase58(scdata.Address)
	if err != nil {
		return nil, nil, err
	}
	progHash, err := hashing.HashValueFromString(scdata.ProgramHash)
	if err != nil {
		return nil, nil, err
	}
	// creating origin transaction just to determine color
	origTx, batch, err := waspapi.CreateOriginData(host, origin.NewOriginParams{
		Address:              addr,
		OwnerSignatureScheme: utxodb.GetSigScheme(utxodb.GetAddress(scdata.OwnerIndexUtxodb)),
		ProgramHash:          progHash,
		Variables:            vars,
	})
	if err != nil {
		return nil, nil, err
	}
	return origTx, batch, nil
}
