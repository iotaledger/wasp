package l1starter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/testcontainers/testcontainers-go"
)

type L1EndpointConfig struct {
	IsLocal   bool
	APIURL    string
	FaucetURL string
}

func GetRootDir() string {
	// Start from current working directory
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Walk up until we find go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			panic("could not find go.mod file")
		}
		dir = parent
	}
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
	c := viper.New()
	c.SetConfigName(".testconfig")
	c.SetConfigType("json")
	c.AddConfigPath(GetRootDir())
	c.SetEnvPrefix("TEST")
	c.AutomaticEnv()

	if err := c.ReadInConfig(); err != nil {
		fmt.Println(".testconfig not found, using local node")

		err = TryDockerAvailability(context.Background())
		if err != nil {
			panic(fmt.Errorf("docker unavailable: %v", err))
		}

		return &L1EndpointConfig{
			IsLocal: true,
		}
	}

	isLocal := c.GetBool("IS_LOCAL")
	if isLocal {
		return &L1EndpointConfig{
			IsLocal: true,
		}
	}

	return &L1EndpointConfig{
		IsLocal:   false,
		APIURL:    c.GetString("API_URL"),
		FaucetURL: c.GetString("FAUCET_URL"),
	}
}
