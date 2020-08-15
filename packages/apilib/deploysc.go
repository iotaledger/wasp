package apilib

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/sctransaction/origin"
	"github.com/iotaledger/wasp/packages/util/multicall"
	"github.com/iotaledger/wasp/packages/vm"
)

type CreateAndDeploySCParams struct {
	Node           nodeclient.NodeClient
	CommitteeNodes []string
	AccessNodes    []string
	N              uint16
	T              uint16
	OwnerSigScheme signaturescheme.SignatureScheme
	ProgramHash    hashing.HashValue
}

// CreateAndDeploySC performs all actions needed to deploy smart contract
func CreateAndDeploySC(par CreateAndDeploySCParams) (*address.Address, *balance.Color, error) {
	// check if SC is hardcoded. If not, require consistent metadata in all nodes
	if ok := vm.IsBuiltinProgramHash(par.ProgramHash.String()); !ok {
		// it is not a builtin smart contract. Check for metadata
		// must exist and be consistent
		if err := CheckProgramMetadata(par.CommitteeNodes, &par.ProgramHash); err != nil {
			return nil, nil, err
		}
	}
	// generate distributed key set on committee nodes
	scAddr, err := GenerateNewDistributedKeySet(par.CommitteeNodes, par.N, par.T)
	if err != nil {
		return nil, nil, err
	}
	ownerAddr := par.OwnerSigScheme.Address()
	allOuts, err := par.Node.GetAccountOutputs(&ownerAddr)
	if err != nil {
		return nil, nil, err
	}
	// create origin transaction
	originTx, err := origin.NewOriginTransaction(origin.NewOriginTransactionParams{
		Address:              *scAddr,
		OwnerSignatureScheme: par.OwnerSigScheme,
		AllInputs:            allOuts,
		ProgramHash:          par.ProgramHash,
		InputColor:           balance.ColorIOTA,
	})
	if err != nil {
		return nil, nil, err
	}
	err = par.Node.PostAndWaitForConfirmation(originTx.Transaction)
	if err != nil {
		return nil, nil, err
	}
	succ, errs := PutSCDataMulti(par.CommitteeNodes, registry.BootupData{
		Address:        *scAddr,
		OwnerAddress:   ownerAddr,
		Color:          (balance.Color)(originTx.ID()),
		CommitteeNodes: par.CommitteeNodes,
		AccessNodes:    par.AccessNodes,
	})
	if !succ {
		return nil, nil, multicall.WrapErrors(errs)
	}
	// TODO not finished with access nodes

	err = ActivateSCMulti(par.CommitteeNodes, scAddr.String())
	if err != nil {
		return nil, nil, err
	}
	color := (balance.Color)(originTx.ID())
	return scAddr, &color, nil
}
