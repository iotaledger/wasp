package parameters_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

func TestL1Syncer(t *testing.T) {
	client := iotaclient.NewHTTP(iotaconn.AlphanetEndpointURL, iotaclient.WaitForEffectsDisabled)
	logger := testlogger.NewLogger(t)
	l1syncer := parameters.NewL1Syncer(client, 1000*time.Millisecond, logger)
	go l1syncer.Start()
	time.Sleep(500 * time.Millisecond)
	fmt.Println("parameters.L1Params", parameters.L1())
}
