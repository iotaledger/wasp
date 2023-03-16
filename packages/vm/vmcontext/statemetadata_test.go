package vmcontext_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func TestStateMetadataSerialization(t *testing.T) {
	s := &transaction.StateMetadata{
		L1Commitment: state.PseudoRandL1Commitment(),
		GasFeePolicy: &gas.FeePolicy{
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
		CustomMetadata: []byte("foo"),
	}
	data := s.Bytes()
	s2, err := transaction.StateMetadataFromBytes(data)
	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(s, s2))
}
