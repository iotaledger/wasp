package apilib

import (
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/sctransaction/origin"
	"github.com/iotaledger/wasp/packages/util/multicall"
	"github.com/iotaledger/wasp/packages/vm/builtins"
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
// noinspection ALL
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
	if !builtins.IsBuiltinProgramHash(par.ProgramHash.String()) {
		fmt.Fprint(textout, "program hash is user-defined.. \n")
		// it is not a builtin smart contract. Check for metadata
		// must exist and be consistent
		if err := CheckProgramMetadata(par.CommitteeApiHosts, &par.ProgramHash); err != nil {
			fmt.Fprintf(textout, "checking program metadata: FAILED: %v\n", err)
			return nil, nil, err
		} else {
			fmt.Fprint(textout, "checking program metadata: OK\n ")
		}
	} else {
		fmt.Fprintf(textout, "builtin program %s. OK\n", par.ProgramHash.String())
	}

	// generate distributed key set on committee nodes
	scAddr, err := GenerateNewDistributedKeySet(par.CommitteeApiHosts, par.N, par.T)

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "generating distributed key set.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		fmt.Fprintf(textout, "generating distributed key set.. OK. Generated address = %s\n", scAddr.String())
	}

	fmt.Fprint(textout, par.Prefix)
	allOuts, err := par.Node.GetAccountOutputs(&ownerAddr)

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "requesting owner address' UTXOs from node.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		fmt.Fprint(textout, "requesting owner address' UTXOs from node.. OK\n")
	}

	// create origin transaction
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
		fmt.Fprintf(textout, "creating origin transaction.. OK. Origin txid = %s\n", originTx.ID().String())
	}

	err = par.Node.PostAndWaitForConfirmation(originTx.Transaction)
	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "posting origin transaction.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		fmt.Fprintf(textout, "posting origin transaction.. OK. Origin txid = %s\n", originTx.ID().String())
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
		fmt.Fprintf(textout, "sending smart contract metadata to Wasp nodes.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		fmt.Fprint(textout, "sending smart contract metadata to Wasp nodes.. OK.\n")
	}
	// TODO not finished with access nodes

	err = ActivateSCMulti(par.CommitteeApiHosts, scAddr.String())
	if err != nil {
		fmt.Fprintf(textout, "activate SC.. FAILED: %v\n", err)
		return nil, nil, err
	}

	fmt.Fprint(textout, par.Prefix)
	if !succ {
		err = multicall.WrapErrors(errs)
		fmt.Fprintf(textout, "activating smart contract on Wasp nodes.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		fmt.Fprint(textout, "activating smart contract on Wasp nodes.. OK.\n")
	}

	fmt.Fprintf(textout, par.Prefix+"waiting for %v for nodes to connect..\n", committee.InitConnectPeriod)
	time.Sleep(committee.InitConnectPeriod)

	color := (balance.Color)(originTx.ID())
	fmt.Fprint(textout, par.Prefix)
	fmt.Fprintf(textout, "smart contract has been created and deployed succesfully. Address: %s, Color: %s, N = %d, T = %d\n",
		scAddr.String(), color.String(), par.N, par.T)

	return scAddr, &color, nil
}
