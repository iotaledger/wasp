package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/peering"
	"github.com/iotaledger/wasp/v2/packages/util/expiringcache"
	"github.com/iotaledger/wasp/v2/packages/webapi/interfaces"
)

type OffLedgerService struct {
	chainService    interfaces.ChainService
	networkProvider peering.NetworkProvider
	requestCache    *expiringcache.ExpiringCache[isc.RequestID, bool]
}

func NewOffLedgerService(chainService interfaces.ChainService, networkProvider peering.NetworkProvider, requestCacheTTL time.Duration) interfaces.OffLedgerService {
	return &OffLedgerService{
		chainService:    chainService,
		networkProvider: networkProvider,
		requestCache:    expiringcache.New[isc.RequestID, bool](requestCacheTTL),
	}
}

func (c *OffLedgerService) ParseRequest(binaryRequest []byte) (isc.Request, error) {
	// check offledger kind (avoid deserialization otherwise)
	if !isc.IsOffledgerKind(binaryRequest[0]) {
		return nil, errors.New("error parsing request: off-ledger request expected")
	}
	request, err := isc.RequestFromBytes(binaryRequest)
	if err != nil {
		return nil, errors.New("error parsing request from payload")
	}

	return request, nil
}

func (c *OffLedgerService) EnqueueOffLedgerRequest(chainID isc.ChainID, binaryRequest []byte) error {
	request, err := c.ParseRequest(binaryRequest)
	if err != nil {
		return err
	}

	reqID := request.ID()

	if c.requestCache.Get(reqID) != nil {
		return errors.New("request already processed")
	}

	asOffLedgerRequest, ok := request.(isc.OffLedgerRequest)
	if !ok {
		return errors.New("error parsing request: off-ledger request expected")
	}

	// check req signature
	if err2 := asOffLedgerRequest.VerifySignature(); err2 != nil {
		return fmt.Errorf("could not verify: %w", err2)
	}

	// check chain exists
	chain, err := c.chainService.GetChain()
	if err != nil {
		return err
	}

	// check req is for the correct chain
	if !asOffLedgerRequest.ChainID().Equals(chainID) {
		// do not add to cache, it can still be sent to the correct chain
		return errors.New("request is for a different chain")
	}

	if err := chain.ReceiveOffLedgerRequest(asOffLedgerRequest, c.networkProvider.Self().PubKey()); err != nil {
		return fmt.Errorf("tx not added to the mempool: %v", err.Error())
	}

	c.requestCache.Set(reqID, true)
	return nil
}
