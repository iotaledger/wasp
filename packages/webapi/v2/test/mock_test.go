package test

import (
	"encoding/json"
	"github.com/iotaledger/wasp/packages/webapi/v2/models"
	"testing"

	"github.com/stretchr/testify/require"

	v2 "github.com/iotaledger/wasp/packages/webapi/v2"
)

func TestMockingOfPtrStructure(t *testing.T) {
	mock := v2.NewMocker()
	mock.LoadMockFiles()
	err := mock.AddModel(&models.ChainInfoResponse{})

	require.NoError(t, err)

	mockedChainInfoResponse := mock.GetMockedStruct(models.ChainInfoResponse{})

	_, err = json.Marshal(mockedChainInfoResponse)
	require.NoError(t, err)
}

func TestMockingOfStructure(t *testing.T) {
	mock := v2.NewMocker()
	mock.LoadMockFiles()
	err := mock.AddModel(models.ChainInfoResponse{})
	require.NoError(t, err)

	mockedChainInfoResponse := mock.GetMockedStruct(models.ChainInfoResponse{})

	_, err = json.Marshal(mockedChainInfoResponse)
	require.NoError(t, err)
}
