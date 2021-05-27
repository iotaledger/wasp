// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package apilib

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/iotaledger/wasp/packages/registry_pkg/chain_record"
	"github.com/iotaledger/wasp/packages/registry_pkg/committee_record"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/goshimmer"
	"github.com/iotaledger/wasp/client/multiclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/webapi/model"
)

type CreateChainParams struct {
	Node                  *goshimmer.Client
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
// TODO: [ED] decouple DKG from chain deployment into separate call
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
		originatorAddr.Base58(), par.N, par.T)
	// check if SC is hardcoded. If not, require consistent metadata in all nodes
	fmt.Fprint(textout, par.Prefix)

	// ----------- run DKG on committee nodes
	dkgInitiatorIndex := rand.Intn(len(par.CommitteeApiHosts))
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
		return nil, nil, xerrors.Errorf("DKSharesPost: %w", err)
	} else {
		fmt.Fprintf(textout, "generating distributed key set.. OK. Generated address = %s\n", dkShares.Address)
	}
	var stateControllerAddr ledgerstate.Address
	if stateControllerAddr, err = ledgerstate.AddressFromBase58EncodedString(dkShares.Address); err != nil {
		return nil, nil, xerrors.Errorf("AddressFromBase58EncodedString: %w", err)
	}
	committee := multiclient.New(par.CommitteeApiHosts)

	// ------------ put committee records to hosts
	err = committee.PutCommitteeRecord(&committee_record.CommitteeRecord{
		Address: stateControllerAddr,
		Nodes:   par.CommitteePeeringHosts,
	})
	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "sending committee record to nodes.. FAILED: %v\n", err)
		return nil, nil, xerrors.Errorf("PutCommitteeRecord: %w", err)
	} else {
		fmt.Fprint(textout, "sending committee record to nodes.. OK\n")
	}

	// ----------- request owner address' outputs from the ledger
	allOuts, err := par.Node.GetConfirmedOutputs(originatorAddr)

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "requesting UTXOs from node.. FAILED: %v\n", err)
		return nil, nil, xerrors.Errorf("GetConfirmedOutputs: %w", err)
	} else {
		fmt.Fprint(textout, "requesting owner address' UTXOs from node.. OK\n")
	}

	// ----------- create origin transaction
	originTx, chainID, err := transaction.NewChainOriginTransaction(
		par.OriginatorKeyPair,
		stateControllerAddr,
		nil,
		time.Now(),
		allOuts...,
	)

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "creating origin transaction.. FAILED: %v\n", err)
		return nil, nil, xerrors.Errorf("NewChainOriginTransaction: %w", err)
	} else {
		fmt.Fprintf(textout, "creating origin transaction.. OK. Origin txid = %s\n", originTx.ID().String())
	}

	// ------------- post origin transaction and wait for confirmation
	err = par.Node.PostAndWaitForConfirmation(originTx)
	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "posting origin transaction.. FAILED: %v\n", err)
		return nil, nil, xerrors.Errorf("posting origin transaction: %w", err)
	} else {
		fmt.Fprintf(textout, "posting origin transaction.. OK. txid: %s\n", originTx.ID().Base58())
	}

	// ------------ put chain records to hosts
	err = committee.PutChainRecord(&chain_record.ChainRecord{
		ChainID: &chainID,
	})

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "sending chain data to Wasp nodes.. FAILED: %v\n", err)
		return nil, nil, xerrors.Errorf("PutChainRecord: %w", err)
	}
	fmt.Fprint(textout, "sending chain data to Wasp nodes.. OK.\n")

	// ------------- activate chain
	err = committee.ActivateChain(chainID)

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "activating chain.. FAILED: %v\n", err)
		return nil, nil, xerrors.Errorf("ActivateChain: %w", err)
	}
	fmt.Fprint(textout, "activating chain.. OK.\n")

	// ============= create root init request for the root contract

	// ------------- get UTXOs of the owner
	allOuts, err = par.Node.GetConfirmedOutputs(originatorAddr)
	if err != nil {
		fmt.Fprintf(textout, "GetConfirmedOutputs.. FAILED: %v\n", err)
		return nil, nil, xerrors.Errorf("GetConfirmedOutputs: %w", err)
	}

	// NOTE: whoever send first init request, is an owner of the chain
	// create root init transaction
	reqTx, err := transaction.NewRootInitRequestTransaction(
		par.OriginatorKeyPair,
		chainID,
		par.Description,
		time.Now(),
		allOuts...,
	)
	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "creating root init request.. FAILED: %v\n", err)
		return nil, nil, xerrors.Errorf("NewRootInitRequestTransaction: %w", err)
	}
	fmt.Fprintf(textout, "creating root init request.. OK\n")
	fmt.Fprintf(textout, "root init txid: %s, reqidBase58: %s\n", reqTx.ID().Base58(), reqTx.Essence().Outputs()[0].ID().Base58())

	// ---------- post root init request transaction and wait for confirmation
	err = par.Node.PostAndWaitForConfirmation(reqTx)
	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "posting root init request transaction.. FAILED: %v\n", err)
		return nil, nil, xerrors.Errorf("posting root init request: %w", err)
	} else {
		fmt.Fprintf(textout, "posting root init request.. OK. txid: %s\n", reqTx.ID().Base58())
	}

	// ---------- wait until the request is processed in all committee nodes
	if err = committee.WaitUntilAllRequestsProcessed(chainID, reqTx, 30*time.Second); err != nil {
		fmt.Fprintf(textout, "waiting root init request transaction.. FAILED: %v\n", err)
		return nil, nil, xerrors.Errorf("WaitUntilAllRequestsProcessed: %w", err)
	}

	fmt.Fprint(textout, par.Prefix)

	fmt.Fprintf(textout, "chain has been created succesfully on the Tangle. ChainID: %s, State address: %s, N = %d, T = %d\n",
		chainID.String(), stateControllerAddr.Base58(), par.N, par.T)

	return &chainID, stateControllerAddr, err
}
