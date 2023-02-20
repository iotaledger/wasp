// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/cmtLog"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

func TestVarLocalView(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()
	j := cmtLog.NewVarLocalView(log)
	require.Nil(t, j.Value())
	tipAO, ok := j.AliasOutputConfirmed(isc.NewAliasOutputWithID(&iotago.AliasOutput{}, iotago.OutputID{}))
	require.True(t, ok)
	require.NotNil(t, tipAO)
	require.NotNil(t, j.Value())
	require.Equal(t, tipAO, j.Value())
}
