// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package committeelog_test

// TODO: Re-enable this test.

// import (
// 	"testing"

// 	"github.com/stretchr/testify/require"

// 	"github.com/iotaledger/wasp/v2/packages/chain/committeelog"
// 	"github.com/iotaledger/wasp/v2/packages/isc"
// 	"github.com/iotaledger/wasp/v2/packages/isc/isctest"
// 	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
// )

// func TestLocalViewVariable(t *testing.T) {
// 	log := testlogger.NewLogger(t)
// 	defer log.Shutdown()
// 	j := committeelog.NewLocalViewVariable(-1, func(anchor *isc.StateAnchor) {}, log)
// 	require.Nil(t, j.Value())
// 	randAnchor := isctest.RandomStateAnchor()
// 	tipAnchor, ok, _ := j.AnchorConfirmed(&randAnchor)
// 	require.True(t, ok)
// 	require.NotNil(t, tipAnchor)
// 	require.NotNil(t, j.Value())
// 	require.Equal(t, tipAnchor, j.Value())
// }
