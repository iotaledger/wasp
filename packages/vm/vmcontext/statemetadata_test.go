package vmcontext_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
)

func TestStateMetadataSerialization(t *testing.T) {
	s := &vmcontext.StateMetadata{
		L1Commitment: state.PseudoRandL1Commitment(),
		GasFeePolicy: &gas.GasFeePolicy{
			GasFeeTokenID:       [38]byte{},
			GasFeeTokenDecimals: 0,
			GasPerToken: util.Ratio32{
				A: 1,
				B: 2,
			},
			EVMGasRatio: util.Ratio32{
				A: 3,
				B: 4,
			},
			ValidatorFeeShare: 5,
		},
		SchemaVersion:  6,
		CustomMetadata: "foo",
	}
	data := s.Bytes()
	s2, err := vmcontext.StateMetadataFromBytes(data)
	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(s, s2))
}
