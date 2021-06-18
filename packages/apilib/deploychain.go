// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package apilib

import (
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/iotaledger/wasp/packages/registry/chainrecord"

	"github.com/iotaledger/wasp/packages/coretypes/chainid"

	"github.com/iotaledger/wasp/packages/registry/committee_record"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/client/goshimmer"
	"github.com/iotaledger/wasp/client/multiclient"
	"github.com/iotaledger/wasp/packages/transaction"
)

type CreateChainParams struct {
	Node                  *goshimmer.Client
	CommitteeAPIHosts     []string
	CommitteePeeringHosts []string
	N                     uint16
	T                     uint16
	OriginatorKeyPair     *ed25519.KeyPair
	Description           string
	Textout               io.Writer
	Prefix                string
}

// DeployChainWithDKG performs all actions needed to deploy the chain
// TODO: [KP] Shouldn't that be in the client packages?
func DeployChainWithDKG(par CreateChainParams) (*chainid.ChainID, ledgerstate.Address, error) {
	stateControllerAddr, err := RunDKG(par.CommitteeAPIHosts, par.CommitteePeeringHosts, par.T)
	if err != nil {
		return nil, nil, err
	}
	chainID, err := DeployChain(par, stateControllerAddr)
	if err != nil {
		return nil, nil, err
	}
	return chainID, stateControllerAddr, nil
}

// DeployChain creates a new chain on specified committee address
// noinspection ALL
//nolint:funlen
func DeployChain(par CreateChainParams, stateControllerAddr ledgerstate.Address) (*chainid.ChainID, error) {
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

	committee := multiclient.New(par.CommitteeAPIHosts)

	// ------------ put committee records to hosts
	err = committee.PutCommitteeRecord(&committee_record.CommitteeRecord{
		Address: stateControllerAddr,
		Nodes:   par.CommitteePeeringHosts,
	})
	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "sending committee record to nodes.. FAILED: %v\n", err)
		return nil, xerrors.Errorf("PutCommitteeRecord: %w", err)
	}
	fmt.Fprint(textout, "sending committee record to nodes.. OK\n")

	// ----------- request owner address' outputs from the ledger
	allOuts, err := par.Node.GetConfirmedOutputs(originatorAddr)

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "requesting UTXOs from node.. FAILED: %v\n", err)
		return nil, xerrors.Errorf("GetConfirmedOutputs: %w", err)
	}
	fmt.Fprint(textout, "requesting owner address' UTXOs from node.. OK\n")

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
		return nil, xerrors.Errorf("NewChainOriginTransaction: %w", err)
	}
	fmt.Fprintf(textout, "creating origin transaction.. OK. Origin txid = %s\n", originTx.ID().String())

	// ------------- post origin transaction and wait for confirmation
	err = par.Node.PostAndWaitForConfirmation(originTx)
	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "posting origin transaction.. FAILED: %v\n", err)
		return nil, xerrors.Errorf("posting origin transaction: %w", err)
	}
	fmt.Fprintf(textout, "posting origin transaction.. OK. txid: %s\n", originTx.ID().Base58())

	// ------------ put chain records to hosts
	err = committee.PutChainRecord(&chainrecord.ChainRecord{
		ChainID: &chainID,
	})

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "sending chain data to Wasp nodes.. FAILED: %v\n", err)
		return nil, xerrors.Errorf("PutChainRecord: %w", err)
	}
	fmt.Fprint(textout, "sending chain data to Wasp nodes.. OK.\n")

	// ------------- activate chain
	err = committee.ActivateChain(chainID)

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "activating chain.. FAILED: %v\n", err)
		return nil, xerrors.Errorf("ActivateChain: %w", err)
	}
	fmt.Fprint(textout, "activating chain.. OK.\n")

	// ============= create root init request for the root contract

	// ------------- get UTXOs of the owner
	allOuts, err = par.Node.GetConfirmedOutputs(originatorAddr)
	if err != nil {
		fmt.Fprintf(textout, "GetConfirmedOutputs.. FAILED: %v\n", err)
		return nil, xerrors.Errorf("GetConfirmedOutputs: %w", err)
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
		return nil, xerrors.Errorf("NewRootInitRequestTransaction: %w", err)
	}
	fmt.Fprintf(textout, "creating root init request.. OK\n")
	fmt.Fprintf(textout, "root init txid: %s, reqidBase58: %s\n", reqTx.ID().Base58(), reqTx.Essence().Outputs()[0].ID().Base58())

	// ---------- post root init request transaction and wait for confirmation
	err = par.Node.PostAndWaitForConfirmation(reqTx)
	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "posting root init request transaction.. FAILED: %v\n", err)
		return nil, xerrors.Errorf("posting root init request: %w", err)
	}
	fmt.Fprintf(textout, "posting root init request.. OK. txid: %s\n", reqTx.ID().Base58())

	// ---------- wait until the request is processed in all committee nodes
	if err = committee.WaitUntilAllRequestsProcessed(chainID, reqTx, 30*time.Second); err != nil {
		fmt.Fprintf(textout, "waiting root init request transaction.. FAILED: %v\n", err)
		return nil, xerrors.Errorf("WaitUntilAllRequestsProcessed: %w", err)
	}

	fmt.Fprint(textout, par.Prefix)

	fmt.Fprintf(textout, "chain has been created successfully on the Tangle. ChainID: %s, State address: %s, N = %d, T = %d\n",
		chainID.String(), stateControllerAddr.Base58(), par.N, par.T)

	return &chainID, err
}
