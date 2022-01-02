// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

//nolint:revive
const (
	TYPE_ADDRESS    int32 = 1
	TYPE_AGENT_ID   int32 = 2
	TYPE_BOOL       int32 = 3
	TYPE_BYTES      int32 = 4
	TYPE_CHAIN_ID   int32 = 5
	TYPE_COLOR      int32 = 6
	TYPE_HASH       int32 = 7
	TYPE_HNAME      int32 = 8
	TYPE_INT8       int32 = 9
	TYPE_INT16      int32 = 10
	TYPE_INT32      int32 = 11
	TYPE_INT64      int32 = 12
	TYPE_MAP        int32 = 13
	TYPE_REQUEST_ID int32 = 14
	TYPE_STRING     int32 = 15
)

type (
	Address   string
	AgentID   string
	ChainID   string
	Color     string
	Hash      string
	Hname     uint32
	RequestID string
)

var TypeSizes = [...]uint8{0, 33, 37, 1, 0, 33, 32, 32, 4, 1, 2, 4, 8, 0, 34, 0}
