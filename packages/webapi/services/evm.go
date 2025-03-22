package services

import (
	"context"
	"net/http"
	"sync"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/labstack/echo/v4"

	hivedb "github.com/iotaledger/hive.go/db"
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/webapi/interfaces"
)

type chainServer struct {
	backend *WaspEVMBackend
	rpc     *rpc.Server
}

type EVMService struct {
	evmBackendMutex sync.Mutex
	evmChainServers map[isc.ChainID]*chainServer

	websocketContextMutex sync.Mutex
	websocketContexts     map[isc.ChainID]*websocketContext

	chainsProvider  chains.Provider
	chainService    interfaces.ChainService
	networkProvider peering.NetworkProvider
	publisher       *publisher.Publisher
	indexDbPath     string
	metrics         *metrics.ChainMetricsProvider
	jsonrpcParams   *jsonrpc.Parameters
	log             log.Logger
}

func NewEVMService(
	chainsProvider chains.Provider,
	chainService interfaces.ChainService,
	networkProvider peering.NetworkProvider,
	pub *publisher.Publisher,
	indexDbPath string,
	metrics *metrics.ChainMetricsProvider,
	jsonrpcParams *jsonrpc.Parameters,
	log log.Logger,
) interfaces.EVMService {
	return &EVMService{
		chainsProvider:        chainsProvider,
		chainService:          chainService,
		evmChainServers:       map[isc.ChainID]*chainServer{},
		evmBackendMutex:       sync.Mutex{},
		websocketContexts:     map[isc.ChainID]*websocketContext{},
		websocketContextMutex: sync.Mutex{},
		networkProvider:       networkProvider,
		publisher:             pub,
		indexDbPath:           indexDbPath,
		metrics:               metrics,
		jsonrpcParams:         jsonrpcParams,
		log:                   log,
	}
}

func (e *EVMService) getEVMBackend(chainID isc.ChainID) (*chainServer, error) {
	e.evmBackendMutex.Lock()
	defer e.evmBackendMutex.Unlock()

	if e.evmChainServers[chainID] != nil {
		return e.evmChainServers[chainID], nil
	}

	chain, err := e.chainService.GetChainByID(chainID)
	if err != nil {
		return nil, err
	}

	nodePubKey := e.networkProvider.Self().PubKey()
	backend := NewWaspEVMBackend(chain, nodePubKey)

	srv, err := jsonrpc.NewServer(
		jsonrpc.NewEVMChain(
			backend,
			e.publisher,
			e.chainsProvider().IsArchiveNode(),
			hivedb.EngineRocksDB,
			e.indexDbPath,
			e.log.NewChildLogger("EVMChain"),
		),
		jsonrpc.NewAccountManager(nil),
		e.metrics.GetChainMetrics(chainID).WebAPI,
		e.jsonrpcParams,
	)
	if err != nil {
		return nil, err
	}

	e.evmChainServers[chainID] = &chainServer{
		backend: backend,
		rpc:     srv,
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

func (e *EVMService) getWebsocketContext(ctx context.Context, chainID isc.ChainID) *websocketContext {
	e.websocketContextMutex.Lock()
	defer e.websocketContextMutex.Unlock()

	if e.websocketContexts[chainID] != nil {
		return e.websocketContexts[chainID]
	}

	e.websocketContexts[chainID] = newWebsocketContext(e.log, e.jsonrpcParams)
	go e.websocketContexts[chainID].runCleanupTimer(ctx)

	return e.websocketContexts[chainID]
}

func (e *EVMService) HandleWebsocket(ctx context.Context, chainID isc.ChainID, echoCtx echo.Context) error {
	evmServer, err := e.getEVMBackend(chainID)
	if err != nil {
		return err
	}

	wsContext := e.getWebsocketContext(ctx, chainID)
	websocketHandler(evmServer, wsContext, echoCtx.RealIP()).ServeHTTP(echoCtx.Response(), echoCtx.Request())
	return nil
}
