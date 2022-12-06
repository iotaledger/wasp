package evm

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

type jsonRPCService struct {
	chains            chains.Provider
	nodePubKey        func() *cryptolib.PublicKey
	chainServers      map[isc.ChainID]*chainServer
	chainServersMutex sync.Mutex
}

type chainServer struct {
	backend *jsonRPCWaspBackend
	rpc     *rpc.Server
}

func AddEndpoints(server echoswagger.ApiGroup, allChains chains.Provider, nodePubKey func() *cryptolib.PublicKey) {
	j := &jsonRPCService{
		chains:       allChains,
		nodePubKey:   nodePubKey,
		chainServers: make(map[isc.ChainID]*chainServer),
	}
	server.EchoGroup().Any(routes.EVMJSONRPC(":chainID"), j.handleJSONRPC)

	reqid := model.NewRequestID(isc.NewRequestID(iotago.TransactionID{}, 0))
	server.GET(routes.RequestIDByEVMTransactionHash(":chainID", ":txHash"), j.handleRequestID).
		SetSummary("Get the ISC request ID for the given Ethereum transaction hash").
		AddResponse(http.StatusOK, "Request ID", "", nil).
		AddResponse(http.StatusNotFound, "Request ID not found", reqid, nil)
}

func (j *jsonRPCService) getChainServer(c echo.Context) (*chainServer, error) {
	chainID, err := isc.ChainIDFromString(c.Param("chainID"))
	if err != nil {
		return nil, httperrors.BadRequest(fmt.Sprintf("Invalid chain ID: %+v", c.Param("chainID")))
	}

	j.chainServersMutex.Lock()
	defer j.chainServersMutex.Unlock()

	if j.chainServers[*chainID] == nil {
		chain := j.chains().Get(chainID)
		if chain == nil {
			return nil, httperrors.NotFound(fmt.Sprintf("Chain not found: %+v", c.Param("chainID")))
		}

		nodePubKey := j.nodePubKey()
		if nodePubKey == nil {
			return nil, fmt.Errorf("node is not authenticated")
		}

		backend := newWaspBackend(chain, nodePubKey, parameters.L1().BaseToken)

		var evmChainID uint16
		{
			r, err := backend.ISCCallView(backend.ISCLatestBlockIndex(), evm.Contract.Name, evm.FuncGetChainID.Name, nil)
			if err != nil {
				return nil, err
			}
			evmChainID, err = evmtypes.DecodeChainID(r.MustGet(evm.FieldResult))
			if err != nil {
				return nil, err
			}
		}

		j.chainServers[*chainID] = &chainServer{
			backend: backend,
			rpc: jsonrpc.NewServer(
				jsonrpc.NewEVMChain(backend, evmChainID),
				jsonrpc.NewAccountManager(nil),
			),
		}
	}

	return j.chainServers[*chainID], nil
}

func (j *jsonRPCService) handleJSONRPC(c echo.Context) error {
	server, err := j.getChainServer(c)
	if err != nil {
		return err
	}
	server.rpc.ServeHTTP(c.Response(), c.Request())
	return nil
}

func (j *jsonRPCService) handleRequestID(c echo.Context) error {
	server, err := j.getChainServer(c)
	if err != nil {
		return err
	}

	txHash := common.HexToHash(c.Param("txHash"))

	reqID, ok := server.backend.RequestIDByTransactionHash(txHash)
	if !ok {
		return httperrors.NotFound("not found")
	}
	return c.JSON(http.StatusOK, model.NewRequestID(reqID))
}
