// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package apilib provides utility functions for API operations and interactions.
package apilib

import (
	"context"
	"fmt"
	"math"
	"time"

	"fortio.org/safecast"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

// RunDKG runs DKG procedure on specific Wasp hosts: generates new keys and puts corresponding committee records
// into nodes. In case of success, generated address is returned
func RunDKG(ctx context.Context, client *apiclient.APIClient, peerPubKeys []string, threshold uint16, timeout ...time.Duration) (*cryptolib.Address, error) {
	to := uint32(60 * 1000)
	if len(timeout) > 0 {
		n := timeout[0].Milliseconds()
		if n < int64(math.MaxUint16) {
			val, err := safecast.Convert[uint32](n)
			if err != nil {
				return nil, err
			}
			to = val
		}
	}

	dkShares, _, err := client.NodeAPI.GenerateDKS(ctx).DKSharesPostRequest(apiclient.DKSharesPostRequest{
		Threshold:      uint32(threshold),
		TimeoutMS:      to,
		PeerIdentities: peerPubKeys,
	}).Execute()
	if err != nil {
		return nil, err
	}

	addr, err := cryptolib.NewAddressFromHexString(dkShares.Address)
	if err != nil {
		return nil, fmt.Errorf("RunDKG: invalid address returned from DKG: %w", err)
	}

	return addr, nil
}
