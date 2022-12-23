package util_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/util"
)

func TestReadIntsAsBits(t *testing.T) {
	dataToSend := []int{0, 10, 17}
	w := bytes.NewBuffer([]byte{})
	require.NoError(t, util.WriteIntsAsBits(w, dataToSend))
	r := bytes.NewReader(w.Bytes())
	ints, err := util.ReadIntsAsBits(r)
	require.NoError(t, err)
	require.Equal(t, ints, dataToSend)
}
