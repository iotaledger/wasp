// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package apilib

import (
	"fmt"
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
	OriginatorSigScheme   signaturescheme.SignatureScheme
	Description           string
	Textout               io.Writer
	Prefix                string
}

// DeployChain performs all actions needed to deploy the chain
// TODO: [KP] Shouldn't that be in the client packages?
// noinspection ALL
func DeployChain(par CreateChainParams) (*coretypes.ChainID, *address.Address, *ledgerstate.Color, error) {
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
	var dkShares *model.DKSharesInfo
	dkShares, err = client.NewWaspClient(par.CommitteeApiHosts[dkgInitiatorIndex]).DKSharesPost(&model.DKSharesPostRequest{
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
	originTx, err := sctransaction.NewChainOriginTransaction(sctransaction.NewChainOriginTransactionParams{
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

	chainColor := ledgerstate.Color(originTx.ID())
	committee := multiclient.New(par.CommitteeApiHosts)
	// ------------ put chain records to hosts
	err = committee.PutChainRecord(&registry.ChainRecord{
		ChainID:        chainID,
		Color:          chainColor,
		CommitteeNodes: par.CommitteePeeringHosts,
	})

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "sending smart contract metadata to Wasp nodes.. FAILED: %v\n", err)
		return nil, nil, nil, err
	}
	fmt.Fprint(textout, "sending smart contract metadata to Wasp nodes.. OK.\n")

	// ------------- activate chain
	err = committee.ActivateChain(chainID)

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
	reqTx, err := sctransaction.NewRootInitRequestTransaction(sctransaction.NewRootInitRequestTransactionParams{
		ChainID:              chainID,
		ChainColor:           chainColor,
		ChainAddress:         chainAddr,
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

	scColor := (ledgerstate.Color)(originTx.ID())
	fmt.Fprint(textout, par.Prefix)

	fmt.Fprintf(textout, "chain has been created succesfully on the Tangle. ChainID: %s, MustAddress: %s, Color: %s, N = %d, T = %d\n",
		chainID.String(), chainAddr.String(), scColor.String(), par.N, par.T)

	return &chainID, &chainAddr, &scColor, err
}
