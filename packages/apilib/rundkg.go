// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package apilib

import (
	"errors"
	"fmt"
	"math"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/webapi/v1/model"
)

// RunDKG runs DKG procedure on specific Wasp hosts: generates new keys and puts corresponding committee records
// into nodes. In case of success, generated address is returned
func RunDKG(authToken string, apiHosts, peerPubKeys []string, threshold, initiatorIndex uint16, timeout ...time.Duration) (iotago.Address, error) {
	to := uint32(60 * 1000)
	if len(timeout) > 0 {
		n := timeout[0].Milliseconds()
		if n < int64(math.MaxUint16) {
			to = uint32(n)
		}
	}
	if int(initiatorIndex) >= len(apiHosts) {
		return nil, errors.New("RunDKG: wrong initiator index")
	}
	dkShares, err := client.NewWaspClient(apiHosts[initiatorIndex]).WithToken(authToken).DKSharesPost(&model.DKSharesPostRequest{
		PeerPubKeys: peerPubKeys,
		Threshold:   threshold,
		TimeoutMS:   to, // 1 min
	})
	if err != nil {
		return nil, err
	}
	_, addr, err := iotago.ParseBech32(dkShares.Address)
	if err != nil {
		return nil, fmt.Errorf("RunDKG: invalid address returned from DKG: %w", err)
	}

	return addr, nil
}
