package apilib

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/sctransaction/origin"
	"github.com/iotaledger/wasp/packages/util/multicall"
	"github.com/iotaledger/wasp/packages/vm"
	"io"
	"io/ioutil"
)

type CreateAndDeploySCParams struct {
	Node                  nodeclient.NodeClient
	CommitteeApiHosts     []string
	CommitteePeeringHosts []string
	AccessNodes           []string
	N                     uint16
	T                     uint16
	OwnerSigScheme        signaturescheme.SignatureScheme
	ProgramHash           hashing.HashValue
	Textout               io.Writer
	Prefix                string
}

// CreateAndDeploySC performs all actions needed to deploy smart contract
func CreateAndDeploySC(par CreateAndDeploySCParams) (*address.Address, *balance.Color, error) {

	textout := ioutil.Discard
	if par.Textout != nil {
		textout = par.Textout
	}
	ownerAddr := par.OwnerSigScheme.Address()

	fmt.Fprint(textout, par.Prefix)
	fmt.Fprintf(textout, "creating and deploying smart contract. Owner address is %s. Parameters N = %d, T = %d\n",
		ownerAddr.String(), par.N, par.T)
	// check if SC is hardcoded. If not, require consistent metadata in all nodes
	fmt.Fprint(textout, par.Prefix)
	fmt.Fprintf(textout, "checking program hash %s.. \n", par.ProgramHash.String())

	fmt.Fprint(textout, par.Prefix)
	if ok := vm.IsBuiltinProgramHash(par.ProgramHash.String()); !ok {
		fmt.Fprint(textout, "user-defined program hash.. \n")
		// it is not a builtin smart contract. Check for metadata
		// must exist and be consistent
		if err := CheckProgramMetadata(par.CommitteeApiHosts, &par.ProgramHash); err != nil {
			fmt.Fprintf(textout, "checking program hash: FAILED: %v\n", err)
			return nil, nil, err
		} else {
			fmt.Fprint(textout, "checking program hash: DONE\n ")
		}
	} else {
		fmt.Fprint(textout, "builtin program hash.. DONE\n")
	}
	// generate distributed key set on committee nodes
	fmt.Fprint(textout, par.Prefix)
	fmt.Fprint(textout, "generating distributed key set..\n")
	scAddr, err := GenerateNewDistributedKeySet(par.CommitteeApiHosts, par.N, par.T)

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "generating distributed key set.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		fmt.Fprint(textout, "generating distributed key set.. DONE\n ")
	}

	fmt.Fprint(textout, par.Prefix)
	fmt.Fprint(textout, "requesting owner address' UTXOs from node..\n")
	allOuts, err := par.Node.GetAccountOutputs(&ownerAddr)

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "requesting owner address' UTXOs from node.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		fmt.Fprint(textout, "requesting owner address' UTXOs from node.. DONE\n ")
	}

	// create origin transaction
	fmt.Fprint(textout, par.Prefix)
	fmt.Fprint(textout, "creating origin transaction..\n")
	originTx, err := origin.NewOriginTransaction(origin.NewOriginTransactionParams{
		Address:              *scAddr,
		OwnerSignatureScheme: par.OwnerSigScheme,
		AllInputs:            allOuts,
		ProgramHash:          par.ProgramHash,
		InputColor:           balance.ColorIOTA,
	})

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "creating origin transaction.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		fmt.Fprintf(textout, "creating origin transaction.. DONE. Origin txid = %s\n", originTx.ID().String())
	}

	fmt.Fprint(textout, par.Prefix)
	fmt.Fprint(textout, "posting and confirming origin transaction to the Tangle..")
	err = par.Node.PostAndWaitForConfirmation(originTx.Transaction)
	if err != nil {
		return nil, nil, err
	}
	succ, errs := PutSCDataMulti(par.CommitteeApiHosts, registry.BootupData{
		Address:        *scAddr,
		OwnerAddress:   ownerAddr,
		Color:          (balance.Color)(originTx.ID()),
		CommitteeNodes: par.CommitteePeeringHosts,
		AccessNodes:    par.AccessNodes,
	})

	fmt.Fprint(textout, par.Prefix)
	if !succ {
		err = multicall.WrapErrors(errs)
		fmt.Fprintf(textout, "posting and confirming origin transaction to the Tangle.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		fmt.Fprint(textout, "posting and confirming origin transaction to the Tangle.. DONE.\n")
	}
	// TODO not finished with access nodes

	fmt.Fprint(textout, par.Prefix)
	fmt.Fprint(textout, "activating smart contract on Wasp nodes..")
	err = ActivateSCMulti(par.CommitteeApiHosts, scAddr.String())

	fmt.Fprint(textout, par.Prefix)
	if !succ {
		err = multicall.WrapErrors(errs)
		fmt.Fprintf(textout, "activating smart contract on Wasp nodes.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		fmt.Fprint(textout, "activating smart contract on Wasp nodes.. DONE.\n")
	}

	color := (balance.Color)(originTx.ID())
	fmt.Fprint(textout, par.Prefix)
	fmt.Fprintf(textout, "smart contract has been created and deployed succesfully\nAddress: %s\n Color: %s\n",
		scAddr.String(), color.String())
	fmt.Fprintf(textout, "Number of committee nodes: %d\n Quorum factor: %d\n", par.N, par.T)

	return scAddr, &color, nil
}
