// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package apilib

import (
	"context"
	"fmt"
	"io"

	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/clients/multiclient"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/registry"
	"github.com/iotaledger/wasp/v2/packages/transaction"
)

type CreateChainParams struct {
	Layer1Client         clients.L1Client
	CommitteeAPIHosts    []string
	Signer               cryptolib.Signer
	Textout              io.Writer
	Prefix               string
	StateMetadata        transaction.StateMetadata
	GasCoinObjectID      *iotago.ObjectID
	GovernanceController *cryptolib.Address
	PackageID            iotago.PackageID
}

// DeployChain creates a new chain on specified committee address
func DeployChain(ctx context.Context, par CreateChainParams, anchorOwner *cryptolib.Address) error {
	var err error
	textout := io.Discard
	if par.Textout != nil {
		textout = par.Textout
	}
	fmt.Fprint(textout, par.Prefix)
	fmt.Fprintf(textout, "Creating new chain -- Anchor owner: %s\n*", anchorOwner)

	referenceGasPrice, err := par.Layer1Client.GetReferenceGasPrice(ctx)
	if err != nil {
		return isc.ChainID{}, fmt.Errorf("failed to get reference gas price: %w", err)
	}

	anchor, err := par.Layer1Client.L2().StartNewChain(
		ctx,
		&iscmoveclient.StartNewChainRequest{
			Signer:        par.Signer,
			AnchorOwner:   anchorOwner,
			PackageID:     par.PackageID,
			StateMetadata: par.StateMetadata.Bytes(),
			GasPrice:      referenceGasPrice.Uint64(),
			GasBudget:     iotaclient.DefaultGasBudget * 10,
		},
	)
	if err != nil {
		return isc.ChainID{}, fmt.Errorf("failed to call isc StartNewChain: %w", err)
	}

	fmt.Fprint(textout, par.Prefix)
	fmt.Fprintf(textout, "Chain has been created successfully on the Tangle.\n* ChainID: %s\n* State address: %s\n",
		anchor.ObjectID.String(), anchorOwner.String())

	fmt.Fprintf(textout, "Make sure to activate the chain on all committee nodes\n")

	return isc.ChainIDFromObjectID(*anchor.ObjectID), nil
}

// ActivateChainOnNodes puts chain records into nodes and activates its
func ActivateChainOnNodes(clientResolver multiclient.ClientResolver, apiHosts []string) error {
	nodes := multiclient.New(clientResolver, apiHosts)
	// ------------ put chain records to hosts
	return nodes.PutChainRecord(registry.NewChainRecord(true, []*cryptolib.PublicKey{}))
}
