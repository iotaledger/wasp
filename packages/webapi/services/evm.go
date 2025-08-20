package services

import (
	"context"
	"net/http"
	"sync"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/labstack/echo/v4"

	hivedb "github.com/iotaledger/hive.go/db"
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/packages/chainrunner"
	"github.com/iotaledger/wasp/v2/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/metrics"
	"github.com/iotaledger/wasp/v2/packages/peering"
	"github.com/iotaledger/wasp/v2/packages/publisher"
	"github.com/iotaledger/wasp/v2/packages/webapi/interfaces"
)

type chainServer struct {
	backend *WaspEVMBackend
	rpc     *rpc.Server
}

type EVMService struct {
	evmBackendMutex sync.Mutex
	evmChainServers map[isc.ChainID]*chainServer

	websocketContextMutex sync.Mutex
	websocketContext      *websocketContext

	chainRunner     *chainrunner.ChainRunner
	chainService    interfaces.ChainService
	networkProvider peering.NetworkProvider
	publisher       *publisher.Publisher
	indexDBPath     string
	metrics         *metrics.ChainMetricsProvider
	jsonrpcParams   *jsonrpc.Parameters
	log             log.Logger
}

func NewEVMService(
	chainRunner *chainrunner.ChainRunner,
	chainService interfaces.ChainService,
	networkProvider peering.NetworkProvider,
	pub *publisher.Publisher,
	indexDBPath string,
	metrics *metrics.ChainMetricsProvider,
	jsonrpcParams *jsonrpc.Parameters,
	log log.Logger,
) interfaces.EVMService {
	return &EVMService{
		chainRunner:           chainRunner,
		chainService:          chainService,
		evmChainServers:       map[isc.ChainID]*chainServer{},
		evmBackendMutex:       sync.Mutex{},
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
			e.chainRunner.IsArchiveNode(),
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

func (e *EVMService) getWebsocketContext(ctx context.Context) *websocketContext {
	e.websocketContextMutex.Lock()
	defer e.websocketContextMutex.Unlock()

	if e.websocketContext != nil {
		return e.websocketContext
	}

	e.websocketContext = newWebsocketContext(e.log, e.jsonrpcParams)
	go e.websocketContext.runCleanupTimer(ctx)

	return e.websocketContext
}

func (e *EVMService) HandleWebsocket(ctx context.Context, echoCtx echo.Context) error {
	evmServer, err := e.getEVMBackend()
	if err != nil {
		return err
	}

	wsContext := e.getWebsocketContext(ctx)
	websocketHandler(evmServer, wsContext, echoCtx.RealIP()).ServeHTTP(echoCtx.Response(), echoCtx.Request())
	return nil
}
