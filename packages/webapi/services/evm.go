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
	indexDBPath     string
	metrics         *metrics.ChainMetricsProvider
	jsonrpcParams   *jsonrpc.Parameters
	log             log.Logger
}

func NewEVMService(
	chainsProvider chains.Provider,
	chainService interfaces.ChainService,
	networkProvider peering.NetworkProvider,
	pub *publisher.Publisher,
	indexDBPath string,
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
		indexDBPath:           indexDBPath,
		metrics:               metrics,
		jsonrpcParams:         jsonrpcParams,
		log:                   log,
	}
}

func (e *EVMService) getEVMBackend() (*chainServer, error) {
	e.evmBackendMutex.Lock()
	defer e.evmBackendMutex.Unlock()

	ch, err := e.chainService.GetChain()
	if err != nil {
		return nil, err
	}

	if e.evmChainServers[ch.ID()] != nil {
		return e.evmChainServers[ch.ID()], nil
	}

	chain, err := e.chainService.GetChain()
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
			e.indexDBPath,
			e.log.NewChildLogger("EVMChain"),
		),
		jsonrpc.NewAccountManager(nil),
		e.metrics.GetChainMetrics(ch.ID()).WebAPI,
		e.jsonrpcParams,
	)
	if err != nil {
		return nil, err
	}

	e.evmChainServers[ch.ID()] = &chainServer{
		backend: backend,
		rpc:     srv,
	}

	return e.evmChainServers[ch.ID()], nil
}

func (e *EVMService) HandleJSONRPC(request *http.Request, response *echo.Response) error {
	evmServer, err := e.getEVMBackend()
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

func (e *EVMService) HandleWebsocket(ctx context.Context, echoCtx echo.Context) error {
	evmServer, err := e.getEVMBackend()
	if err != nil {
		return err
	}

	ch, err := e.chainService.GetChain()
	if err != nil {
		return err
	}

	wsContext := e.getWebsocketContext(ctx, ch.ID())
	websocketHandler(evmServer, wsContext, echoCtx.RealIP()).ServeHTTP(echoCtx.Response(), echoCtx.Request())
	return nil
}
