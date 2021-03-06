package vm

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/publisher"
)

type ContractEventPublisher struct {
	chainID  *iscp.ChainID
	contract iscp.Hname
	log      *logger.Logger
}

func NewContractEventPublisher(chainID *iscp.ChainID, contract iscp.Hname, log *logger.Logger) ContractEventPublisher {
	return ContractEventPublisher{
		chainID:  chainID,
		contract: contract,
		log:      log,
	}
}

func (c ContractEventPublisher) Publish(msg string) {
	c.log.Info(c.chainID.String() + "::" + c.contract.String() + "/event " + msg)
	publisher.Publish("vmmsg", c.chainID.Base58(), c.contract.String(), msg)
}
