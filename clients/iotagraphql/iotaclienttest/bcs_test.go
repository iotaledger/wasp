package iotaclienttest

import (
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
)

func TestUnmarshalBCS(t *testing.T) {
	v := "hello"
	vEnc := bcs.MustMarshal(&v)

	var vDec string
	err := iotaclient.UnmarshalBCS(vEnc, &vDec)
	require.NoError(t, err)
	require.Equal(t, v, vDec)

	vDec = ""
	vEncWithExcess := append(vEnc, 0x1)
	err = iotaclient.UnmarshalBCS(vEncWithExcess, &vDec)
	require.Error(t, err)
}
