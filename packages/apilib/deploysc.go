package apilib

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/sctransaction/origin"
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

func CreateAndDeploySC(params CreateAndDeploySCParams) error {
	// generate distributed key set on committee nodes
	scAddr, err := GenerateNewDistributedKeySet(params.CommitteeNodes, params.N, params.T)
	if err != nil {
		return err
	}
	ownerAddr := params.OwnerSigScheme.Address()
	allOuts, err := params.Node.GetAccountOutputs(&ownerAddr)
	if err != nil {
		return err
	}
	// create origin transaction
	originTx, err := origin.NewOriginTransaction(origin.NewOriginTransactionParams{
		Address:              *scAddr,
		OwnerSignatureScheme: params.OwnerSigScheme,
		AllInputs:            allOuts,
		ProgramHash:          params.ProgramHash,
		InputColor:           balance.ColorIOTA,
	})
	if err != nil {
		return err
	}
	err = params.Node.PostAndWaitForConfirmation(originTx.Transaction)
	if err != nil {
		return err
	}
	bd := registry.BootupData{
		Address:        *scAddr,
		OwnerAddress:   ownerAddr,
		Color:          (balance.Color)(originTx.ID()),
		CommitteeNodes: params.CommitteeNodes,
		AccessNodes:    params.AccessNodes,
	}
	for _, host := range params.CommitteeNodes {
		err = PutSCData(host, bd)
		if err != nil {
			return err
		}
	}
	// TODO not finished with access nodes
	for _, host := range params.AccessNodes {
		err = PutSCData(host, bd)
		if err != nil {
			return err
		}
	}

	for _, host := range params.CommitteeNodes {
		err = ActivateSC(host, scAddr.String())
		if err != nil {
			return err
		}
	}
	// TODO not finished with access nodes
	for _, host := range params.AccessNodes {
		err = ActivateSC(host, scAddr.String())
		if err != nil {
			return err
		}
	}
	return nil
}
