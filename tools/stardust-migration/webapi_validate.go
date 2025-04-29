package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"

	"github.com/stretchr/testify/require"
	cmd "github.com/urfave/cli/v2"

	stardust_apiclient "github.com/nnikolash/wasp-types-exported/clients/apiextensions"
	old_parameters "github.com/nnikolash/wasp-types-exported/packages/parameters"

	old_iotago "github.com/iotaledger/iota.go/v3"
	rebased_apiclient "github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
	webapi_validation "github.com/iotaledger/wasp/tools/stardust-migration/webapi-validation"
	"github.com/iotaledger/wasp/tools/stardust-migration/webapi-validation/base"
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

	old_parameters.InitL1(&old_parameters.L1Params{
		Protocol: &old_iotago.ProtocolParameters{
			Bech32HRP: old_iotago.PrefixMainnet,
		},
		BaseToken: &old_parameters.BaseToken{
			Decimals: 6,
		},
	})

	ctx := context.Background()

	testContext := base.NewValidationContext(ctx, sClient, rClient)
	chainValidation := webapi_validation.NewChainValidation(testContext)
	coreBlockValidation := webapi_validation.NewCoreBlockLogValidation(testContext)
	accountValidation := webapi_validation.NewAccountValidation(testContext)
	governanceValidation := webapi_validation.NewGovernanceValidation(testContext)
	evmValidation := webapi_validation.NewEvmValidation(stardustEndPoint, rebasedEndpoint, testContext)
	latestBlock, _, err := rClient.CorecontractsAPI.BlocklogGetLatestBlockInfo(ctx).Execute()
	require.NoError(base.T, err)

	log.Printf("Starting Stardust/Rebased WebAPI validation. From 1 => %d", latestBlock.BlockIndex)

	for i := uint32(1); i < latestBlock.BlockIndex; i++ {
		if i%100 == 0 {
			fmt.Printf("StateIndex: %d \n", i)
		}

		chainValidation.Validate(i)
		coreBlockValidation.Validate(i)
		accountValidation.ValidateAccountBalances(i)
		accountValidation.ValidateNFTs(i)
		accountValidation.ValidateNonce(i)
		accountValidation.ValidateTotalAssets(i)
		governanceValidation.ValidateGovernance(i)
	}

	evmResults, err := os.Create("evm_validation_results.log")
	if err != nil {
		return fmt.Errorf("failed to create evm_validation_results file: %v", err)
	}
	defer evmResults.Close()

	lastNBlocks := 50000
	blockCount := atomic.Int64{}
	sem := make(chan struct{}, 20)
	wg := sync.WaitGroup{}
	wg.Add(lastNBlocks)
	for i := uint32(latestBlock.BlockIndex - uint32(lastNBlocks) + 1); i <= latestBlock.BlockIndex; i++ {
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			err := evmValidation.ValidateEvm(uint64(i))
			if err != nil {
				evmResults.WriteString(fmt.Sprintf("block %d: %v\n", i, err))
			}
			blockCount.Add(1)
			fmt.Printf("processed %d blocks\r", blockCount.Load())
			<-sem
		}()
	}

	wg.Wait()
	fmt.Printf("processed %d evm blocks\n", blockCount.Load())

	return nil
}
