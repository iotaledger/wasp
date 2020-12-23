package vm

import (
	"fmt"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/publisher"
)

type ContractEventPublisher struct {
	contractID coretypes.ContractID
	log        *logger.Logger
}

func NewContractEventPublisher(contractID coretypes.ContractID, log *logger.Logger) ContractEventPublisher {
	return ContractEventPublisher{
		contractID: contractID,
		log:        log,
	}
}

func (c ContractEventPublisher) Publish(msg string) {
	c.log.Info(c.contractID.String() + "/event " + msg)
	publisher.Publish("vmmsg", c.contractID.ChainID().String(), fmt.Sprintf("%d", c.contractID.Hname()), msg)
}

func (c ContractEventPublisher) Publishf(format string, args ...interface{}) {
	c.log.Infof(c.contractID.String()+"/event "+format, args...)
	publisher.Publish("vmmsg", c.contractID.ChainID().String(), fmt.Sprintf("%d", c.contractID.Hname()), fmt.Sprintf(format, args...))
}
