package test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/webapi/v2/controllers/models"

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
