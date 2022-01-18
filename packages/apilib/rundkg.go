// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package apilib

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"golang.org/x/xerrors"
)

// RunDKG runs DKG procedure on specific Wasp hosts: generates new keys and puts corresponding committee records
// into nodes. In case of success, generated address is returned
func RunDKG(apiHosts, peerPubKeys []string, threshold, initiatorIndex uint16, timeout ...time.Duration) (ledgerstate.Address, error) {
	to := uint32(60 * 1000)
	if len(timeout) > 0 {
		n := timeout[0].Milliseconds()
		if n < int64(util.MaxUint16) {
			to = uint32(n)
		}
	}
	if int(initiatorIndex) >= len(apiHosts) {
		return nil, xerrors.New("RunDKG: wrong initiator index")
	}
	dkShares, err := client.NewWaspClient(apiHosts[initiatorIndex]).DKSharesPost(&model.DKSharesPostRequest{
		PeerPubKeys: peerPubKeys,
		Threshold:   threshold,
		TimeoutMS:   to, // 1 min
	})
	if err != nil {
		return nil, err
	}
	var addr ledgerstate.Address
	if addr, err = ledgerstate.AddressFromBase58EncodedString(dkShares.Address); err != nil {
		return nil, xerrors.Errorf("RunDKG: invalid address returned from DKG: %w", err)
	}

	return addr, nil
}
