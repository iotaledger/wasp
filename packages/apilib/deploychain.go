package apilib

import (
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/multiclient"
	"github.com/iotaledger/wasp/packages/coretypes"
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
func CreateChain(par CreateChainParams) (*coretypes.ChainID, *address.Address, *balance.Color, error) {
	textout := ioutil.Discard
	if par.Textout != nil {
		textout = par.Textout
	}
	ownerAddr := par.OwnerSigScheme.Address()

	fmt.Fprint(textout, par.Prefix)
	fmt.Fprintf(textout, "creating new chain. Owner address is %s. Parameters N = %d, T = %d\n",
		ownerAddr.String(), par.N, par.T)
	// check if SC is hardcoded. If not, require consistent metadata in all nodes
	fmt.Fprint(textout, par.Prefix)

	// generate distributed key set on committee nodes
	chainAddr, err := GenerateNewDistributedKeySet(par.CommitteeApiHosts, par.N, par.T)

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "generating distributed key set.. FAILED: %v\n", err)
		return nil, nil, nil, err
	} else {
		fmt.Fprintf(textout, "generating distributed key set.. OK. Generated address = %s\n", chainAddr.String())
	}

	allOuts, err := par.Node.GetConfirmedAccountOutputs(&ownerAddr)

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "requesting owner address' UTXOs from node.. FAILED: %v\n", err)
		return nil, nil, nil, err
	} else {
		fmt.Fprint(textout, "requesting owner address' UTXOs from node.. OK\n")
	}

	// create origin transaction
	originTx, err := origin.NewOriginTransaction(origin.NewOriginTransactionParams{
		OriginAddress:        *chainAddr,
		OwnerSignatureScheme: par.OwnerSigScheme,
		AllInputs:            allOuts,
	})

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "creating origin transaction.. FAILED: %v\n", err)
		return nil, nil, nil, err
	} else {
		fmt.Fprintf(textout, "creating origin transaction.. OK. Origin txid = %s\n", originTx.ID().String())
	}

	// post origin transaction and wait for confirmation
	err = par.Node.PostAndWaitForConfirmation(originTx.Transaction)
	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "posting origin transaction.. FAILED: %v\n", err)
		return nil, nil, nil, err
	} else {
		fmt.Fprintf(textout, "posting origin transaction.. OK. Origin txid = %s\n", originTx.ID().String())
	}

	chainid := (coretypes.ChainID)(*chainAddr) // using address as chain id
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
		return nil, nil, nil, err
	}
	fmt.Fprint(textout, "sending smart contract metadata to Wasp nodes.. OK.\n")

	allOuts, err = par.Node.GetConfirmedAccountOutputs(&ownerAddr)

	reqTx, err := origin.NewBootupRequestTransaction(origin.NewBootupRequestTransactionParams{
		ChainID:              chainid,
		OwnerSignatureScheme: par.OwnerSigScheme,
		AllInputs:            allOuts,
		CoreContractBinary:   []byte(builtin.DummyBuiltinProgramHash),
		VMType:               "builtin",
	})
	if err != nil {
		fmt.Fprintf(textout, "creating bootup request.. FAILED: %v\n", err)
		return nil, nil, nil, err
	}
	fmt.Fprintf(textout, "creating bootup request.. OK: %v\n", err)

	// post bootup request transaction and wait for confirmation
	err = par.Node.PostAndWaitForConfirmation(reqTx.Transaction)
	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "posting bootup request transaction.. FAILED: %v\n", err)
		return nil, nil, nil, err
	} else {
		fmt.Fprintf(textout, "posting bootup request transaction.. OK. Origin txid = %s\n", reqTx.ID().String())
	}

	scColor := (balance.Color)(originTx.ID())
	fmt.Fprint(textout, par.Prefix)

	fmt.Fprintf(textout, "chain has been created succesfully on the Tangle. ChainID: %s, Address: %s, Color: %s, N = %d, T = %d\n",
		chainid.String(), chainAddr.String(), scColor.String(), par.N, par.T)

	return &chainid, chainAddr, &scColor, err
}
