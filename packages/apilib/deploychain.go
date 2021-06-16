// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package apilib

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/iotaledger/wasp/packages/util"

	"github.com/iotaledger/wasp/packages/registry/chainrecord"

	"github.com/iotaledger/wasp/packages/coretypes/chainid"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/client/goshimmer"
	"github.com/iotaledger/wasp/client/multiclient"
	"github.com/iotaledger/wasp/packages/transaction"
)

// TODO DeployChain on peering domain, not on committee

type CreateChainParams struct {
	Node                  *goshimmer.Client
	AllPeers              []string
	CommitteeApiHosts     []string
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
	if len(par.AllPeers) > 0 {
		// all committee nodes most also be among allPeers
		if !util.IsSubset(par.CommitteeApiHosts, par.AllPeers) {
			return nil, nil, xerrors.Errorf("DeployChainWithDKG: committee nodes must all be among peers")
		}
	}

	dkgInitiatorIndex := uint16(rand.Intn(len(par.CommitteeApiHosts)))
	stateControllerAddr, err := RunDKG(par.CommitteeApiHosts, par.CommitteePeeringHosts, par.T, dkgInitiatorIndex)
	if err != nil {
		return nil, nil, err
	}
	chainId, err := DeployChain(par, stateControllerAddr)
	if err != nil {
		return nil, nil, err
	}
	return chainId, stateControllerAddr, nil
}

// DeployChain creates a new chain on specified committee address
// noinspection ALL
func DeployChain(par CreateChainParams, stateControllerAddr ledgerstate.Address) (*chainid.ChainID, error) {
	allPeers := par.CommitteeApiHosts
	if len(par.AllPeers) > 0 {
		allPeers = par.AllPeers
	}
	var err error
	textout := ioutil.Discard
	if par.Textout != nil {
		textout = par.Textout
	}
	originatorAddr := ledgerstate.NewED25519Address(par.OriginatorKeyPair.PublicKey)

	fmt.Fprint(textout, par.Prefix)
	fmt.Fprintf(textout, "creating new chain. Owner address: %s. State controlled: %s, N = %d, T = %d\n",
		originatorAddr.Base58(), stateControllerAddr.Base58(), par.N, par.T)
	fmt.Fprint(textout, par.Prefix)

	chainID, initRequestTx, err := CreateChainOrigin(par.Node, par.OriginatorKeyPair, stateControllerAddr, par.Description)
	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "creating chain origin and init transaction.. FAILED: %v\n", err)
		return nil, xerrors.Errorf("DeployChain: %w", err)
	} else {
		fmt.Fprint(textout, "creating chain origin and init transaction.. OK\n")
	}

	err = ActivateChainOnAccessNodes(allPeers, chainID)
	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "activating chain.. FAILED: %v\n", err)
		return nil, xerrors.Errorf("DeployChain: %w", err)
	}
	fmt.Fprint(textout, "activating chain.. OK.\n")

	peers := multiclient.New(par.CommitteeApiHosts)

	// ---------- wait until the request is processed at least in all committee nodes
	if err = peers.WaitUntilAllRequestsProcessed(*chainID, initRequestTx, 30*time.Second); err != nil {
		fmt.Fprintf(textout, "waiting root init request transaction.. FAILED: %v\n", err)
		return nil, xerrors.Errorf("DeployChain: %w", err)
	}

	fmt.Fprint(textout, par.Prefix)
	fmt.Fprintf(textout, "chain has been created succesfully on the Tangle. ChainID: %s, State address: %s, N = %d, T = %d\n",
		chainID.String(), stateControllerAddr.Base58(), par.N, par.T)

	return chainID, err
}

// CreateChainOrigin creates and confirms origin transaction of the chain and init request transaction to initialize state of it
func CreateChainOrigin(node *goshimmer.Client, originator *ed25519.KeyPair, stateController ledgerstate.Address, dscr string) (*chainid.ChainID, *ledgerstate.Transaction, error) {
	originatorAddr := ledgerstate.NewED25519Address(originator.PublicKey)
	// ----------- request owner address' outputs from the ledger
	allOuts, err := node.GetConfirmedOutputs(originatorAddr)
	if err != nil {
		return nil, nil, xerrors.Errorf("CreateChainOrigin: %w", err)
	}

	// ----------- create origin transaction
	originTx, chainID, err := transaction.NewChainOriginTransaction(
		originator,
		stateController,
		nil,
		time.Now(),
		allOuts...,
	)
	if err != nil {
		return nil, nil, xerrors.Errorf("CreateChainOrigin: %w", err)
	}

	// ------------- post origin transaction and wait for confirmation
	err = node.PostAndWaitForConfirmation(originTx)
	if err != nil {
		return nil, nil, xerrors.Errorf("CreateChainOrigin: %w", err)
	}

	allOuts, err = node.GetConfirmedOutputs(originatorAddr)
	if err != nil {
		return nil, nil, xerrors.Errorf("CreateChainOrigin: %w", err)
	}

	// NOTE: whoever send first init request, is an owner of the chain
	// create root init transaction
	reqTx, err := transaction.NewRootInitRequestTransaction(
		originator,
		chainID,
		dscr,
		time.Now(),
		allOuts...,
	)
	if err != nil {
		return nil, nil, xerrors.Errorf("CreateChainOrigin: %w", err)
	}

	// ---------- post root init request transaction and wait for confirmation
	err = node.PostAndWaitForConfirmation(reqTx)
	if err != nil {
		return nil, nil, xerrors.Errorf("CreateChainOrigin: %w", err)
	}

	return &chainID, reqTx, nil
}

// ActivateChainOnAccessNodes puts chain records into nodes and activates its
// TODO needs refactoring and optimization
func ActivateChainOnAccessNodes(apiHosts []string, chainID *chainid.ChainID) error {
	peers := multiclient.New(apiHosts)
	// ------------ put chain records to hosts
	err := peers.PutChainRecord(&chainrecord.ChainRecord{
		ChainID: chainID,
	})
	if err != nil {
		return xerrors.Errorf("ActivateChainOnAccessNodes: %w")
	}
	// ------------- activate chain
	err = peers.ActivateChain(*chainID)
	if err != nil {
		return xerrors.Errorf("ActivateChainOnAccessNodes: %w")
	}
	return nil
}
