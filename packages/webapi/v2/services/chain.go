package services

import (
	"github.com/iotaledger/hive.go/core/logger"
	chainpkg "github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	metricspkg "github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	walpkg "github.com/iotaledger/wasp/packages/wal"
	"github.com/iotaledger/wasp/packages/webapi/v2/dto"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type ChainService struct {
	log *logger.Logger

	chainsProvider   chains.Provider
	metrics          *metricspkg.Metrics
	registryProvider registry.Provider
	vmService        interfaces.VM
	wal              *walpkg.WAL
}

func NewChainService(log *logger.Logger, chainsProvider chains.Provider, metrics *metricspkg.Metrics, registryProvider registry.Provider, vmService interfaces.VM, wal *walpkg.WAL) interfaces.Chain {
	return &ChainService{
		log: log,

		chainsProvider:   chainsProvider,
		metrics:          metrics,
		registryProvider: registryProvider,
		vmService:        vmService,
		wal:              wal,
	}
}

func (c *ChainService) ActivateChain(chainID *isc.ChainID) error {
	chainRecord, err := c.registryProvider().ActivateChainRecord(chainID)
	if err != nil {
		return err
	}

	c.log.Debugw("calling Chains.Activate", "chainID", chainID.String())

	err = c.chainsProvider().
		Activate(chainRecord, c.registryProvider, c.metrics, c.wal)

	return err
}

func (c *ChainService) DeactivateChain(chainID *isc.ChainID) error {
	chainRecord, err := c.registryProvider().DeactivateChainRecord(chainID)
	if err != nil {
		return err
	}

	c.log.Debugw("calling Chains.Activate", "chainID", chainID.String())

	err = c.chainsProvider().Deactivate(chainRecord)

	return err
}

func (c *ChainService) GetChainByID(chainID *isc.ChainID) chainpkg.Chain {
	chain := c.chainsProvider().Get(chainID, true)

	return chain
}

func (c *ChainService) GetEVMChainID(chainID *isc.ChainID) (uint16, error) {
	ret, err := c.vmService.CallViewByChainID(chainID, evm.Contract.Hname(), evm.FuncGetChainID.Hname(), nil)
	if err != nil {
		return 0, err
	}

	return codec.DecodeUint16(ret.MustGet(evm.FieldResult))
}

func (c *ChainService) GetAllChainIDs() ([]*isc.ChainID, error) {
	records, err := c.registryProvider().GetChainRecords()
	if err != nil {
		return nil, err
	}

	chainIDs := make([]*isc.ChainID, 0, len(records))

	for _, chainRecord := range records {
		chainIDs = append(chainIDs, &chainRecord.ChainID)
	}

	return chainIDs, nil
}

func (c *ChainService) GetChainInfoByChainID(chainID *isc.ChainID) (dto.ChainInfo, error) {
	info, err := c.vmService.CallViewByChainID(chainID, governance.Contract.Hname(), governance.ViewGetChainInfo.Hname(), nil)
	if err != nil {
		return nil, err
	}

	chainInfo, err := governance.GetChainInfo(info)

	return chainInfo, err
}

func (c *ChainService) GetContracts(chainID *isc.ChainID) (dto.ContractsMap, error) {
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
