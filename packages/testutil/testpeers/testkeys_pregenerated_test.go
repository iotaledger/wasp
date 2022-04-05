// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testpeers_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/stretchr/testify/require"
)

func TestPregenerateDKS(t *testing.T) {
	t.Skip("This test was used only to pre-generate the keys once.") // Comment that temporarily, if you need to regenerate the keys.
	// t.Run("N=1/F=0", func(t *testing.T) { testPregenerateDKS(t, 1, 0) }) // TODO: XXX: Uncomment, when LowN will be fixed.
	t.Run("N=4/F=0", func(t *testing.T) { testPregenerateDKS(t, 4, 0) })
	t.Run("N=4/F=1", func(t *testing.T) { testPregenerateDKS(t, 4, 1) })
	t.Run("N=10/F=3", func(t *testing.T) { testPregenerateDKS(t, 10, 3) })
	t.Run("N=22/F=7", func(t *testing.T) { testPregenerateDKS(t, 22, 7) })
	t.Run("N=31/F=10", func(t *testing.T) { testPregenerateDKS(t, 31, 10) })
	t.Run("N=40/F=13", func(t *testing.T) { testPregenerateDKS(t, 40, 13) })
	// t.Run("N=70/F=23", func(t *testing.T) { testPregenerateDKS(t, 70, 23) })	// TODO: XXX: Uncomment, when timeouts will be fixed.
	// t.Run("N=100/F=33", func(t *testing.T) { testPregenerateDKS(t, 100, 33) }) // TODO: XXX: Uncomment, when timeouts will be fixed.
}

func testPregenerateDKS(t *testing.T, n, f uint16) {
	var err error
	log := testlogger.NewLogger(t)
	defer log.Sync()
	threshold := n - f
	require.GreaterOrEqual(t, threshold, (n*2)/3+1)
	netIDs, identities := testpeers.SetupKeys(n)
	dksAddr, dksRegistries := testpeers.SetupDkg(t, threshold, netIDs, identities, tcrypto.DefaultBlsSuite(), log.Named("dkg"))
	var buf bytes.Buffer
	util.WriteUint16(&buf, uint16(len(dksRegistries)))
	for i := range dksRegistries {
		var dki *tcrypto.DKShareImpl
		var dkb []byte
		dkiInt, err := dksRegistries[i].LoadDKShare(dksAddr, identities[i].GetPrivateKey())
		dki = dkiInt.(*tcrypto.DKShareImpl)
		require.Nil(t, err)
		if i > 0 {
			// Remove it here to make serialized object smaller.
			// Will restore it from dks[0].
			dki.ClearCommonData()
		}
		// NodePubKeys will be set in the tests again, so we remove them here to save space.
		dki.AssignNodePubKeys(make([]*cryptolib.PublicKey, 0))
		dkb = dki.Bytes()
		require.Nil(t, util.WriteBytes16(&buf, dkb))
	}
	err = os.WriteFile(fmt.Sprintf("testkeys_pregenerated-%v-%v.bin", n, threshold), buf.Bytes(), 0o644)
	require.Nil(t, err)
}
