// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package apilib

import (
	"fmt"
	"io"
	"math/rand"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/client/multiclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/nodeconn"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/transaction"
	"golang.org/x/xerrors"
)

// TODO DeployChain on peering domain, not on committee

type CreateChainParams struct {
	Layer1Client      nodeconn.L1Client
	CommitteeAPIHosts []string
	CommitteePubKeys  []string
	N                 uint16
	T                 uint16
	OriginatorKeyPair *cryptolib.KeyPair
	Description       string
	Textout           io.Writer
	Prefix            string
	InitParams        dict.Dict
}

// DeployChainWithDKG performs all actions needed to deploy the chain
// TODO: [KP] Shouldn't that be in the client packages?
func DeployChainWithDKG(par CreateChainParams) (*iscp.ChainID, iotago.Address, error) {
	dkgInitiatorIndex := uint16(rand.Intn(len(par.CommitteeAPIHosts)))
	stateControllerAddr, err := RunDKG(par.CommitteeAPIHosts, par.CommitteePubKeys, par.T, dkgInitiatorIndex)
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

func DeployChain(par CreateChainParams, stateControllerAddr iotago.Address) (*iscp.ChainID, error) {
	var err error
	textout := io.Discard
	if par.Textout != nil {
		textout = par.Textout
	}
	originatorAddr := par.OriginatorKeyPair.GetPublicKey().AsEd25519Address()

	fmt.Fprint(textout, par.Prefix)
	fmt.Fprintf(textout, "creating new chain. Owner address: %s. State controller: %s, N = %d, T = %d\n",
		originatorAddr, stateControllerAddr, par.N, par.T)
	fmt.Fprint(textout, par.Prefix)

	chainID, initRequestTx, err := CreateChainOrigin(par.Layer1Client, par.OriginatorKeyPair, stateControllerAddr, par.Description, par.InitParams)
	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "creating chain origin and init transaction.. FAILED: %v\n", err)
		return nil, xerrors.Errorf("DeployChain: %w", err)
	}
	txID, err := initRequestTx.ID()
	if err != nil {
		fmt.Fprintf(textout, "creating chain origin and init transaction.. FAILED: %v\n", err)
		return nil, xerrors.Errorf("DeployChain: %w", err)
	}
	fmt.Fprintf(textout, "creating chain origin and init transaction %s.. OK\n", txID)
	fmt.Fprint(textout, "sending committee record to nodes.. OK\n")

	err = ActivateChainOnAccessNodes(par.CommitteeAPIHosts, chainID)
	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "activating chain %s.. FAILED: %v\n", chainID.AsAddress(), err)
		return nil, xerrors.Errorf("DeployChain: %w", err)
	}
	fmt.Fprintf(textout, "activating chain %s.. OK.\n", chainID.AsAddress())

	peers := multiclient.New(par.CommitteeAPIHosts)

	// ---------- wait until the request is processed at least in all committee nodes
	if err = peers.WaitUntilAllRequestsProcessed(chainID, initRequestTx, 30*time.Second); err != nil {
		fmt.Fprintf(textout, "waiting root init request transaction.. FAILED: %v\n", err)
		return nil, xerrors.Errorf("DeployChain: %w", err)
	}

	fmt.Fprint(textout, par.Prefix)
	fmt.Fprintf(textout, "chain has been created successfully on the Tangle. ChainID: %s, State address: %s, N = %d, T = %d\n",
		chainID.String(), stateControllerAddr, par.N, par.T)

	return chainID, err
}

func utxoIDsFromUtxoMap(utxoMap iotago.OutputSet) iotago.OutputIDs {
	var utxoIDs iotago.OutputIDs
	for id := range utxoMap {
		utxoIDs = append(utxoIDs, id)
	}
	return utxoIDs
}

// CreateChainOrigin creates and confirms origin transaction of the chain and init request transaction to initialize state of it
func CreateChainOrigin(layer1Client nodeconn.L1Client, originator *cryptolib.KeyPair, stateController iotago.Address, dscr string, initParams dict.Dict) (*iscp.ChainID, *iotago.Transaction, error) {
	originatorAddr := originator.GetPublicKey().AsEd25519Address()
	// ----------- request owner address' outputs from the ledger
	utxoMap, err := layer1Client.OutputMap(originatorAddr)
	if err != nil {
		return nil, nil, xerrors.Errorf("CreateChainOrigin: %w", err)
	}

	// ----------- create origin transaction
	originTx, chainID, err := transaction.NewChainOriginTransaction(
		originator,
		stateController,
		stateController,
		0,
		utxoMap,
		utxoIDsFromUtxoMap(utxoMap),
		layer1Client.L1Params(),
	)
	if err != nil {
		return nil, nil, xerrors.Errorf("CreateChainOrigin: %w", err)
	}

	// ------------- post origin transaction and wait for confirmation
	err = layer1Client.PostTx(originTx)
	if err != nil {
		return nil, nil, xerrors.Errorf("CreateChainOrigin: %w", err)
	}

	utxoMap, err = layer1Client.OutputMap(originatorAddr)
	if err != nil {
		return nil, nil, xerrors.Errorf("CreateChainOrigin: %w", err)
	}

	// NOTE: whoever send first init request, is an owner of the chain
	// create root init transaction
	reqTx, err := transaction.NewRootInitRequestTransaction(
		originator,
		chainID,
		dscr,
		utxoMap,
		utxoIDsFromUtxoMap(utxoMap),
		layer1Client.L1Params(),
		initParams,
	)
	if err != nil {
		return nil, nil, xerrors.Errorf("CreateChainOrigin: %w", err)
	}

	// ---------- post root init request transaction and wait for confirmation
	err = layer1Client.PostTx(reqTx)
	if err != nil {
		return nil, nil, xerrors.Errorf("CreateChainOrigin: %w", err)
	}

	return chainID, reqTx, nil
}

// ActivateChainOnAccessNodes puts chain records into nodes and activates its
// TODO needs refactoring and optimization
func ActivateChainOnAccessNodes(apiHosts []string, chainID *iscp.ChainID) error {
	nodes := multiclient.New(apiHosts)
	// ------------ put chain records to hosts
	err := nodes.PutChainRecord(&registry.ChainRecord{
		ChainID: *chainID,
	})
	if err != nil {
		return xerrors.Errorf("ActivateChainOnAccessNodes: %w", err)
	}
	// ------------- activate chain
	err = nodes.ActivateChain(chainID)
	if err != nil {
		return xerrors.Errorf("ActivateChainOnAccessNodes: %w", err)
	}
	return nil
}
