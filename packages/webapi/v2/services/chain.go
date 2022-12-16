package services

import (
	"errors"
	"time"

	"github.com/iotaledger/hive.go/core/logger"
	chainpkg "github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/webapi/v2/corecontracts"
	"github.com/iotaledger/wasp/packages/webapi/v2/dto"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

const MaxTimeout = 30 * time.Second

type ChainService struct {
	log *logger.Logger

	governance                  *corecontracts.Governance
	chainsProvider              chains.Provider
	nodeConnectionMetrics       nodeconnmetrics.NodeConnectionMetrics
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider
	vmService                   interfaces.VMService
}

func NewChainService(log *logger.Logger, chainsProvider chains.Provider, nodeConnectionMetrics nodeconnmetrics.NodeConnectionMetrics, chainRecordRegistryProvider registry.ChainRecordRegistryProvider, vmService interfaces.VMService) interfaces.ChainService {
	return &ChainService{
		log: log,

		governance:                  corecontracts.NewGovernance(vmService),
		chainsProvider:              chainsProvider,
		nodeConnectionMetrics:       nodeConnectionMetrics,
		chainRecordRegistryProvider: chainRecordRegistryProvider,
		vmService:                   vmService,
	}
}

func (c *ChainService) ActivateChain(chainID isc.ChainID) error {
	_, err := c.chainRecordRegistryProvider.ActivateChainRecord(chainID)
	if err != nil {
		return err
	}

	c.log.Debugw("calling Chains.Activate", "chainID", chainID.String())

	err = c.chainsProvider().Activate(chainID)

	return err
}

func (c *ChainService) DeactivateChain(chainID isc.ChainID) error {
	_, err := c.chainRecordRegistryProvider.DeactivateChainRecord(chainID)
	if err != nil {
		return err
	}

	c.log.Debugw("calling Chains.Activate", "chainID", chainID.String())

	err = c.chainsProvider().Deactivate(chainID)

	return err
}

func (c *ChainService) HasChain(chainID isc.ChainID) bool {
	return c.GetChainByID(chainID) != nil
}

func (c *ChainService) GetChainByID(chainID isc.ChainID) chainpkg.Chain {
	chain := c.chainsProvider().Get(chainID)

	return chain
}

func (c *ChainService) GetEVMChainID(chainID isc.ChainID) (uint16, error) {
	ret, err := c.vmService.CallViewByChainID(chainID, evm.Contract.Hname(), evm.FuncGetChainID.Hname(), nil)
	if err != nil {
		return 0, err
	}

	return codec.DecodeUint16(ret.MustGet(evm.FieldResult))
}

func (c *ChainService) GetAllChainIDs() ([]isc.ChainID, error) {
	records, err := c.chainRecordRegistryProvider.ChainRecords()
	if err != nil {
		return nil, err
	}

	chainIDs := make([]isc.ChainID, 0, len(records))

	for _, chainRecord := range records {
		chainIDs = append(chainIDs, chainRecord.ChainID())
	}

	return chainIDs, nil
}

func (c *ChainService) GetChainInfoByChainID(chainID isc.ChainID) (*dto.ChainInfo, error) {
	governanceChainInfo, err := c.governance.GetChainInfo(chainID)
	if err != nil {
		return nil, err
	}

	chainRecord, err := c.chainRecordRegistryProvider.ChainRecord(chainID)
	if err != nil {
		return nil, err
	}

	chainInfo := dto.MapChainInfo(governanceChainInfo, chainRecord.Active)

	return chainInfo, nil
}

func (c *ChainService) GetContracts(chainID isc.ChainID) (dto.ContractsMap, error) {
	recs, err := c.vmService.CallViewByChainID(chainID, root.Contract.Hname(), root.ViewGetContractRecords.Hname(), nil)
	if err != nil {
		return nil, err
	}

	contracts, err := root.DecodeContractRegistry(collections.NewMapReadOnly(recs, root.StateVarContractRegistry))
	if err != nil {
		return nil, err
	}

	return contracts, nil
}

func (c *ChainService) GetState(chainID isc.ChainID, stateKey []byte) (state []byte, err error) {
	chain := c.chainsProvider().Get(chainID)

	latestState, err := chain.GetStateReader().LatestState()
	if err != nil {
		return nil, err
	}

	return latestState.Get(kv.Key(stateKey))
}

func (c *ChainService) WaitForRequestProcessed(chainID isc.ChainID, requestID isc.RequestID, timeout time.Duration) (*isc.Receipt, *isc.VMError, error) {
	chain := c.chainsProvider().Get(chainID)

	if chain == nil {
		return nil, nil, errors.New("chain does not exist")
	}

	receipt, vmError, err := c.vmService.GetReceipt(chainID, requestID)
	if err != nil {
		return nil, vmError, err
	}

	if receipt != nil {
		return receipt, vmError, nil
	}

	// subscribe to event
	requestProcessed := make(chan bool)
	attachID := chain.AttachToRequestProcessed(func(rid isc.RequestID) {
		if rid == requestID {
			requestProcessed <- true
		}
	})
	defer chain.DetachFromRequestProcessed(attachID)

	adjustedTimeout := timeout

	if timeout > MaxTimeout {
		adjustedTimeout = MaxTimeout
	}

	select {
	case <-requestProcessed:
		receipt, vmError, err = c.vmService.GetReceipt(chainID, requestID)
		if receipt != nil {
			return receipt, vmError, err
		}
		return nil, nil, errors.New("unexpected error, receipt not found after request was processed")
	case <-time.After(adjustedTimeout):
		// check again, in case event was triggered just before we subscribed
		receipt, vmError, err = c.vmService.GetReceipt(chainID, requestID)
		if receipt != nil {
			return receipt, vmError, err
		}
		return nil, nil, errors.New("timeout while waiting for request to be processed")
	}
}
