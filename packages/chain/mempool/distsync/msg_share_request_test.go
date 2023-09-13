// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distsync

import (
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestMsgShareRequestSerialization(t *testing.T) {
	{
		req := isc.NewOffLedgerRequest(isc.RandomChainID(), 3, 14, dict.New(), 1337, 100).Sign(cryptolib.NewKeyPair())
		msg := &msgShareRequest{
			gpa.BasicMessage{},
			req,
			byte(rand.Intn(math.MaxUint8)),
		}

		rwutil.ReadWriteTest(t, msg, new(msgShareRequest))
	}
	{
		sender := tpkg.RandAliasAddress()
		requestMetadata := &isc.RequestMetadata{
			SenderContract: isc.ContractIdentityFromHname(isc.Hn("sender_contract")),
			TargetContract: isc.Hn("target_contract"),
			EntryPoint:     isc.Hn("entrypoint"),
			Allowance:      isc.NewAssetsBaseTokens(1),
			GasBudget:      1000,
		}
		basicOutput := &iotago.BasicOutput{
			Amount: 123,
			NativeTokens: iotago.NativeTokens{
				&iotago.NativeToken{
					ID:     [iotago.NativeTokenIDLength]byte{1},
					Amount: big.NewInt(100),
				},
			},
			Features: iotago.Features{
				&iotago.SenderFeature{Address: sender},
				&iotago.MetadataFeature{Data: requestMetadata.Bytes()},
			},
			Conditions: iotago.UnlockConditions{
				&iotago.AddressUnlockCondition{Address: sender},
			},
		}
		req, err := isc.OnLedgerFromUTXO(basicOutput, iotago.OutputID{})
		require.NoError(t, err)

		msg := &msgShareRequest{
			gpa.BasicMessage{},
			req,
			byte(rand.Intn(math.MaxUint8)),
		}

		rwutil.ReadWriteTest(t, msg, new(msgShareRequest))
	}
}
