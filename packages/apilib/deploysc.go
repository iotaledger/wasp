package apilib

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"io"
	"io/ioutil"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/multiclient"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/sctransaction/origin"
	"github.com/iotaledger/wasp/packages/subscribe"
	"github.com/iotaledger/wasp/packages/util/multicall"
)

type CreateChainParams struct {
	Node                  nodeclient.NodeClient
	CommitteeApiHosts     []string
	CommitteePeeringHosts []string
	AccessNodes           []string
	N                     uint16
	T                     uint16
	OwnerSigScheme        signaturescheme.SignatureScheme
	ProgramHash           hashing.HashValue
	Description           string
	Textout               io.Writer
	Prefix                string
}

type ActivateChainParams struct {
	ChainID           coretypes.ChainID
	ApiHosts          []string
	WaitForCompletion bool
	PublisherHosts    []string
	Timeout           time.Duration
}

func ActivateChain(par ActivateChainParams) error {
	funs := make([]func() error, 0)
	for _, host := range par.ApiHosts {
		h := host
		funs = append(funs, func() error {
			return client.NewWaspClient(h).ActivateChain(&par.ChainID)
		})
	}
	if !par.WaitForCompletion {
		_, errs := multicall.MultiCall(funs, 1*time.Second)
		return multicall.WrapErrors(errs)
	}
	subs, err := subscribe.SubscribeMulti(par.PublisherHosts, "state")
	if err != nil {
		return err
	}
	defer subs.Close()
	_, errs := multicall.MultiCall(funs, 1*time.Second)
	err = multicall.WrapErrors(errs)
	if err != nil {
		return err
	}
	// Chain is initialized when it reaches state index #1
	patterns := [][]string{{"state", par.ChainID.String(), "1"}}
	succ := subs.WaitForPatterns(patterns, par.Timeout)
	if !succ {
		return fmt.Errorf("didn't receive activation message in %v", par.Timeout)
	}
	return nil
}

func DeactivateChain(par ActivateChainParams) error {
	funs := make([]func() error, 0)
	for _, host := range par.ApiHosts {
		h := host
		funs = append(funs, func() error {
			return client.NewWaspClient(h).DeactivateChain(&par.ChainID)
		})
	}
	if !par.WaitForCompletion {
		_, errs := multicall.MultiCall(funs, 1*time.Second)
		return multicall.WrapErrors(errs)
	}
	subs, err := subscribe.SubscribeMulti(par.PublisherHosts, "dismissed_committee")
	if err != nil {
		return err
	}
	defer subs.Close()
	_, errs := multicall.MultiCall(funs, 1*time.Second)
	err = multicall.WrapErrors(errs)
	if err != nil {
		return err
	}
	patterns := [][]string{{"dismissed_committee", par.ChainID.String(), "1"}}
	succ := subs.WaitForPatterns(patterns, par.Timeout)
	if !succ {
		return fmt.Errorf("didn't receive deactivation message in %v", par.Timeout)
	}
	return nil
}

// CreateChain performs all actions needed to deploy smart contract, except activation
// noinspection ALL
func CreateChain(par CreateChainParams) (*address.Address, *balance.Color, error) {
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
	md, err := multiclient.New(par.CommitteeApiHosts).CheckProgramMetadata(&par.ProgramHash)
	if err != nil {
		fmt.Fprintf(textout, "checking program metadata: FAILED: %v\n", err)
		return nil, nil, err
	}
	fmt.Fprintf(textout, "checking program metadata: OK. VMType: '%s', description: '%s'\n",
		md.VMType, md.Description)

	// generate distributed key set on committee nodes
	scAddr, err := GenerateNewDistributedKeySet(par.CommitteeApiHosts, par.N, par.T)

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "generating distributed key set.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		fmt.Fprintf(textout, "generating distributed key set.. OK. Generated address = %s\n", scAddr.String())
	}

	allOuts, err := par.Node.GetConfirmedAccountOutputs(&ownerAddr)

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "requesting owner address' UTXOs from node.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		fmt.Fprint(textout, "requesting owner address' UTXOs from node.. OK\n")
	}

	// create origin transaction
	originTx, err := origin.NewOriginTransaction(origin.NewOriginTransactionParams{
		OriginAddress:        *scAddr,
		OwnerSignatureScheme: par.OwnerSigScheme,
		AllInputs:            allOuts,
		ProgramHash:          par.ProgramHash,
		Description:          par.Description,
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

	chainid := (coretypes.ChainID)(*scAddr)
	err = multiclient.New(par.CommitteeApiHosts).PutBootupData(&registry.BootupData{
		ChainID:        chainid,
		OwnerAddress:   ownerAddr,
		Color:          (balance.Color)(originTx.ID()),
		CommitteeNodes: par.CommitteePeeringHosts,
		AccessNodes:    par.AccessNodes,
	})

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "sending smart contract metadata to Wasp nodes.. FAILED: %v\n", err)
		return nil, nil, err
	}
	fmt.Fprint(textout, "sending smart contract metadata to Wasp nodes.. OK.\n")
	// TODO not finished with access nodes

	scColor := (balance.Color)(originTx.ID())
	fmt.Fprint(textout, par.Prefix)
	fmt.Fprintf(textout, "smart contract has been created succesfully. Target: %s, Color: %s, N = %d, T = %d\n",
		scAddr.String(), scColor.String(), par.N, par.T)
	return scAddr, &scColor, err
}
