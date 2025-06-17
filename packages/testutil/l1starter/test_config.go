package l1starter

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"

	"github.com/iotaledger/wasp/packages/testutil/testconfig"
)

type L1EndpointConfig struct {
	IsLocal       bool
	RandomizeSeed bool
	APIURL        string
	FaucetURL     string
}

func TryDockerAvailability(ctx context.Context) error {
	provider, err := testcontainers.ProviderDocker.GetProvider()
	if err != nil {
		return err
	}

	err = provider.Health(ctx)
	if err != nil {
		return err
	}

	return nil
}

func LoadConfig() *L1EndpointConfig {
	c, configFound := testconfig.LoadConfig("l1starter")

	if !configFound {
		fmt.Println("No l1starter config found - using local node")

		err := TryDockerAvailability(context.Background())
		if err != nil {
			panic(fmt.Errorf("docker unavailable: %v", err))
		}

		return &L1EndpointConfig{
			IsLocal:       true,
			RandomizeSeed: true,
		}
	}

	isLocal := c.Bool("IS_LOCAL")

	if isLocal {
		return &L1EndpointConfig{
			IsLocal: true,
		}
	}

	return &L1EndpointConfig{
		IsLocal:   false,
		APIURL:    c.String("API_URL"),
		FaucetURL: c.String("FAUCET_URL"),
	}
}
