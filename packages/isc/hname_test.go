// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isc_test

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
)

func TestHnameSerialize(t *testing.T) {
	data1 := make([]byte, isc.HnameLength)
	rand.Read(data1)
	hname1, err := isc.HnameFromBytes(data1)
	require.NoError(t, err)
	data2 := hname1.Bytes()
	require.True(t, bytes.Equal(data1, data2))

	s := hname1.String()
	hname2, err := isc.HnameFromHexString(s)
	require.NoError(t, err)
	require.EqualValues(t, hname1, hname2)
}

func TestHnameCollision(t *testing.T) {
	hn1 := isc.Hn("doNothing")
	hn2 := isc.Hn("incCounter")

	require.NotEqualValues(t, hn1, hn2)
}
