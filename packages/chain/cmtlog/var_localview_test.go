// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtlog_test

// TODO: Re-enable this test.

// import (
// 	"testing"

// 	"github.com/stretchr/testify/require"

// 	"github.com/iotaledger/wasp/packages/chain/cmtlog"
// 	"github.com/iotaledger/wasp/packages/isc"
// 	"github.com/iotaledger/wasp/packages/isc/isctest"
// 	"github.com/iotaledger/wasp/packages/testutil/testlogger"
// )

// func TestVarLocalView(t *testing.T) {
// 	log := testlogger.NewLogger(t)
// 	defer log.Shutdown()
// 	j := cmtlog.NewVarLocalView(-1, func(anchor *isc.StateAnchor) {}, log)
// 	require.Nil(t, j.Value())
// 	randAnchor := isctest.RandomStateAnchor()
// 	tipAO, ok, _ := j.AliasOutputConfirmed(&randAnchor)
// 	require.True(t, ok)
// 	require.NotNil(t, tipAO)
// 	require.NotNil(t, j.Value())
// 	require.Equal(t, tipAO, j.Value())
// }
