package isc

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/buffer"
)

func TestCallArgsSerialization(t *testing.T) {
	testBuffer1 := []byte{1, 0, 7, 4}
	testBuffer2 := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	testBuffer3 := []byte{9, 1, 8, 2, 7, 3, 6, 4, 5}

	callArgs := NewCallArguments(testBuffer1, testBuffer2, testBuffer3)
	require.Equal(t, callArgs.Length(), 3)

	var buf buffer.Buffer
	err := callArgs.Write(&buf)
	require.NoError(t, err)

	callArgs2 := NewCallArguments()
	reader := bytes.NewReader(buf.Bytes())
	callArgs2.Read(reader)

	require.Equal(t, len(callArgs2), 3)

	require.Equal(t, testBuffer1, callArgs2.MustAt(0))
	require.Equal(t, testBuffer2, callArgs2.MustAt(1))
	require.Equal(t, testBuffer3, callArgs2.MustAt(2))
}
