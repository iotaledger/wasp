package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/core/marshalutil"
	"github.com/iotaledger/wasp/packages/chainutil"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/util/expiringcache"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type OffLedgerService struct {
	chainService    interfaces.ChainService
	networkProvider peering.NetworkProvider
	requestCache    *expiringcache.ExpiringCache
}

func NewOffLedgerService(chainService interfaces.ChainService, networkProvider peering.NetworkProvider, requestCacheTTL time.Duration) interfaces.OffLedgerService {
	return &OffLedgerService{
		chainService:    chainService,
		networkProvider: networkProvider,
		requestCache:    expiringcache.New(requestCacheTTL),
	}
}

func (c *OffLedgerService) ParseRequest(binaryRequest []byte) (isc.OffLedgerRequest, error) {
	request, err := isc.NewRequestFromMarshalUtil(marshalutil.New(binaryRequest))
	if err != nil {
		return nil, errors.New("error parsing request from payload")
	}

	req, ok := request.(isc.OffLedgerRequest)
	if !ok {
		return nil, errors.New("error parsing request: off-ledger request expected")
	}

	return req, nil
}

func (c *OffLedgerService) EnqueueOffLedgerRequest(chainID isc.ChainID, binaryRequest []byte) error {
	request, err := c.ParseRequest(binaryRequest)
	if err != nil {
		return err
	}

	reqID := request.ID()

	if c.requestCache.Get(reqID) != nil {
		return fmt.Errorf("request already processed")
	}

	// check req signature
	if err := request.VerifySignature(); err != nil {
		c.requestCache.Set(reqID, true)
		return fmt.Errorf("could not verify: %s", err.Error())
	}

	// check req is for the correct chain
	if !request.ChainID().Equals(chainID) {
		// do not add to cache, it can still be sent to the correct chain
		return fmt.Errorf("Request is for a different chain")
	}

	// check chain exists
	chain := c.chainService.GetChainByID(chainID)
	if chain == nil {
		return fmt.Errorf("Unknown chain: %s", chainID.String())
	}

	alreadyProcessed, err := chainutil.HasRequestBeenProcessed(chain, reqID)
	if err != nil {
		return fmt.Errorf("internal error")
	}

	defer c.requestCache.Set(reqID, true)

	if alreadyProcessed {
		return fmt.Errorf("request already processed")
	}

	// check user has on-chain balance
	assets, err := chainutil.GetAccountBalance(chain, request.SenderAccount())
	if err != nil {
		return fmt.Errorf("Unable to get account balance")
	}

	if assets.IsEmpty() {
		return fmt.Errorf("No balance on account %s", request.SenderAccount().String())
	}

	if err := chainutil.CheckNonce(chain, request); err != nil {
		return fmt.Errorf("invalid nonce, %v", err)
	}

	chain.ReceiveOffLedgerRequest(request, c.networkProvider.Self().PubKey())

	return nil
}
