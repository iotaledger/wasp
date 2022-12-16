package services

import (
	"errors"
	"net/http"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/labstack/echo/v4"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type chainServer struct {
	backend *jsonrpc.WaspEVMBackend
	rpc     *rpc.Server
}

type EVMService struct {
	log *logger.Logger

	evmBackendMutex sync.Mutex
	evmChainServers map[isc.ChainID]*chainServer

	chainService    interfaces.ChainService
	networkProvider peering.NetworkProvider
}

func NewEVMService(log *logger.Logger, chainService interfaces.ChainService, networkProvider peering.NetworkProvider) interfaces.EVMService {
	return &EVMService{
		log: log,

		chainService:    chainService,
		networkProvider: networkProvider,
	}
}

func (e *EVMService) getEVMBackend(chainID isc.ChainID) (*chainServer, error) {
	e.evmBackendMutex.Lock()
	defer e.evmBackendMutex.Unlock()

	if e.evmChainServers[chainID] != nil {
		return e.evmChainServers[chainID], nil
	}

	chain := e.chainService.GetChainByID(chainID)

	if chain == nil {
		return nil, errors.New("chain is invalid")
	}

	evmChainID, err := e.chainService.GetEVMChainID(chainID)
	if err != nil {
		return nil, err
	}

	nodePubKey := e.networkProvider.Self().PubKey()
	backend := jsonrpc.NewWaspEVMBackend(chain, nodePubKey, parameters.L1().BaseToken)

	e.evmChainServers[chainID] = &chainServer{
		backend: backend,
		rpc: jsonrpc.NewServer(
			jsonrpc.NewEVMChain(backend, evmChainID),
			jsonrpc.NewAccountManager(nil),
		),
	}

	return e.evmChainServers[chainID], nil
}

func (e *EVMService) HandleJSONRPC(chainID isc.ChainID, request *http.Request, response *echo.Response) error {
	evmServer, err := e.getEVMBackend(chainID)
	if err != nil {
		return err
	}

	evmServer.rpc.ServeHTTP(response, request)

	return nil
}

func (e *EVMService) GetRequestID(chainID isc.ChainID, hash string) (*isc.RequestID, error) {
	evmServer, err := e.getEVMBackend(chainID)
	if err != nil {
		return nil, err
	}

	txHash := common.HexToHash(hash)
	reqID, ok := evmServer.backend.RequestIDByTransactionHash(txHash)

	if !ok {
		return nil, errors.New("request id not found")
	}

	return &reqID, nil
}
