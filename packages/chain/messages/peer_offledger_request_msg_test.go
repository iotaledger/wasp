// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package messages

import (
	"bytes"
	"testing"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/packages/iscp/requestargs"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/stretchr/testify/require"
)

const foo = "foo"

func TestMarshalling(t *testing.T) {
	// construct a dummy offledger request
	contract := iscp.Hn("somecontract")
	entrypoint := iscp.Hn("someentrypoint")
	args := requestargs.New(
		dict.Dict{foo: []byte("bar")},
	)

	msg := &OffLedgerRequestMsg{
		ChainID: iscp.RandomChainID(),
		Req:     request.NewOffLedger(iscp.RandomChainID(), contract, entrypoint, args),
	}

	// marshall the msg
	msgBytes := msg.Bytes()

	// unmashal the message from bytes and ensure everything checks out
	unmarshalledMsg, err := NewOffLedgerRequestMsg(msgBytes)
	require.NoError(t, err)

	require.True(t, unmarshalledMsg.ChainID.AliasAddress.Equals(msg.ChainID.AliasAddress))
	require.True(t, bytes.Equal(unmarshalledMsg.Req.Bytes(), msg.Req.Bytes()))
}
