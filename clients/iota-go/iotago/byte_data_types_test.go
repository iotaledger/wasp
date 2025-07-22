package iotago_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

func TestSerialization(t *testing.T) {
	hexStr := "0x12333aabcc"
	targetBytes := []byte{5, 0x12, 0x33, 0x3a, 0xab, 0xcc}
	hexdata, err := iotago.NewHexData(hexStr)
	require.Nil(t, err)
	require.Equal(t, hexStr, hexdata.String())

	b64data := iotago.Bytes(hexdata.Data()).GetBase64Data()
	b64Bcs, err := bcs.Marshal(&b64data)
	require.NoError(t, err)
	require.Equal(t, "EjM6q8w=", b64data.String())
	require.Equal(t, targetBytes, b64Bcs)
}

func TestHexdataJson(t *testing.T) {
	hexdata, err := iotago.NewHexData("0x12333aabcc")
	require.Nil(t, err)

	dataJson, err := json.Marshal(hexdata)
	require.Nil(t, err)

	hexdata2 := iotago.HexData{}
	err = json.Unmarshal(dataJson, &hexdata2)
	require.Nil(t, err)
	require.Equal(t, hexdata.Data(), hexdata2.Data())

	base64data := iotago.Bytes(hexdata.Data()).GetBase64Data()
	dataJsonb, err := json.Marshal(base64data)
	require.Nil(t, err)

	base64data2 := iotago.Base64Data{}
	err = json.Unmarshal(dataJsonb, &base64data2)
	require.Nil(t, err)
	require.Equal(t, base64data.Data(), base64data2.Data())
	require.Equal(t, hexdata.Data(), base64data2.Data())
}
