package main

import (
	"context"
	"fmt"
	"os"

	cmd "github.com/urfave/cli/v2"

	stardust_apiclient "github.com/nnikolash/wasp-types-exported/clients/apiextensions"

	rebased_apiclient "github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
	webapi_validation "github.com/iotaledger/wasp/tools/stardust-migration/webapi-validation"
)

func validateWebAPI(c *cmd.Context) error {
	go func() {
		<-c.Done()
		cli.Logf("Interrupted")
		os.Exit(1)
	}()

	firstIndex := uint32(c.Uint64("from-block"))
	lastIndex := uint32(c.Uint64("to-block"))

	stardustEndPoint := c.Args().Get(0)
	rebasedEndpoint := c.Args().Get(1)

	cli.DebugLoggingEnabled = true

	_ = firstIndex
	_ = lastIndex

	defer func() {
		if err := recover(); err != nil {
			cli.Logf("Validation panicked")
			panic(err)
		}
	}()

	sClient, err := stardust_apiclient.WaspAPIClientByHostName(stardustEndPoint)
	if err != nil {
		return fmt.Errorf("failed to create Rebased API Client: %v for '%s'", err, rebasedEndpoint)
	}

	rClient, err := rebased_apiclient.WaspAPIClientByHostName(rebasedEndpoint)
	if err != nil {
		return fmt.Errorf("failed to create Rebased API Client: %v for '%s'", err, rebasedEndpoint)
	}

	ctx := context.Background()

	testContext := webapi_validation.NewValidationContext(ctx, sClient, rClient)
	chainValidation := webapi_validation.NewChainValidation(testContext)
	coreAccountsValidation := webapi_validation.NewCoreAccountsValidation(testContext)

	chainValidation.Validate(0)
	coreAccountsValidation.Validate(0)

	return nil
}
