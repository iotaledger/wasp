// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package messages

import (
	"bytes"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/testutil/testkey"
	"github.com/stretchr/testify/require"
)

const foo = "foo"

func TestMarshalling(t *testing.T) {
	// construct a dummy offledger request
	contract := iscp.Hn("somecontract")
	entrypoint := iscp.Hn("someentrypoint")
	args := dict.Dict{foo: []byte("bar")}
	nonce := uint64(time.Now().UnixNano())
	chainID := iscp.RandomChainID()
	key, _ := testkey.GenKeyAddr()

	msg := NewOffLedgerRequestMsg(
		chainID,
		iscp.NewOffLedgerRequest(iscp.RandomChainID(), contract, entrypoint, args, nonce).WithGasBudget(1000).Sign(key),
	)

	// marshall the msg
	msgBytes := msg.Bytes()

	// unmashal the message from bytes and ensure everything checks out
	unmarshalledMsg, err := OffLedgerRequestMsgFromBytes(msgBytes)
	require.NoError(t, err)

	require.Equal(t, unmarshalledMsg.ChainID.AsAliasAddress(), msg.ChainID.AsAliasAddress())
	require.True(t, bytes.Equal(unmarshalledMsg.Req.Bytes(), msg.Req.Bytes()))
}
