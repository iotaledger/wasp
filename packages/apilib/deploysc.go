package apilib

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/sctransaction/origin"
	"github.com/iotaledger/wasp/packages/util/multicall"
	"github.com/iotaledger/wasp/packages/vm"
	"io"
	"io/ioutil"
	"time"
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

	_, _ = fmt.Fprint(textout, par.Prefix)
	_, _ = fmt.Fprintf(textout, "creating and deploying smart contract. Owner address is %s. Parameters N = %d, T = %d\n",
		ownerAddr.String(), par.N, par.T)
	// check if SC is hardcoded. If not, require consistent metadata in all nodes
	_, _ = fmt.Fprint(textout, par.Prefix)
	_, _ = fmt.Fprintf(textout, "checking program hash %s.. \n", par.ProgramHash.String())

	_, _ = fmt.Fprint(textout, par.Prefix)
	if ok := vm.IsBuiltinProgramHash(par.ProgramHash.String()); !ok {
		_, _ = fmt.Fprint(textout, "program hash is user-defined.. \n")
		// it is not a builtin smart contract. Check for metadata
		// must exist and be consistent
		if err := CheckProgramMetadata(par.CommitteeApiHosts, &par.ProgramHash); err != nil {
			_, _ = fmt.Fprintf(textout, "checking program metadata: FAILED: %v\n", err)
			return nil, nil, err
		} else {
			_, _ = fmt.Fprint(textout, "checking program metadata: OK\n ")
		}
	} else {
		_, _ = fmt.Fprintf(textout, "builtin program %s. OK\n", par.ProgramHash.String())
	}
	// generate distributed key set on committee nodes
	scAddr, err := GenerateNewDistributedKeySet(par.CommitteeApiHosts, par.N, par.T)

	_, _ = fmt.Fprint(textout, par.Prefix)
	if err != nil {
		_, _ = fmt.Fprintf(textout, "generating distributed key set.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		_, _ = fmt.Fprintf(textout, "generating distributed key set.. OK. Generated address = %s\n", scAddr.String())
	}

	_, _ = fmt.Fprint(textout, par.Prefix)
	allOuts, err := par.Node.GetAccountOutputs(&ownerAddr)

	_, _ = fmt.Fprint(textout, par.Prefix)
	if err != nil {
		_, _ = fmt.Fprintf(textout, "requesting owner address' UTXOs from node.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		_, _ = fmt.Fprint(textout, "requesting owner address' UTXOs from node.. OK\n")
	}

	// create origin transaction
	originTx, err := origin.NewOriginTransaction(origin.NewOriginTransactionParams{
		Address:              *scAddr,
		OwnerSignatureScheme: par.OwnerSigScheme,
		AllInputs:            allOuts,
		ProgramHash:          par.ProgramHash,
		InputColor:           balance.ColorIOTA,
	})

	_, _ = fmt.Fprint(textout, par.Prefix)
	if err != nil {
		_, _ = fmt.Fprintf(textout, "creating origin transaction.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		_, _ = fmt.Fprintf(textout, "creating origin transaction.. OK. Origin txid = %s\n", originTx.ID().String())
	}

	err = par.Node.PostAndWaitForConfirmation(originTx.Transaction)
	_, _ = fmt.Fprint(textout, par.Prefix)
	if err != nil {
		_, _ = fmt.Fprintf(textout, "posting origin transaction.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		_, _ = fmt.Fprintf(textout, "posting origin transaction.. OK. Origin txid = %s\n", originTx.ID().String())
	}

	succ, errs := PutSCDataMulti(par.CommitteeApiHosts, registry.BootupData{
		Address:        *scAddr,
		OwnerAddress:   ownerAddr,
		Color:          (balance.Color)(originTx.ID()),
		CommitteeNodes: par.CommitteePeeringHosts,
		AccessNodes:    par.AccessNodes,
	})

	_, _ = fmt.Fprint(textout, par.Prefix)
	if !succ {
		err = multicall.WrapErrors(errs)
		_, _ = fmt.Fprintf(textout, "sending smart contract metadata to Wasp nodes.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		_, _ = fmt.Fprint(textout, "sending smart contract metadata to Wasp nodes.. OK.\n")
	}
	// TODO not finished with access nodes

	err = ActivateSCMulti(par.CommitteeApiHosts, scAddr.String())

	_, _ = fmt.Fprint(textout, par.Prefix)
	if !succ {
		err = multicall.WrapErrors(errs)
		_, _ = fmt.Fprintf(textout, "activating smart contract on Wasp nodes.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		_, _ = fmt.Fprint(textout, "activating smart contract on Wasp nodes.. OK.\n")
	}

	_, _ = fmt.Fprintf(textout, par.Prefix+"waiting for %v for nodes to connect..\n", committee.InitConnectPeriod)
	time.Sleep(committee.InitConnectPeriod)

	color := (balance.Color)(originTx.ID())
	_, _ = fmt.Fprint(textout, par.Prefix)
	_, _ = fmt.Fprintf(textout, "smart contract has been created and deployed succesfully. Address: %s, Color: %s, N = %d, T = %d\n",
		scAddr.String(), color.String(), par.N, par.T)

	return scAddr, &color, nil
}
