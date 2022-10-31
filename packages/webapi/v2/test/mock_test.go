package test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	"github.com/stretchr/testify/require"

	v2 "github.com/iotaledger/wasp/packages/webapi/v2"
)

func TestMockingOfPtrStructure(t *testing.T) {
	mock := v2.NewMocker()
	mock.LoadMockFiles()
	mockedChainInfoResponse := mock.Get(&models.PeeringNodeStatusResponse{})

	_, err := json.Marshal(mockedChainInfoResponse)
	require.NoError(t, err)
}

func TestMockingOfStructure(t *testing.T) {
	mock := v2.NewMocker()
	mock.LoadMockFiles()
	mockedChainInfoResponse := mock.Get(models.PeeringNodeStatusResponse{})

	_, err := json.Marshal(mockedChainInfoResponse)
	require.NoError(t, err)
}

func TestMockingOfStructureArray(t *testing.T) {
	mock := v2.NewMocker()
	mock.LoadMockFiles()
	mockedChainInfoResponse := mock.Get([]models.PeeringNodeStatusResponse{})

	_, err := json.Marshal(mockedChainInfoResponse)
	require.NoError(t, err)
}

func TestMockingOfChainInfo(t *testing.T) {
	mock := v2.NewMocker()
	mock.LoadMockFiles()
	mockedChainInfoResponse := mock.Get([]models.ChainInfoResponse{})

	_, err := json.Marshal(mockedChainInfoResponse)
	require.NoError(t, err)
	fmt.Print(mockedChainInfoResponse)
}

func TestMockingOfCommitteeInfo(t *testing.T) {
	mock := v2.NewMocker()
	mock.LoadMockFiles()
	mockedChainInfoResponse := mock.Get(models.CommitteeInfoResponse{})

	result, err := json.Marshal(mockedChainInfoResponse)
	require.NoError(t, err)
	fmt.Print(mockedChainInfoResponse)
	fmt.Print(string(result))
}
