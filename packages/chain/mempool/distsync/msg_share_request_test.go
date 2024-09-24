// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distsync

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/sui-go/sui"
)

func TestMsgShareRequestSerialization(t *testing.T) {
	{
		req := isc.NewOffLedgerRequest(isc.RandomChainID(), isc.NewMessage(3, 14, isc.NewCallArguments()), 1337, 100).Sign(cryptolib.NewKeyPair())
		msg := &msgShareRequest{
			gpa.BasicMessage{},
			req,
			byte(rand.Intn(math.MaxUint8)),
		}

		rwutil.ReadWriteTest(t, msg, new(msgShareRequest), rwutil.SimpleEqualFun)
	}
	{
		sender := cryptolib.NewRandomAddress()
		requestRef := sui.RandomObjectRef()
		assetsBagID := sui.RandomAddress()
		request := &iscmove.RefWithObject[iscmove.Request]{
			ObjectRef: *requestRef,
			Object: &iscmove.Request{
				ID:     *requestRef.ObjectID,
				Sender: sender,
				AssetsBag: iscmove.Referent[iscmove.AssetsBagWithBalances]{
					ID: *assetsBagID,
					Value: &iscmove.AssetsBagWithBalances{
						AssetsBag: iscmove.AssetsBag{
							ID:   *assetsBagID,
							Size: 5,
						},
						Balances: iscmove.AssetsBagBalances{},
					},
				},
				Message: iscmove.Message{
					Contract: uint32(isc.Hn("target_contract")),
					Function: uint32(isc.Hn("entrypoint")),
					Args:     [][]byte{},
				},
				GasBudget: 1000,
			},
		}
		req, err := isc.OnLedgerFromRequest(request, sender)
		require.NoError(t, err)

		msg := &msgShareRequest{
			gpa.BasicMessage{},
			req,
			byte(rand.Intn(math.MaxUint8)),
		}

		rwutil.ReadWriteTest(t, msg, new(msgShareRequest), rwutil.SimpleEqualFun)
	}
}
