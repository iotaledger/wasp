package services

import (
	"context"
	"errors"
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	chainpkg "github.com/iotaledger/wasp/v2/packages/chain"
	"github.com/iotaledger/wasp/v2/packages/chains"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/metrics"
	"github.com/iotaledger/wasp/v2/packages/registry"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/root"
	"github.com/iotaledger/wasp/v2/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/v2/packages/webapi/common"
	"github.com/iotaledger/wasp/v2/packages/webapi/corecontracts"
	"github.com/iotaledger/wasp/v2/packages/webapi/dto"
	"github.com/iotaledger/wasp/v2/packages/webapi/interfaces"
)

type ChainService struct {
	log                         log.Logger
	chainsProvider              chains.Provider
	chainMetricsProvider        *metrics.ChainMetricsProvider
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider
}

func NewChainService(
	logger log.Logger,
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

func (c *ChainService) ActivateChain() error {
	_, err := c.chainRecordRegistryProvider.ActivateChainRecord(chainID)
	if err != nil {
		return err
	}

	return c.chainsProvider().Activate(chainID)
}

func (c *ChainService) DeactivateChain() error {
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

	c.log.LogInfof("StoredChainRec %v %v", storedChainRec, err)

	if storedChainRec != nil {
		_, err = c.chainRecordRegistryProvider.UpdateChainRecord(
			chainRecord.ChainID(),
			func(rec *registry.ChainRecord) bool {
				rec.AccessNodes = chainRecord.AccessNodes
				rec.Active = chainRecord.Active
				return true
			},
		)
		c.log.LogInfof("UpdatechainRec %v %v", chainRecord, err)

		if err != nil {
			return err
		}
	} else {
		if err := c.chainRecordRegistryProvider.AddChainRecord(chainRecord); err != nil {
			c.log.LogInfof("AddChainRec %v %v", chainRecord, err)

			return err
		}
	}

	// Activate/deactivate the chain accordingly.
	c.log.LogInfof("Chainrecord active %v", chainRecord.Active)

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

func (c *ChainService) GetChain() (chainpkg.Chain, error) {
	return c.chainsProvider().GetFirst()
}

func (c *ChainService) GetEVMChainID(blockIndexOrTrieRoot string) (uint16, error) {
	ch, err := c.GetChain()
	if err != nil {
		return 0, err
	}
	ret, err := common.CallView(ch, evm.ViewGetChainID.Message(), blockIndexOrTrieRoot)
	if err != nil {
		return 0, err
	}
	return evm.ViewGetChainID.DecodeOutput(ret)
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

func (c *ChainService) GetChainInfo(blockIndexOrTrieRoot string) (*dto.ChainInfo, error) {
	ch, err := c.GetChain()
	if err != nil {
		return nil, err
	}

	chainRecord, err := c.chainRecordRegistryProvider.ChainRecord(ch.ID())
	if err != nil {
		return nil, err
	}

	governanceChainInfo, err := corecontracts.GetChainInfo(ch, blockIndexOrTrieRoot)
	if err != nil {
		if chainRecord != nil && errors.Is(err, interfaces.ErrChainNotFound) {
			return &dto.ChainInfo{ChainID: ch.ID(), IsActive: false}, nil
		}

		return nil, err
	}

	chainInfo := dto.MapChainInfo(governanceChainInfo, chainRecord.Active)

	return chainInfo, nil
}

func (c *ChainService) GetContracts(blockIndexOrTrieRoot string) ([]lo.Tuple2[*isc.Hname, *root.ContractRecord], error) {
	ch, err := c.GetChain()
	if err != nil {
		return nil, err
	}
	res, err := common.CallView(ch, root.ViewGetContractRecords.Message(), blockIndexOrTrieRoot)
	if err != nil {
		return nil, err
	}
	return root.ViewGetContractRecords.DecodeOutput(res)
}

func (c *ChainService) GetState(stateKey []byte) (state []byte, err error) {
	ch, err := c.GetChain()
	if err != nil {
		return nil, err
	}

	latestState, err := ch.LatestState(chainpkg.ActiveOrCommittedState)
	if err != nil {
		return nil, err
	}

	return latestState.Get(kv.Key(stateKey)), nil
}

func (c *ChainService) WaitForRequestProcessed(ctx context.Context, requestID isc.RequestID, waitForL1Confirmation bool, timeout time.Duration) (*isc.Receipt, error) {
	ch, err := c.GetChain()
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

func (c *ChainService) RotateTo(ctx context.Context, rotateToAddress *iotago.Address) error {
	ch, err := c.GetChain()
	if err != nil {
		return err
	}
	ch.RotateTo(rotateToAddress)
	return nil
}
