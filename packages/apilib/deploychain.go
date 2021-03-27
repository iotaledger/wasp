// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package apilib

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"io"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/level1"
	"github.com/iotaledger/wasp/client/multiclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/webapi/model"
)

type CreateChainParams struct {
	Node                  level1.Level1Client
	CommitteeApiHosts     []string
	CommitteePeeringHosts []string
	N                     uint16
	T                     uint16
	OriginatorKeyPair     *ed25519.KeyPair
	Description           string
	Textout               io.Writer
	Prefix                string
}

// DeployChain performs all actions needed to deploy the chain
// TODO: [KP] Shouldn't that be in the client packages?
// noinspection ALL
func DeployChain(par CreateChainParams) (*coretypes.ChainID, ledgerstate.Address, error) {
	var err error
	textout := ioutil.Discard
	if par.Textout != nil {
		textout = par.Textout
	}
	originatorAddr := ledgerstate.NewED25519Address(par.OriginatorKeyPair.PublicKey)

	fmt.Fprint(textout, par.Prefix)
	fmt.Fprintf(textout, "creating new chain. Owner address is %s. Parameters N = %d, T = %d\n",
		originatorAddr.String(), par.N, par.T)
	// check if SC is hardcoded. If not, require consistent metadata in all nodes
	fmt.Fprint(textout, par.Prefix)

	// ----------- run DKG on committee nodes
	var dkgInitiatorIndex = rand.Intn(len(par.CommitteeApiHosts))
	var dkShares *model.DKSharesInfo
	dkShares, err = client.NewWaspClient(par.CommitteeApiHosts[dkgInitiatorIndex]).DKSharesPost(&model.DKSharesPostRequest{
		PeerNetIDs:  par.CommitteePeeringHosts,
		PeerPubKeys: nil,
		Threshold:   par.T,
		TimeoutMS:   60 * 1000, // 1 min
	})
	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "generating distributed key set.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		fmt.Fprintf(textout, "generating distributed key set.. OK. Generated address = %s\n", dkShares.Address)
	}
	var stateControllerAddr ledgerstate.Address
	if stateControllerAddr, err = ledgerstate.AddressFromBase58EncodedString(dkShares.Address); err != nil {
		return nil, nil, err
	}
	committee := multiclient.New(par.CommitteeApiHosts)

	// ------------ put committee records to hosts
	err = committee.PutCommitteeRecord(&registry.CommitteeRecord{
		Address:        stateControllerAddr,
		CommitteeNodes: par.CommitteeApiHosts,
	})
	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "sending committee record to nodes.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		fmt.Fprint(textout, "sending committee record to nodes.. OK\n")
	}

	// ----------- request owner address' outputs from the ledger
	allOuts, err := par.Node.GetConfirmedAccountOutputs(originatorAddr)

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "requesting UTXOs from node.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		fmt.Fprint(textout, "requesting owner address' UTXOs from node.. OK\n")
	}

	// ----------- create origin transaction
	originTx, chainID, err := sctransaction.NewChainOriginTransaction(
		par.OriginatorKeyPair,
		stateControllerAddr,
		nil,
		allOuts...,
	)

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "creating origin transaction.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		fmt.Fprintf(textout, "creating origin transaction.. OK. Origin txid = %s\n", originTx.ID().String())
	}

	// ------------- post origin transaction and wait for confirmation
	err = par.Node.PostAndWaitForConfirmation(originTx)
	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "posting origin transaction.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		fmt.Fprintf(textout, "posting origin transaction.. OK. Origin txid = %s\n", originTx.ID().String())
	}

	// ------------ put chain records to hosts
	err = committee.PutChainRecord(&registry.ChainRecord{
		ChainID:         chainID,
		StateAddressTmp: stateControllerAddr,
	})

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "sending chain data to Wasp nodes.. FAILED: %v\n", err)
		return nil, nil, err
	}
	fmt.Fprint(textout, "sending smart contract metadata to Wasp nodes.. OK.\n")

	// ------------- activate chain
	err = committee.ActivateChain(chainID)

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "activating chain.. FAILED: %v\n", err)
		return nil, nil, err
	}
	fmt.Fprint(textout, "activating chain.. OK.\n")

	// ============= create root init request for the root contract

	// ------------- get UTXOs of the owner
	allOuts, err = par.Node.GetConfirmedAccountOutputs(originatorAddr)
	if err != nil {
		fmt.Fprintf(textout, "GetConfirmedAccountOutputs.. FAILED: %v\n", err)
		return nil, nil, err
	}

	// NOTE: whoever send first init request, is an owner of the chain
	// create root init transaction
	reqTx, err := sctransaction.NewRootInitRequestTransaction(
		par.OriginatorKeyPair,
		chainID,
		par.Description,
		allOuts...,
	)
	if err != nil {
		fmt.Fprintf(textout, "creating root init request.. FAILED: %v\n", err)
		return nil, nil, err
	}
	fmt.Fprintf(textout, "creating root init request.. OK: %v\n", err)

	// ---------- post root init request transaction and wait for confirmation
	err = par.Node.PostAndWaitForConfirmation(reqTx)
	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "posting root init request transaction.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		fmt.Fprintf(textout, "posting root init request transaction.. OK. Origin txid = %s\n", reqTx.ID().String())
	}

	// ---------- wait until the request is processed in all committee nodes
	if err = committee.WaitUntilAllRequestsProcessed(reqTx, 30*time.Second); err != nil {
		fmt.Fprintf(textout, "waiting root init request transaction.. FAILED: %v\n", err)
		return nil, nil, err
	}

	scColor := (ledgerstate.Color)(originTx.ID())
	fmt.Fprint(textout, par.Prefix)

	fmt.Fprintf(textout, "chain has been created succesfully on the Tangle. ChainID: %s, MustAddress: %s, Color: %s, N = %d, T = %d\n",
		chainID.String(), stateControllerAddr.String(), scColor.String(), par.N, par.T)

	return &chainID, stateControllerAddr, err
}
