package isc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestCallArgsSerialization(t *testing.T) {
	testBuffer1 := []byte{1, 0, 7, 4}
	testBuffer2 := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	testBuffer3 := []byte{9, 1, 8, 2, 7, 3, 6, 4, 5}

	callArgs := NewCallArguments(testBuffer1, testBuffer2, testBuffer3)
	require.Equal(t, callArgs.Length(), 3)

	enc, err := bcs.Marshal(&callArgs)
	require.NoError(t, err)

	callArgsDec, err := bcs.Unmarshal[CallArguments](enc)
	require.NoError(t, err)

	require.Equal(t, len(callArgsDec), 3)

	require.Equal(t, testBuffer1, callArgsDec.MustAt(0))
	require.Equal(t, testBuffer2, callArgsDec.MustAt(1))
	require.Equal(t, testBuffer3, callArgsDec.MustAt(2))
}

func TestCallArgsJSON(t *testing.T) {
	callArgs := NewCallArguments([]byte{1, 2, 3, 4}, []byte{9, 8, 7, 6})

	jsonStr, err := json.Marshal(&callArgs)
	require.NoError(t, err)

	newCallArgs := NewCallArguments()
	err = json.Unmarshal(jsonStr, &newCallArgs)

	require.NoError(t, err)

	require.EqualValues(t, callArgs[:], newCallArgs[:])
}
