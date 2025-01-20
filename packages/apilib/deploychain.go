// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package apilib

import (
	"context"
	"fmt"
	"io"

	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/clients/multiclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/transaction"
)

type CreateChainParams struct {
	Layer1Client         clients.L1Client
	CommitteeAPIHosts    []string
	N                    uint16
	T                    uint16
	OriginatorKeyPair    cryptolib.Signer
	Textout              io.Writer
	Prefix               string
	StateMetadata        transaction.StateMetadata
	GasCoinObjectID      *iotago.ObjectID
	GovernanceController *cryptolib.Address
	PackageID            iotago.PackageID
}

// DeployChain creates a new chain on specified committee address
func DeployChain(ctx context.Context, par CreateChainParams, stateControllerAddr *cryptolib.Address) (isc.ChainID, error) {
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

	referenceGasPrice, err := par.Layer1Client.GetReferenceGasPrice(ctx)
	if err != nil {
		return isc.ChainID{}, err
	}

	var gasPayments []*iotago.ObjectRef
	if par.GasCoinObjectID != nil {
		resGetObj, err := par.Layer1Client.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: par.StateMetadata.GasCoinObjectID})
		if err != nil {
			return isc.ChainID{}, err
		}
		ref := resGetObj.Data.Ref()
		gasPayments = append(gasPayments, &ref)
	}

	anchor, err := par.Layer1Client.L2().StartNewChain(
		ctx,
		&iscmoveclient.StartNewChainRequest{
			Signer:            par.OriginatorKeyPair,
			ChainOwnerAddress: stateControllerAddr,
			PackageID:         par.PackageID,
			StateMetadata:     par.StateMetadata.Bytes(),
			GasPayments:       gasPayments,
			GasPrice:          referenceGasPrice.Uint64(),
			GasBudget:         iotaclient.DefaultGasBudget * 10,
		},
	)
	if err != nil {
		return isc.ChainID{}, err
	}

	fmt.Fprint(textout, par.Prefix)
	fmt.Fprintf(textout, "Chain has been created successfully on the Tangle.\n* ChainID: %s\n* State address: %s\n* committee size = %d\n* quorum = %d\n",
		anchor.ObjectID.String(), stateControllerAddr.String(), par.N, par.T)

	fmt.Fprintf(textout, "Make sure to activate the chain on all committee nodes\n")

	return isc.ChainIDFromObjectID(*anchor.ObjectID), err
}

// ActivateChainOnNodes puts chain records into nodes and activates its
func ActivateChainOnNodes(clientResolver multiclient.ClientResolver, apiHosts []string, chainID isc.ChainID) error {
	nodes := multiclient.New(clientResolver, apiHosts)
	// ------------ put chain records to hosts
	return nodes.PutChainRecord(registry.NewChainRecord(chainID, true, []*cryptolib.PublicKey{}))
}
