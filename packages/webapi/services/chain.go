package services

import (
	"context"
	"errors"
	"time"

	"github.com/iotaledger/hive.go/logger"
	chainpkg "github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/common"
	"github.com/iotaledger/wasp/packages/webapi/corecontracts"
	"github.com/iotaledger/wasp/packages/webapi/dto"
	"github.com/iotaledger/wasp/packages/webapi/interfaces"
)

type ChainService struct {
	log                         *logger.Logger
	chainsProvider              chains.Provider
	chainMetricsProvider        *metrics.ChainMetricsProvider
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider
}

func NewChainService(
	logger *logger.Logger,
	chainsProvider chains.Provider,
	chainMetricsProvider *metrics.ChainMetricsProvider,
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider,
) interfaces.ChainService {
	return &ChainService{
		log:                         logger,
		chainsProvider:              chainsProvider,
		chainMetricsProvider:        chainMetricsProvider,
		chainRecordRegistryProvider: chainRecordRegistryProvider,
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
	storedChainRec, err := c.chainRecordRegistryProvider.ChainRecord(chainID)
	if err != nil {
		c.log.Infof("hasChain err:[%v]", err)
		return false
	}

	return storedChainRec != nil
}

func (c *ChainService) GetChainByID(chainID isc.ChainID) (chainpkg.Chain, error) {
	return c.chainsProvider().Get(chainID)
}

func (c *ChainService) GetEVMChainID(chainID isc.ChainID, blockIndexOrTrieRoot string) (uint16, error) {
	ch, err := c.GetChainByID(chainID)
	if err != nil {
		return 0, err
	}
	ret, err := common.CallView(ch, evm.Contract.Hname(), evm.FuncGetChainID.Hname(), nil, blockIndexOrTrieRoot)
	if err != nil {
		return 0, err
	}

	return codec.DecodeUint16(ret.Get(evm.FieldResult))
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

func (c *ChainService) GetChainInfoByChainID(chainID isc.ChainID, blockIndexOrTrieRoot string) (*dto.ChainInfo, error) {
	chainRecord, err := c.chainRecordRegistryProvider.ChainRecord(chainID)
	if err != nil {
		return nil, err
	}

	ch, err := c.GetChainByID(chainID)
	if err != nil {
		return nil, err
	}

	governanceChainInfo, err := corecontracts.GetChainInfo(ch, blockIndexOrTrieRoot)
	if err != nil {
		if chainRecord != nil && errors.Is(err, interfaces.ErrChainNotFound) {
			return &dto.ChainInfo{ChainID: chainID, IsActive: false}, nil
		}

		return nil, err
	}

	chainInfo := dto.MapChainInfo(governanceChainInfo, chainRecord.Active)

	return chainInfo, nil
}

func (c *ChainService) GetContracts(chainID isc.ChainID, blockIndexOrTrieRoot string) (dto.ContractsMap, error) {
	ch, err := c.GetChainByID(chainID)
	if err != nil {
		return nil, err
	}
	recs, err := common.CallView(ch, root.Contract.Hname(), root.ViewGetContractRecords.Hname(), nil, blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}

	contracts, err := root.DecodeContractRegistry(collections.NewMapReadOnly(recs, root.VarContractRegistry))
	if err != nil {
		return nil, err
	}

	return contracts, nil
}

func (c *ChainService) GetState(chainID isc.ChainID, stateKey []byte) (state []byte, err error) {
	ch, err := c.GetChainByID(chainID)
	if err != nil {
		return nil, err
	}

	latestState, err := ch.LatestState(chainpkg.ActiveOrCommittedState)
	if err != nil {
		return nil, err
	}

	return latestState.Get(kv.Key(stateKey)), nil
}

func (c *ChainService) WaitForRequestProcessed(ctx context.Context, chainID isc.ChainID, requestID isc.RequestID, waitForL1Confirmation bool, timeout time.Duration) (*isc.Receipt, error) {
	ch, err := c.GetChainByID(chainID)
	if err != nil {
		return nil, err
	}

	ctxTimeout, ctxCancel := context.WithTimeout(ctx, timeout)
	defer ctxCancel()

	select {
	case receiptResponse := <-ch.AwaitRequestProcessed(ctxTimeout, requestID, waitForL1Confirmation):
		if receiptResponse == nil {
			return nil, nil
		}
		return common.ParseReceipt(ch, receiptResponse)
	case <-ctxTimeout.Done():
		return nil, apierrors.Timeout("timeout while waiting for request to be processed")
	}
}
