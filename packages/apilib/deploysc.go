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
	"github.com/iotaledger/wasp/packages/subscribe"
	"github.com/iotaledger/wasp/packages/util/multicall"
	"github.com/iotaledger/wasp/packages/vm/builtins"
	"io"
	"io/ioutil"
	"sync"
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
	// wait for init
	WaitForInitialization   bool
	CommitteePublisherHosts []string
	Timeout                 time.Duration
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

	//fmt.Fprintf(textout, par.Prefix+"waiting for %v for nodes to connect..\n", committee.InitConnectPeriod)
	//time.Sleep(committee.InitConnectPeriod)

	scColor := (balance.Color)(originTx.ID())
	fmt.Fprint(textout, par.Prefix)
	fmt.Fprintf(textout, "smart contract has been created succesfully. Address: %s, Color: %s, N = %d, T = %d\n",
		scAddr.String(), scColor.String(), par.N, par.T)

	if !par.WaitForInitialization {
		return scAddr, &scColor, nil
	}
	fmt.Fprint(textout, par.Prefix)
	err = WaitSmartContractInitialized(par.CommitteePublisherHosts, scAddr, &scColor, par.Timeout)
	if err != nil {
		fmt.Fprintf(textout, "waiting for smart contract to run its initialization code.. FAIL: %v\n", err)
		return nil, nil, err
	}
	fmt.Fprintf(textout, "waiting for smart contract run its initialization code.. SUCCESS\n")
	return scAddr, &scColor, nil
}

// WaitSmartContractInitialized waits for the wasp hosts to publish init requests were settled successfully
func WaitSmartContractInitialized(hosts []string, scAddr *address.Address, scColor *balance.Color, timeout time.Duration) error {
	var err error
	pattern := []string{"request_out", scAddr.String(), scColor.String(), "0"}
	var wg sync.WaitGroup
	wg.Add(1)
	subscribe.ListenForPatternMulti(hosts, pattern, func(ok bool) {
		if !ok {
			err = fmt.Errorf("smart contract wasn't deployed correctly in %v", timeout)
		}
		wg.Done()
	}, timeout)
	return err
}
