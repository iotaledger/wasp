// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/aaa2/cmtLog"
	"github.com/iotaledger/wasp/packages/isc"
)

func TestLocalView(t *testing.T) {
	j := cmtLog.NewLocalView()
	require.Nil(t, j.GetBaseAliasOutput())
	j.AliasOutputConfirmed(isc.NewAliasOutputWithID(&iotago.AliasOutput{}, &iotago.UTXOInput{}))
	require.NotNil(t, j.GetBaseAliasOutput())
}
