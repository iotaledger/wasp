package services

import (
	"context"
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
	"github.com/iotaledger/wasp/packages/webapi/corecontracts"
	"github.com/iotaledger/wasp/packages/webapi/dto"
	"github.com/iotaledger/wasp/packages/webapi/interfaces"
)

type ChainService struct {
	log                         *logger.Logger
	governance                  *corecontracts.Governance
	chainsProvider              chains.Provider
	nodeConnectionMetrics       nodeconnmetrics.NodeConnectionMetrics
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider
	vmService                   interfaces.VMService
}

func NewChainService(logger *logger.Logger, chainsProvider chains.Provider, nodeConnectionMetrics nodeconnmetrics.NodeConnectionMetrics, chainRecordRegistryProvider registry.ChainRecordRegistryProvider, vmService interfaces.VMService) interfaces.ChainService {
	return &ChainService{
		log:                         logger,
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

	return c.chainsProvider().Activate(chainID)
}

func (c *ChainService) DeactivateChain(chainID isc.ChainID) error {
	_, err := c.chainRecordRegistryProvider.DeactivateChainRecord(chainID)
	if err != nil {
		return err
	}

	return c.chainsProvider().Deactivate(chainID)
}

func (c *ChainService) SetChainRecord(chainRecord *registry.ChainRecord) error {
	storedChainRec, err := c.chainRecordRegistryProvider.ChainRecord(chainRecord.ChainID())
	if err != nil {
		return err
	}

	c.log.Infof("StoredChainRec %v %v", storedChainRec, err)

	if storedChainRec != nil {
		_, err = c.chainRecordRegistryProvider.UpdateChainRecord(
			chainRecord.ChainID(),
			func(rec *registry.ChainRecord) bool {
				rec.AccessNodes = chainRecord.AccessNodes
				rec.Active = chainRecord.Active
				return true
			},
		)
		c.log.Infof("UpdatechainRec %v %v", chainRecord, err)

		if err != nil {
			return err
		}
	} else {
		if err := c.chainRecordRegistryProvider.AddChainRecord(chainRecord); err != nil {
			c.log.Infof("AddChainRec %v %v", chainRecord, err)

			return err
		}
	}

	// Activate/deactivate the chain accordingly.
	c.log.Infof("Chainrecord active %v", chainRecord.Active)

	if chainRecord.Active {
		if err := c.chainsProvider().Activate(chainRecord.ChainID()); err != nil {
			return err
		}
	} else if storedChainRec != nil {
		if err := c.chainsProvider().Deactivate(chainRecord.ChainID()); err != nil {
			return err
		}
	}

	return nil
}

func (c *ChainService) HasChain(chainID isc.ChainID) bool {
	return c.GetChainByID(chainID) != nil
}

func (c *ChainService) GetChainByID(chainID isc.ChainID) chainpkg.Chain {
	return c.chainsProvider().Get(chainID)
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
	chainRecord, err := c.chainRecordRegistryProvider.ChainRecord(chainID)
	if err != nil {
		return nil, err
	}

	governanceChainInfo, err := c.governance.GetChainInfo(chainID)
	if err != nil {
		if chainRecord != nil && errors.Is(err, interfaces.ErrChainNotFound) {
			return &dto.ChainInfo{ChainID: chainID, IsActive: false}, nil
		}

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
	ch := c.chainsProvider().Get(chainID)

	latestState, err := ch.LatestState(chainpkg.ActiveOrCommittedState)
	if err != nil {
		return nil, err
	}

	return latestState.Get(kv.Key(stateKey))
}

func (c *ChainService) WaitForRequestProcessed(ctx context.Context, chainID isc.ChainID, requestID isc.RequestID, timeout time.Duration) (*isc.Receipt, *isc.VMError, error) {
	chain := c.chainsProvider().Get(chainID)

	if chain == nil {
		return nil, nil, errors.New("chain does not exist")
	}

	receipt, vmError, _ := c.vmService.GetReceipt(chainID, requestID)
	if receipt != nil {
		return receipt, vmError, nil
	}

	timeoutCtx, cancelCtx := context.WithTimeout(ctx, timeout)
	defer cancelCtx()
	receiptResponse := <-chain.AwaitRequestProcessed(timeoutCtx, requestID, true)

	// If receipt is available, return it
	if receiptResponse != nil {
		return c.vmService.ParseReceipt(chain, receiptResponse)
	}

	// Otherwise, poll it again one last time before failing.
	receipt, vmError, err := c.vmService.GetReceipt(chainID, requestID)
	if receipt != nil {
		return receipt, vmError, err
	}

	return nil, nil, errors.New("timeout while waiting for request to be processed")
}
