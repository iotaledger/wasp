// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAgentIDCoretypes(t *testing.T) {
	aid := NewRandomAgentID()

	t.Logf("random AgentID = %s", aid.String())

	chainID := RandomChainID()

	hname := Hn("dummy")
	aid = NewContractAgentID(chainID, hname)

	t.Logf("agent ID string: %s", aid.String())

	aidBack, err := AgentIDFromBytes(aid.Bytes())
	require.NoError(t, err)
	require.True(t, aid.Equals(aidBack))

	aid = NewContractAgentID(chainID, hname)
	require.True(t, chainID.AsAddress().Equal(aid.(*ContractAgentID).Address()))
	require.EqualValues(t, aid.(*ContractAgentID).Hname(), hname)

	aidBack, err = NewAgentIDFromString(aid.String())
	require.NoError(t, err)
	require.True(t, aid.Equals(aidBack))

	aidBack, err = NewAgentIDFromString(aid.String())
	require.NoError(t, err)
	require.True(t, aid.Equals(aidBack))
}

func TestHname(t *testing.T) {
	hn1 := Hn("first")

	hn1bytes := hn1.Bytes()
	hn1back, err := HnameFromBytes(hn1bytes)
	require.NoError(t, err)
	require.EqualValues(t, hn1, hn1back)

	s := hn1.String()
	hn1back, err = HnameFromString(s)
	require.NoError(t, err)
	require.EqualValues(t, hn1, hn1back)
}

func TestHnameCollision(t *testing.T) {
	hn1 := Hn("doNothing")
	hn2 := Hn("incCounter")

	require.NotEqualValues(t, hn1, hn2)
}
