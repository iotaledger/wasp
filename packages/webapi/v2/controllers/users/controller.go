package users

import (
	"github.com/pangpanglabs/echoswagger/v2"

	loggerpkg "github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type Controller struct {
	log *loggerpkg.Logger

	offLedgerService interfaces.OffLedgerService
	peeringService   interfaces.PeeringService
	vmService        interfaces.VMService
}

func NewUsersController(log *loggerpkg.Logger, offLedgerService interfaces.OffLedgerService, peeringService interfaces.PeeringService, vmService interfaces.VMService) interfaces.APIController {
	return &Controller{
		log:              log,
		offLedgerService: offLedgerService,
		peeringService:   peeringService,
		vmService:        vmService,
	}
}

func (c *Controller) Name() string {
	return "requests"
}

func (c *Controller) RegisterPublic(publicAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {
}

func (c *Controller) RegisterAdmin(adminAPI echoswagger.ApiGroup, mocker interfaces.Mocker) {

}
