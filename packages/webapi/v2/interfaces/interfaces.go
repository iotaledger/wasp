package interfaces

import (
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/wasp/packages/cryptolib"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/webapi/v2/dto"
)

type APIController interface {
	Name() string
	RegisterExampleData(mocker Mocker)
	RegisterPublic(publicAPI echoswagger.ApiGroup, mocker Mocker)
	RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker Mocker)
}

type Chain interface {
	ActivateChain(chainID *isc.ChainID) error
	DeactivateChain(chainID *isc.ChainID) error
	GetAllChainIDs() ([]*isc.ChainID, error)
	GetChainByID(chainID *isc.ChainID) chain.Chain
	GetChainInfoByChainID(chainID *isc.ChainID) (*dto.ChainInfo, error)
	GetContracts(chainID *isc.ChainID) (dto.ContractsMap, error)
	GetEVMChainID(chainID *isc.ChainID) (uint16, error)
	SaveChainRecord(chainID *isc.ChainID, active bool) error
}

type Registry interface {
	GetChainRecordByChainID(chainID *isc.ChainID) (*registry.ChainRecord, error)
}

type Node interface {
	GetNodeInfo(chain chain.Chain) (*dto.ChainNodeInfo, error)
	GetPublicKey() *cryptolib.PublicKey
}

type OffLedger interface {
	ParseRequest(payload []byte) (isc.OffLedgerRequest, error)
	EnqueueOffLedgerRequest(chainID *isc.ChainID, request []byte) error
}

type VM interface {
	CallView(chain chain.Chain, contractName isc.Hname, functionName isc.Hname, params dict.Dict) (dict.Dict, error)
	CallViewByChainID(chainID *isc.ChainID, contractName isc.Hname, functionName isc.Hname, params dict.Dict) (dict.Dict, error)
}

type Mocker interface {
	AddModel(i interface{})
	GetMockedStruct(i interface{}) interface{}
}
