// Package interfaces defines interfaces for various webapi systems
package interfaces

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/webapi/dto"
	"github.com/iotaledger/wasp/packages/webapi/models"
)

var (
	ErrChainNotFound      = errors.New("chain not found")
	ErrCantDeleteLastUser = errors.New("you can't delete the last user")
)

type APIController interface {
	Name() string
	RegisterPublic(publicAPI echoswagger.ApiGroup, mocker Mocker)
	RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker Mocker)
}

type ChainService interface {
	ActivateChain(chainID isc.ChainID) error
	SetChainRecord(chainRecord *registry.ChainRecord) error
	DeactivateChain(chainID isc.ChainID) error
	GetAllChainIDs() ([]isc.ChainID, error)
	GetChain() (chain.Chain, error)
	GetChainInfo(blockIndexOrTrieRoot string) (*dto.ChainInfo, error)
	GetContracts(blockIndexOrTrieRoot string) ([]lo.Tuple2[*isc.Hname, *root.ContractRecord], error)
	GetEVMChainID(blockIndexOrTrieRoot string) (uint16, error)
	GetState(stateKey []byte) (state []byte, err error)
	WaitForRequestProcessed(ctx context.Context, requestID isc.RequestID, waitForL1Confirmation bool, timeout time.Duration) (*isc.Receipt, error)
	RotateTo(ctx context.Context, rotateToAddress *iotago.Address) error
}

type EVMService interface {
	HandleJSONRPC(request *http.Request, response *echo.Response) error
	HandleWebsocket(ctx context.Context, echoCtx echo.Context) error
}

type MetricsService interface {
	GetNodeMessageMetrics() *dto.NodeMessageMetrics
	GetChainMessageMetrics(chainID isc.ChainID) *dto.ChainMessageMetrics
	GetChainConsensusPipeMetrics(chainID isc.ChainID) *models.ConsensusPipeMetrics
	GetChainConsensusWorkflowMetrics(chainID isc.ChainID) *models.ConsensusWorkflowMetrics
	GetMaxChainConfirmedStateLag() uint32
}

var ErrPeerNotFound = errors.New("couldn't find peer")

type NodeService interface {
	AddAccessNode(chainID isc.ChainID, peer string) error
	DeleteAccessNode(chainID isc.ChainID, peer string) error
	NodeOwnerCertificate() []byte
	ShutdownNode()
	L1Params(context.Context) (*parameters.L1Params, error)
}

type RegistryService interface {
	GetChainRecordByChainID(chainID isc.ChainID) (*registry.ChainRecord, error)
}

type CommitteeService interface {
	GetCommitteeInfo(chainID isc.ChainID) (*dto.ChainNodeInfo, error)
	GetPublicKey() *cryptolib.PublicKey
}

type PeeringService interface {
	DistrustPeer(name string) (*dto.PeeringNodeIdentity, error)
	GetIdentity() *dto.PeeringNodeIdentity
	GetRegisteredPeers() []*dto.PeeringNodeStatus
	GetTrustedPeers() ([]*dto.PeeringNodeIdentity, error)
	IsPeerTrusted(publicKey *cryptolib.PublicKey) error
	TrustPeer(name string, pubkey *cryptolib.PublicKey, peeringURL string) (*dto.PeeringNodeIdentity, error)
}

type OffLedgerService interface {
	EnqueueOffLedgerRequest(chainID isc.ChainID, request []byte) error
	ParseRequest(payload []byte) (isc.Request, error)
}

type UserService interface {
	AddUser(username string, password string, permissions []string) error
	DeleteUser(username string) error
	GetUser(username string) (*models.User, error)
	GetUsers() []*models.User
	UpdateUserPassword(username string, password string) error
	UpdateUserPermissions(username string, permissions []string) error
}

type Mocker interface {
	Get(i any) any
}
