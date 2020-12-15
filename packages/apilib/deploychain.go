// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package apilib

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/multiclient"
	"github.com/iotaledger/wasp/packages/coretypes"
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
	OriginatorSigScheme   signaturescheme.SignatureScheme
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

// DeployChain performs all actions needed to deploy the chain
// TODO: [KP] Shouldn't that be in the client packages?
// noinspection ALL
func DeployChain(par CreateChainParams) (*coretypes.ChainID, *address.Address, *balance.Color, error) {
	var err error
	textout := ioutil.Discard
	if par.Textout != nil {
		textout = par.Textout
	}
	originatorAddr := par.OriginatorSigScheme.Address()

	fmt.Fprint(textout, par.Prefix)
	fmt.Fprintf(textout, "creating new chain. Owner address is %s. Parameters N = %d, T = %d\n",
		originatorAddr.String(), par.N, par.T)
	// check if SC is hardcoded. If not, require consistent metadata in all nodes
	fmt.Fprint(textout, par.Prefix)

	// ----------- run DKG on committee nodes
	var dkgInitiatorIndex = rand.Intn(len(par.CommitteeApiHosts))
	var dkShares *client.DKSharesInfo
	dkShares, err = client.NewWaspClient(par.CommitteeApiHosts[dkgInitiatorIndex]).DKSharesPost(&client.DKSharesPostRequest{
		PeerNetIDs:  par.CommitteePeeringHosts,
		PeerPubKeys: nil,
		Threshold:   par.T,
		TimeoutMS:   60000, // 1 min
	})
	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "generating distributed key set.. FAILED: %v\n", err)
		return nil, nil, nil, err
	} else {
		fmt.Fprintf(textout, "generating distributed key set.. OK. Generated address = %s\n", dkShares.Address)
	}
	var chainAddr address.Address
	if chainAddr, err = address.FromBase58(dkShares.Address); err != nil {
		return nil, nil, nil, err
	}
	var chainID coretypes.ChainID = coretypes.ChainID(chainAddr) // That's temporary, a color should be used later.

	// ----------- request owner address' outputs from the ledger
	allOuts, err := par.Node.GetConfirmedAccountOutputs(&originatorAddr)

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "requesting owner address' UTXOs from node.. FAILED: %v\n", err)
		return nil, nil, nil, err
	} else {
		fmt.Fprint(textout, "requesting owner address' UTXOs from node.. OK\n")
	}

	// ----------- create origin transaction
	originTx, err := origin.NewOriginTransaction(origin.NewOriginTransactionParams{
		OriginAddress:             chainAddr,
		OriginatorSignatureScheme: par.OriginatorSigScheme,
		AllInputs:                 allOuts,
	})

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "creating origin transaction.. FAILED: %v\n", err)
		return nil, nil, nil, err
	} else {
		fmt.Fprintf(textout, "creating origin transaction.. OK. Origin txid = %s\n", originTx.ID().String())
	}

	// ------------- post origin transaction and wait for confirmation
	err = par.Node.PostAndWaitForConfirmation(originTx.Transaction)
	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "posting origin transaction.. FAILED: %v\n", err)
		return nil, nil, nil, err
	} else {
		fmt.Fprintf(textout, "posting origin transaction.. OK. Origin txid = %s\n", originTx.ID().String())
	}

	committee := multiclient.New(par.CommitteeApiHosts)

	// ------------ put chain records to hosts
	err = committee.PutChainRecord(&registry.ChainRecord{
		ChainID:        chainID,
		Color:          (balance.Color)(originTx.ID()),
		CommitteeNodes: par.CommitteePeeringHosts,
	})

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "sending smart contract metadata to Wasp nodes.. FAILED: %v\n", err)
		return nil, nil, nil, err
	}
	fmt.Fprint(textout, "sending smart contract metadata to Wasp nodes.. OK.\n")

	// ------------- activate chain
	err = committee.ActivateChain(&chainID)

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "activating chain.. FAILED: %v\n", err)
		return nil, nil, nil, err
	}
	fmt.Fprint(textout, "activating chain.. OK.\n")

	// ============= create root init request for the root contract

	// TODO to via chainclient with timeout etc

	// ------------- get UTXOs of the owner
	allOuts, err = par.Node.GetConfirmedAccountOutputs(&originatorAddr)
	if err != nil {
		fmt.Fprintf(textout, "GetConfirmedAccountOutputs.. FAILED: %v\n", err)
		return nil, nil, nil, err
	}

	// NOTE: whoever send first init request, is an owner of the chain
	// create root init transaction
	reqTx, err := origin.NewRootInitRequestTransaction(origin.NewRootInitRequestTransactionParams{
		ChainID:              chainID,
		OwnerSignatureScheme: par.OriginatorSigScheme,
		AllInputs:            allOuts,
		Description:          par.Description,
	})
	if err != nil {
		fmt.Fprintf(textout, "creating root init request.. FAILED: %v\n", err)
		return nil, nil, nil, err
	}
	fmt.Fprintf(textout, "creating root init request.. OK: %v\n", err)

	// ---------- post root init request transaction and wait for confirmation
	err = par.Node.PostAndWaitForConfirmation(reqTx.Transaction)
	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "posting root init request transaction.. FAILED: %v\n", err)
		return nil, nil, nil, err
	} else {
		fmt.Fprintf(textout, "posting root init request transaction.. OK. Origin txid = %s\n", reqTx.ID().String())
	}

	// ---------- wait until the request is processed in all committee nodes
	if err = committee.WaitUntilAllRequestsProcessed(reqTx, 30*time.Second); err != nil {
		fmt.Fprintf(textout, "waiting root init request transaction.. FAILED: %v\n", err)
		return nil, nil, nil, err
	}

	scColor := (balance.Color)(originTx.ID())
	fmt.Fprint(textout, par.Prefix)

	fmt.Fprintf(textout, "chain has been created succesfully on the Tangle. ChainID: %s, MustAddress: %s, Color: %s, N = %d, T = %d\n",
		chainID.String(), chainAddr.String(), scColor.String(), par.N, par.T)

	return &chainID, &chainAddr, &scColor, err
}
