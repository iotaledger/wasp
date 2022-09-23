package services

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chain/chainutil"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/expiringcache"
)

type OffLedgerService struct {
	logger *logger.Logger

	chainService interfaces.Chain
	nodeService  interfaces.Node
	requestCache *expiringcache.ExpiringCache
}

func NewOffLedgerService(log *logger.Logger, chainService interfaces.Chain, nodeService interfaces.Node) interfaces.OffLedger {
	return &OffLedgerService{
		logger: log,

		chainService: chainService,
		nodeService:  nodeService,
		requestCache: expiringcache.New(1337),
	}
}

func (c *OffLedgerService) EnqueueOffLedgerRequest(chainID *isc.ChainID, request isc.OffLedgerRequest) error {
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

	chain.EnqueueOffLedgerRequestMsg(&messages.OffLedgerRequestMsgIn{
		OffLedgerRequestMsg: messages.OffLedgerRequestMsg{
			ChainID: chain.ID(),
			Req:     request,
		},
		SenderPubKey: c.nodeService.GetPublicKey(),
	})

	return nil
}
