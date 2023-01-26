package test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	v2 "github.com/iotaledger/wasp/packages/webapi"
	models2 "github.com/iotaledger/wasp/packages/webapi/models"
)

func TestMockingOfPtrStructure(t *testing.T) {
	mock := v2.NewMocker()
	mock.LoadMockFiles()
	mockedChainInfoResponse := mock.Get(&models2.PeeringNodeStatusResponse{})

	_, err := json.Marshal(mockedChainInfoResponse)
	require.NoError(t, err)
}

func TestMockingOfStructure(t *testing.T) {
	mock := v2.NewMocker()
	mock.LoadMockFiles()
	mockedChainInfoResponse := mock.Get(models2.PeeringNodeStatusResponse{})

	_, err := json.Marshal(mockedChainInfoResponse)
	require.NoError(t, err)
}

func TestMockingOfStructureArray(t *testing.T) {
	mock := v2.NewMocker()
	mock.LoadMockFiles()
	mockedChainInfoResponse := mock.Get([]models2.PeeringNodeStatusResponse{})

	_, err := json.Marshal(mockedChainInfoResponse)
	require.NoError(t, err)
}

func TestMockingOfChainInfo(t *testing.T) {
	mock := v2.NewMocker()
	mock.LoadMockFiles()
	mockedChainInfoResponse := mock.Get([]models2.ChainInfoResponse{})

	_, err := json.Marshal(mockedChainInfoResponse)
	require.NoError(t, err)
}

func TestMockingOfCommitteeInfo(t *testing.T) {
	mock := v2.NewMocker()
	mock.LoadMockFiles()
	mockedChainInfoResponse := mock.Get(models2.CommitteeInfoResponse{})

	_, err := json.Marshal(mockedChainInfoResponse)
	require.NoError(t, err)
}
