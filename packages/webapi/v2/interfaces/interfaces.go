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
	RegisterPublic(publicAPI echoswagger.ApiGroup, mocker Mocker)
	RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker Mocker)
}

type ChainService interface {
	ActivateChain(chainID *isc.ChainID) error
	DeactivateChain(chainID *isc.ChainID) error
	GetAllChainIDs() ([]*isc.ChainID, error)
	GetChainByID(chainID *isc.ChainID) chain.Chain
	GetChainInfoByChainID(chainID *isc.ChainID) (*dto.ChainInfo, error)
	GetContracts(chainID *isc.ChainID) (dto.ContractsMap, error)
	GetEVMChainID(chainID *isc.ChainID) (uint16, error)
	SaveChainRecord(chainID *isc.ChainID, active bool) error
}

type MetricsService interface {
	GetAllChainsMetrics() *dto.ChainMetricsReport
	GetChainConsensusPipeMetrics(chainID *isc.ChainID) *dto.ConsensusPipeMetrics
	GetChainConsensusWorkflowMetrics(chainID *isc.ChainID) *dto.ConsensusWorkflowMetrics
	GetChainMetrics(chainID *isc.ChainID) *dto.ChainMetricsReport
}

type RegistryService interface {
	GetChainRecordByChainID(chainID *isc.ChainID) (*registry.ChainRecord, error)
}

type CommitteeService interface {
	GetCommitteeInfo(chain chain.Chain) (*dto.ChainNodeInfo, error)
	GetPublicKey() *cryptolib.PublicKey
}

type PeeringService interface {
	DistrustPeer(publicKey *cryptolib.PublicKey) (*dto.PeeringNodeIdentity, error)
	GetIdentity() *dto.PeeringNodeIdentity
	GetRegisteredPeers() *[]dto.PeeringNodeStatus
	GetTrustedPeers() (*[]dto.PeeringNodeIdentity, error)
	IsPeerTrusted(publicKey *cryptolib.PublicKey) error
	TrustPeer(peer *cryptolib.PublicKey, netID string) (*dto.PeeringNodeIdentity, error)
}

type OffLedgerService interface {
	EnqueueOffLedgerRequest(chainID *isc.ChainID, request []byte) error
	ParseRequest(payload []byte) (isc.OffLedgerRequest, error)
}

type VMService interface {
	CallView(chain chain.Chain, contractName isc.Hname, functionName isc.Hname, params dict.Dict) (dict.Dict, error)
	CallViewByChainID(chainID *isc.ChainID, contractName isc.Hname, functionName isc.Hname, params dict.Dict) (dict.Dict, error)
}

type Mocker interface {
	Get(i interface{}) interface{}
}
