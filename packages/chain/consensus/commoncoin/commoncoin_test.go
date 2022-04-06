// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package commoncoin_test

import (
	"fmt"
	"testing"

	"github.com/anthdm/hbbft"
	"github.com/iotaledger/wasp/packages/chain/consensus/commoncoin"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/stretchr/testify/require"
)

func TestBlsCommonCoin(t *testing.T) {
	var err error
	netIDs, identities := testpeers.SetupKeys(10)
	address, regProviders := testpeers.SetupDkgPregenerated(t, 7, identities, tcrypto.DefaultBlsSuite())

	ccs := make([]hbbft.CommonCoin, len(netIDs))
	salt := []byte{0, 1, 2, 3}
	for i := range ccs {
		var dkShare tcrypto.DKShare
		dkShare, err = regProviders[i].LoadDKShare(address)
		require.NoError(t, err)
		ccs[i] = commoncoin.NewBlsCommonCoin(dkShare, salt, true)
	}

	for epoch := uint32(0); epoch <= 10; epoch++ {
		t.Run(fmt.Sprintf("TestBlsCommonCoin/epoch=%v", epoch), func(tt *testing.T) {
			testBlsCommonCoin(tt, epoch, ccs)
		})
	}
}

// Runs CC for a single epoch.
func testBlsCommonCoin(t *testing.T, epoch uint32, ccs []hbbft.CommonCoin) {
	var err error
	coins := make([]*bool, len(ccs))
	msgs := make([]interface{}, 0)

	for i := range ccs {
		var ccsMsgs []interface{}
		coins[i], ccsMsgs, err = ccs[i].StartCoinFlip(epoch)
		require.NoError(t, err)
		msgs = append(msgs, ccsMsgs...)
	}

	for len(msgs) != 0 {
		msg := msgs[0]
		msgs = msgs[1:]
		for i := range ccs {
			var ccsMsgs []interface{}
			coins[i], ccsMsgs, err = ccs[i].HandleRequest(epoch, msg)
			require.NoError(t, err)
			msgs = append(msgs, ccsMsgs...)
		}
	}

	require.NotNil(t, coins[0])
	for i := range coins {
		require.Equal(t, coins[0], coins[i])
	}
}
