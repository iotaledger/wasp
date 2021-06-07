// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testpeers_test

import (
	"fmt"
	"testing"

	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
)

func TestPregenerateDKS(t *testing.T) {
	t.Skip("This test was used only to pre-generate the keys once.")
	t.Run("N=1/F=0", func(t *testing.T) { testPregenerateDKS(t, 1) })
	t.Run("N=4/F=1", func(t *testing.T) { testPregenerateDKS(t, 4) })
	t.Run("N=10/F=3", func(t *testing.T) { testPregenerateDKS(t, 10) })
	t.Run("N=22/F=7", func(t *testing.T) { testPregenerateDKS(t, 22) })
	t.Run("N=31/F=10", func(t *testing.T) { testPregenerateDKS(t, 31) })
	t.Run("N=40/F=13", func(t *testing.T) { testPregenerateDKS(t, 40) })
	t.Run("N=70/F=23", func(t *testing.T) { testPregenerateDKS(t, 70) })
	t.Run("N=100/F=33", func(t *testing.T) { testPregenerateDKS(t, 100) })
}

func testPregenerateDKS(t *testing.T, N uint16) {
	var err error
	suite := pairing.NewSuiteBn256()
	log := testlogger.NewLogger(t)
	defer log.Sync()
	netIDs, pubKeys, privKeys := testpeers.SetupKeys(N, suite)
	dksAddr, dksRegistries := testpeers.SetupDkg(t, uint16((len(netIDs)*2)/3+1), netIDs, pubKeys, privKeys, suite, log.Named("dkg"))
	fmt.Printf("func pregeneratedDks%v() []string {\n", N)
	fmt.Printf("\tdks := make([]string, %v)\n", N)
	for i := range dksRegistries {
		var dki *tcrypto.DKShare
		var dkb []byte
		dki, err = dksRegistries[i].LoadDKShare(dksAddr)
		require.Nil(t, err)
		if i > 0 {
			// Remove it here to make serialized object smaller.
			// Will restore it from dks[0].
			dki.PublicCommits = make([]kyber.Point, 0)
			dki.PublicShares = make([]kyber.Point, 0)
		}
		dkb, err = dki.Bytes()
		require.Nil(t, err)
		fmt.Printf("\tdks[%v] = ", i)
		printWrappedText(base58.Encode(dkb))
	}
	fmt.Printf("\treturn dks\n")
	fmt.Printf("}\n\n")
}

func printWrappedText(s string) {
	fmt.Printf("\"\" +\n")
	for {
		if len(s) == 0 {
			break
		}
		if len(s) > 80 {
			fmt.Printf("\t\t\"%v\" +\n", s[:80])
			s = s[80:]
			continue
		}
		fmt.Printf("\t\t\"%v\"\n", s)
		break
	}
}
