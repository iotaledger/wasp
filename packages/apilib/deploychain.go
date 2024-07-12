// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package apilib

import (
	"context"
	"errors"
	"fmt"
	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/multiclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

// TODO DeployChain on peering domain, not on committee

type CreateChainParams struct {
	Layer1Client         clients.L1Client
	CommitteeAPIHosts    []string
	N                    uint16
	T                    uint16
	OriginatorKeyPair    cryptolib.Signer
	Textout              io.Writer
	Prefix               string
	InitParams           dict.Dict
	GovernanceController *cryptolib.Address
	PackageID            sui.PackageID
}

// DeployChain creates a new chain on specified committee address
func DeployChain(ctx context.Context, par CreateChainParams, stateControllerAddr, govControllerAddr *cryptolib.Address) (isc.ChainID, error) {
	var err error
	textout := io.Discard
	if par.Textout != nil {
		textout = par.Textout
	}
	originatorAddr := par.OriginatorKeyPair.Address()

	fmt.Fprint(textout, par.Prefix)
	fmt.Fprintf(textout, "Creating new chain\n* Owner address:    %s\n* State controller: %s\n* committee size = %d\n* quorum = %d\n",
		originatorAddr, stateControllerAddr, par.N, par.T)
	fmt.Fprint(textout, par.Prefix)

	anchorBytes, err := par.Layer1Client.L2Client().StartNewChain(ctx, par.OriginatorKeyPair, par.PackageID, nil, suiclient.DefaultGasPrice, suiclient.DefaultGasBudget, par.InitParams.Bytes(), false)
	if err != nil {
		return isc.ChainID{}, err
	}

	txnResponse, err := par.Layer1Client.SignAndExecuteTransaction(
		context.Background(),
		cryptolib.SignerToSuiSigner(par.OriginatorKeyPair),
		anchorBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		})
	if err != nil {
		return isc.ChainID{}, err
	}

	anchor, err := par.Layer1Client.L2Client().GetAnchorFromSuiTransactionBlockResponse(context.Background(), txnResponse)
	if err != nil {
		fmt.Fprintf(textout, "Creating chain origin and init transaction.. FAILED: %v\n", err)
		return isc.ChainID{}, fmt.Errorf("DeployChain: %w", err)
	}
	
	fmt.Fprint(textout, par.Prefix)
	fmt.Fprintf(textout, "Chain has been created successfully on the Tangle.\n* ChainID: %s\n* State address: %s\n* committee size = %d\n* quorum = %d\n",
		anchor.Ref.ObjectID.String(), stateControllerAddr.String(), par.N, par.T)

	fmt.Fprintf(textout, "Make sure to activate the chain on all committee nodes\n")

	return isc.ChainIDFromObjectID(*anchor.Ref.ObjectID), err
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
	layer1Client clients.L1Client,
	originator cryptolib.Signer,
	stateController *cryptolib.Address,
	governanceController *cryptolib.Address,
	initParams dict.Dict,
) (isc.ChainID, error) {

	// originatorAddr := originator.Address()
	// ----------- request owner address' outputs from the ledger
	/*
		utxoMap, err := layer2Client.OutputMap(originatorAddr)
		if err != nil {
			return isc.ChainID{}, fmt.Errorf("CreateChainOrigin: %w", err)
		}
	*/

	// ----------- create origin transaction
	panic("refactor me: origin.NewChainOriginTransaction")
	var chainID isc.ChainID
	err := errors.New("refactor me: CreateChainOrigin")

	if err != nil {
		return isc.ChainID{}, fmt.Errorf("CreateChainOrigin: %w", err)
	}

	// ------------- post origin transaction and wait for confirmation
	/*_, err = layer2Client.PostTxAndWaitUntilConfirmation(originTx)
	if err != nil {
		return isc.ChainID{}, fmt.Errorf("CreateChainOrigin: %w", err)
	}*/

	return chainID, nil
}

// ActivateChainOnNodes puts chain records into nodes and activates its
func ActivateChainOnNodes(clientResolver multiclient.ClientResolver, apiHosts []string, chainID isc.ChainID) error {
	nodes := multiclient.New(clientResolver, apiHosts)
	// ------------ put chain records to hosts
	return nodes.PutChainRecord(registry.NewChainRecord(chainID, true, []*cryptolib.PublicKey{}))
}
