package chain

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	models2 "github.com/iotaledger/wasp/packages/webapi/controllers/chain/models"
	"github.com/iotaledger/wasp/packages/webapi/controllers/controllerutils"
	"github.com/iotaledger/wasp/packages/webapi/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/models"
	"github.com/iotaledger/wasp/packages/webapi/params"
	"github.com/iotaledger/wasp/packages/webapi/services"
)

func (c *Controller) getCommitteeInfo(e echo.Context) error {
	controllerutils.SetOperation(e, "get_committee_info")

	chain, err := c.chainService.GetChainInfo(e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return apierrors.ChainNotFoundError()
	}

	chainNodeInfo, err := c.committeeService.GetCommitteeInfo(chain.ChainID)
	if err != nil {
		if errors.Is(err, services.ErrNotInCommittee) {
			return e.JSON(http.StatusOK, models.CommitteeInfoResponse{})
		}
		return err
	}

	chainInfo := models.CommitteeInfoResponse{
		ChainID:        chain.ChainID.String(),
		Active:         chain.IsActive,
		StateAddress:   chainNodeInfo.Address.String(),
		CommitteeNodes: models.MapCommitteeNodes(chainNodeInfo.CommitteeNodes),
		AccessNodes:    models.MapCommitteeNodes(chainNodeInfo.AccessNodes),
		CandidateNodes: models.MapCommitteeNodes(chainNodeInfo.CandidateNodes),
	}

	return e.JSON(http.StatusOK, chainInfo)
}

func (c *Controller) getChainInfo(e echo.Context) error {
	controllerutils.SetOperation(e, "get_chain_info")
	chainInfo, err := c.chainService.GetChainInfo(e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if errors.Is(err, interfaces.ErrChainNotFound) {
		return e.NoContent(http.StatusNotFound)
	} else if err != nil {
		return err
	}

	evmChainID := uint16(0)
	if chainInfo.IsActive {
		evmChainID, err = c.chainService.GetEVMChainID(e.QueryParam(params.ParamBlockIndexOrTrieRoot))
		if err != nil {
			return err
		}
	}

	chainInfoResponse := models.MapChainInfoResponse(chainInfo, evmChainID)

	return e.JSON(http.StatusOK, chainInfoResponse)
}

func (c *Controller) getState(e echo.Context) error {
	controllerutils.SetOperation(e, "get_state")

	stateKey, err := cryptolib.DecodeHex(e.Param(params.ParamStateKey))
	if err != nil {
		return apierrors.InvalidPropertyError(params.ParamStateKey, err)
	}

	state, err := c.chainService.GetState(stateKey)
	if err != nil {
		panic(err)
	}

	response := models.StateResponse{
		State: hexutil.Encode(state),
	}

	return e.JSON(http.StatusOK, response)
}

func (c *Controller) dumpAccounts(e echo.Context) error {
	ch := lo.Must(c.chainService.GetChain())

	blockIndex := e.QueryParam(params.ParamBlockIndexOrTrieRoot)
	var chainState state.State

	if blockIndex == "" {
		chainState = lo.Must(ch.LatestState(chain.ActiveOrCommittedState))
	} else {
		var blockIndexNum uint64
		blockIndexNum, err := strconv.ParseUint(blockIndex, 10, 64)
		if err != nil {
			return apierrors.InvalidPropertyError(params.ParamBlockIndexOrTrieRoot, err)
		}

		chainState, err = ch.Store().StateByIndex(uint32(blockIndexNum))
		if err != nil {
			return apierrors.InvalidPropertyError(params.ParamBlockIndexOrTrieRoot, err)
		}
	}

	res := &models2.DumpAccountsResponse{
		StateIndex: chainState.BlockIndex(),
		Accounts:   make(map[string]models2.AccountBalance),
	}

	sa := accounts.NewStateReaderFromChainState(allmigrations.DefaultScheme.LatestSchemaVersion(), chainState)

	sa.AllAccountsAsDict().ForEach(func(key kv.Key, value []byte) bool {
		agentID := lo.Must(accounts.AgentIDFromKey(key))
		accountAssets := sa.GetAccountFungibleTokens(agentID)
		accountObjects := sa.GetAccountObjects(agentID)

		res.Accounts[agentID.String()] = models2.AccountBalance{
			Coins:   accountAssets,
			Objects: accountObjects,
		}

		return true
	})

	return e.JSON(http.StatusOK, res)
}
