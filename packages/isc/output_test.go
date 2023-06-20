// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isc_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
)

func TestAliasOutputWithIDSerialization(t *testing.T) {
	outputTest := isc.NewAliasOutputWithID(&iotago.AliasOutput{}, iotago.OutputID{})
	data1 := outputTest.Bytes()
	output, err := isc.AliasOutputWithIDFromBytes(data1)
	require.NoError(t, err)
	require.Equal(t, outputTest, output)
	data2 := output.Bytes()
	require.Equal(t, data1, data2)
}
