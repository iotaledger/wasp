package test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/webapi"
	"github.com/iotaledger/wasp/packages/webapi/models"
)

func TestMockingOfPtrStructure(t *testing.T) {
	mock := webapi.NewMocker()
	mock.LoadMockFiles()
	mockedChainInfoResponse := mock.Get(&models.PeeringNodeStatusResponse{})

	_, err := json.Marshal(mockedChainInfoResponse)
	require.NoError(t, err)
}

func TestMockingOfStructure(t *testing.T) {
	mock := webapi.NewMocker()
	mock.LoadMockFiles()
	mockedChainInfoResponse := mock.Get(models.PeeringNodeStatusResponse{})

	_, err := json.Marshal(mockedChainInfoResponse)
	require.NoError(t, err)
}

func TestMockingOfStructureArray(t *testing.T) {
	mock := webapi.NewMocker()
	mock.LoadMockFiles()
	mockedChainInfoResponse := mock.Get([]models.PeeringNodeStatusResponse{})

	_, err := json.Marshal(mockedChainInfoResponse)
	require.NoError(t, err)
}

func TestMockingOfChainInfo(t *testing.T) {
	mock := webapi.NewMocker()
	mock.LoadMockFiles()
	mockedChainInfoResponse := mock.Get([]models.ChainInfoResponse{})

	_, err := json.Marshal(mockedChainInfoResponse)
	require.NoError(t, err)
}

func TestMockingOfCommitteeInfo(t *testing.T) {
	mock := webapi.NewMocker()
	mock.LoadMockFiles()
	mockedChainInfoResponse := mock.Get(models.CommitteeInfoResponse{})

	_, err := json.Marshal(mockedChainInfoResponse)
	require.NoError(t, err)
}
