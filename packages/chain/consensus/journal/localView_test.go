// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package journal_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/isc"
)

func TestLocalView(t *testing.T) {
	j := journal.NewLocalView()
	require.Nil(t, j.GetBaseAliasOutputID())
	j.AliasOutputReceived(isc.NewAliasOutputWithID(&iotago.AliasOutput{}, &iotago.UTXOInput{}))
	require.NotNil(t, j.GetBaseAliasOutputID())
	jBin, err := j.AsBytes()
	require.NoError(t, err)
	jj, err := journal.NewLocalViewFromBytes(jBin)
	require.NotNil(t, jj.GetBaseAliasOutputID())
	require.NoError(t, err)
	require.NotNil(t, jj)
}
