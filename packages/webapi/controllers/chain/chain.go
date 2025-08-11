package chain

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/packages/chain"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/v2/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/v2/packages/webapi/controllers/controllerutils"
	"github.com/iotaledger/wasp/v2/packages/webapi/interfaces"
	"github.com/iotaledger/wasp/v2/packages/webapi/models"
	"github.com/iotaledger/wasp/v2/packages/webapi/params"
	"github.com/iotaledger/wasp/v2/packages/webapi/services"
)

func (c *Controller) getCommitteeInfo(e echo.Context) error {
	controllerutils.SetOperation(e, "get_committee_info")

	chain, err := c.chainService.GetChainInfo(e.QueryParam(params.ParamBlockIndexOrTrieRoot))
	if err != nil {
		return apierrors.ChainNotFoundError()
	}

	chainNodeInfo, err := c.committeeService.GetCommitteeInfo()
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

var dumpAccountsMutex = sync.Mutex{}

func (c *Controller) dumpAccounts(e echo.Context) error {
	ch := lo.Must(c.chainService.GetChain())

	if !dumpAccountsMutex.TryLock() {
		return e.String(http.StatusLocked, "account dump in progress")
	}

	go func() {
		defer dumpAccountsMutex.Unlock()
		chainState := lo.Must(ch.LatestState(chain.ActiveOrCommittedState))
		blockIndex := chainState.BlockIndex()
		stateRoot := chainState.TrieRoot()
		filename := fmt.Sprintf("block_%d_stateroot_%s.json", blockIndex, stateRoot.String())

		err := os.MkdirAll(filepath.Join(c.accountDumpsPath, ch.ID().String()), os.ModePerm)
		if err != nil {
			c.log.LogErrorf("dumpAccounts - Creating dir failed: %s", err.Error())
			return
		}
		f, err := os.Create(filepath.Join(c.accountDumpsPath, ch.ID().String(), filename))
		if err != nil {
			c.log.LogErrorf("dumpAccounts - Creating account dump file failed: %s", err.Error())
			return
		}
		_, err = f.WriteString("{")
		if err != nil {
			c.log.LogErrorf("dumpAccounts - writing to account dump file failed: %s", err.Error())
			return
		}
		sa := accounts.NewStateReaderFromChainState(allmigrations.DefaultScheme.LatestSchemaVersion(), chainState)

		// because we don't know when the last account will be, we save each account string and write it in the next iteration
		// this way we can remove the trailing comma, thus getting a valid JSON
		prevString := ""

		sa.AllAccountsAsDict().ForEach(func(key kv.Key, value []byte) bool {
			if prevString != "" {
				_, err2 := f.WriteString(prevString)
				if err2 != nil {
					c.log.LogErrorf("dumpAccounts - writing to account dump file failed: %s", err2.Error())
					return false
				}
			}
			agentID := lo.Must(accounts.AgentIDFromKey(key))
			accountAssets := sa.GetAccountFungibleTokens(agentID)
			assetsJSON, err2 := json.Marshal(models.ToCoinBalancesJSON(accountAssets))
			if err2 != nil {
				c.log.LogErrorf("dumpAccounts - generating JSON for account %s assets failed%s", agentID.String(), err2.Error())
				return false
			}
			prevString = fmt.Sprintf("%q:%s,", agentID.String(), string(assetsJSON))
			return true
		})
		// delete last ',' for a valid json
		prevString = prevString[:len(prevString)-1]
		_, err = fmt.Fprintf(f, "%s}\n", prevString)
		if err != nil {
			c.log.LogErrorf("dumpAccounts - writing to account dump file failed: %s", err.Error())
		}
	}()

	return e.NoContent(http.StatusAccepted)
}
