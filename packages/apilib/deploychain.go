// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package apilib

import (
	"fmt"
	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/multiclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/l1connection"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/transaction"
)

// TODO DeployChain on peering domain, not on committee

type CreateChainParams struct {
	Layer1Client         l1connection.Client
	CommitteeAPIHosts    []string
	N                    uint16
	T                    uint16
	OriginatorKeyPair    *cryptolib.KeyPair
	Textout              io.Writer
	Prefix               string
	InitParams           dict.Dict
	GovernanceController iotago.Address
}

// DeployChain creates a new chain on specified committee address
// noinspection ALL

func DeployChain(par CreateChainParams, stateControllerAddr, govControllerAddr iotago.Address) (isc.ChainID, error) {
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

	chainID, err := CreateChainOrigin(
		par.Layer1Client,
		par.OriginatorKeyPair,
		stateControllerAddr,
		govControllerAddr,
		par.InitParams,
	)
	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "creating chain origin and init transaction.. FAILED: %v\n", err)
		return isc.ChainID{}, fmt.Errorf("DeployChain: %w", err)
	}
	fmt.Fprint(textout, par.Prefix)
	fmt.Fprintf(textout, "chain has been created successfully on the Tangle. ChainID: %s, State address: %s, N = %d, T = %d\n",
		chainID.String(), stateControllerAddr.Bech32(parameters.L1().Protocol.Bech32HRP), par.N, par.T)

	fmt.Fprintf(textout, "make sure to activate the sure on all committee nodes\n")

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
func CreateChainOrigin(
	layer1Client l1connection.Client,
	originator *cryptolib.KeyPair,
	stateController iotago.Address,
	governanceController iotago.Address,
	initParams dict.Dict,
) (isc.ChainID, error) {
	originatorAddr := originator.GetPublicKey().AsEd25519Address()
	// ----------- request owner address' outputs from the ledger
	utxoMap, err := layer1Client.OutputMap(originatorAddr)
	if err != nil {
		return isc.ChainID{}, fmt.Errorf("CreateChainOrigin: %w", err)
	}

	// ----------- create origin transaction
	originTx, _, chainID, err := transaction.NewChainOriginTransaction(
		originator,
		stateController,
		governanceController,
		0,
		initParams,
		utxoMap,
		utxoIDsFromUtxoMap(utxoMap),
	)
	if err != nil {
		return isc.ChainID{}, fmt.Errorf("CreateChainOrigin: %w", err)
	}

	// ------------- post origin transaction and wait for confirmation
	_, err = layer1Client.PostTxAndWaitUntilConfirmation(originTx)
	if err != nil {
		return isc.ChainID{}, fmt.Errorf("CreateChainOrigin: %w", err)
	}

	return chainID, nil
}

// ActivateChainOnNodes puts chain records into nodes and activates its
func ActivateChainOnNodes(clientResolver multiclient.ClientResolver, apiHosts []string, chainID isc.ChainID) error {
	nodes := multiclient.New(clientResolver, apiHosts)
	// ------------ put chain records to hosts
	return nodes.PutChainRecord(registry.NewChainRecord(chainID, true, []*cryptolib.PublicKey{}))
}
