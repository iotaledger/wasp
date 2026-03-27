package l1starter

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcnetwork "github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
)

var WaitUntilEffectsVisible = &iotagraphql.WaitParams{
	Attempts:             10,
	DelayBetweenAttempts: 1 * time.Second,
}

type LocalIotaNode struct {
	config           Config
	iscPackageOwner  iotasigner.Signer
	iscPackageID     iotago.PackageID
	nodeContainer    testcontainers.Container
	pgContainer      testcontainers.Container
	indexerContainer testcontainers.Container
	graphqlContainer testcontainers.Container
	network          *testcontainers.DockerNetwork
}

func NewLocalIotaNode(iscPackageOwner iotasigner.Signer) *LocalIotaNode {
	return &LocalIotaNode{
		iscPackageOwner: iscPackageOwner,
		config: Config{
			Host:   "http://localhost",
			Ports:  Ports{},
			Logger: Logger{},
		},
	}
}

func (in *LocalIotaNode) start(ctx context.Context) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 4*time.Minute)
	defer cancel()

	imagePlatform := in.getImagePlatform()
	networkName := in.setupNetwork(ctxTimeout)
	in.startPostgresContainer(ctxTimeout, networkName)

	now := time.Now()
	in.startNodeContainer(ctxTimeout, networkName, imagePlatform)
	in.startIndexerContainer(ctxTimeout, networkName, imagePlatform)
	in.startGraphQLContainer(ctxTimeout, networkName, imagePlatform)

	in.logf("Waiting for indexer to sync initial data...")
	time.Sleep(1 * time.Second)

	in.logf("Starting LocalIotaNode... done! took: %v", time.Since(now).Truncate(time.Millisecond))
	in.waitAllHealthy(ctxTimeout)
	in.deployISCContracts(ctxTimeout)
	in.logf("LocalIotaNode started successfully")
}

func (in *LocalIotaNode) getImagePlatform() string {
	if runtime.GOARCH == "arm64" {
		return "linux/arm64"
	}
	return "linux/amd64"
}

func (in *LocalIotaNode) setupNetwork(ctx context.Context) string {
	network, err := tcnetwork.New(ctx, tcnetwork.WithLabels(map[string]string{
		"com.wasp.test": "l1starter",
	}))
	if err != nil {
		// If network creation fails due to a stale reaper container conflict
		// (e.g. from a previous CI run), clean up and retry once.
		if strings.Contains(err.Error(), "reaper") {
			in.logf("Network creation failed due to stale reaper, cleaning up and retrying: %s", err)
			in.removeStaleReaperContainers(ctx)
			network, err = tcnetwork.New(ctx, tcnetwork.WithLabels(map[string]string{
				"com.wasp.test": "l1starter",
			}))
		}
		if err != nil {
			panic(fmt.Errorf("failed to create network: %w", err))
		}
	}
	in.network = network
	return network.Name
}

func (in *LocalIotaNode) removeStaleReaperContainers(ctx context.Context) {
	out, err := exec.CommandContext(ctx, "docker", "ps", "-aq", "--filter", "name=reaper_").Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		return
	}
	for _, id := range strings.Fields(strings.TrimSpace(string(out))) {
		in.logf("Removing stale reaper container: %s", id)
		_ = exec.CommandContext(ctx, "docker", "rm", "-f", id).Run()
	}
}

func (in *LocalIotaNode) startPostgresContainer(ctx context.Context, networkName string) {
	pgReq := testcontainers.ContainerRequest{
		Image:        "postgres:18",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgrespw",
			"POSTGRES_DB":       "iota_indexer",
		},
		Cmd:      []string{"-c", "max_connections=200"},
		Networks: []string{networkName},
		NetworkAliases: map[string][]string{
			networkName: {"postgres"},
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(2 * time.Minute),
	}

	in.logf("Starting Postgres container for indexer...")
	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: pgReq,
		Started:          true,
	})
	if err != nil {
		panic(fmt.Errorf("failed to start postgres container: %w", err))
	}
	in.pgContainer = pgContainer
}

func (in *LocalIotaNode) startNodeContainer(ctx context.Context, networkName, imagePlatform string) {
	nodeReq := testcontainers.ContainerRequest{
		Image:           "iotaledger/iota-tools:devnet",
		ImagePlatform:   imagePlatform,
		AlwaysPullImage: true,
		ExposedPorts:    []string{"9000/tcp", "9123/tcp"},
		Networks:      []string{networkName},
		NetworkAliases: map[string][]string{
			networkName: {"iota-node"},
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("9000/tcp"),
			wait.ForListeningPort("9123/tcp"),
		).WithDeadline(4 * time.Minute),
		Cmd: []string{
			"iota",
			"start",
			"--force-regenesis",
			"--with-faucet=0.0.0.0:9123",
			fmt.Sprintf("--faucet-amount=%d", iotagraphql.SingleCoinFundsFromFaucetAmount),
		},
	}

	if runtime.GOOS == "linux" {
		nodeReq.Tmpfs = map[string]string{"/tmp": ""}
	}

	in.logf("Starting LocalIotaNode...")
	nodeContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: nodeReq,
		Started:          true,
	})
	if err != nil {
		panic(fmt.Errorf("failed to start node container: %w", err))
	}
	in.nodeContainer = nodeContainer

	webAPIPort, err := nodeContainer.MappedPort(ctx, "9000")
	if err != nil {
		panic(fmt.Errorf("failed to get web API port: %w", err))
	}

	faucetPort, err := nodeContainer.MappedPort(ctx, "9123")
	if err != nil {
		panic(fmt.Errorf("failed to get faucet port: %w", err))
	}

	in.config.Ports.RPC = webAPIPort.Int()
	in.config.Ports.Faucet = faucetPort.Int()
}

func (in *LocalIotaNode) startIndexerContainer(ctx context.Context, networkName, imagePlatform string) {
	indexerReq := testcontainers.ContainerRequest{
		Image:           "iotaledger/iota-indexer:devnet",
		ImagePlatform:   imagePlatform,
		AlwaysPullImage: true,
		Networks:        []string{networkName},
		Entrypoint:    []string{"iota-indexer"},
		Cmd: []string{
			"--db-url=postgres://postgres:postgrespw@postgres:5432/iota_indexer",
			"--rpc-client-url=http://iota-node:9000",
			"--fullnode-sync-worker",
			"--reset-db",
		},
		WaitingFor: wait.ForLog("IOTA Indexer Writer").WithStartupTimeout(2 * time.Minute),
	}

	in.logf("Starting Indexer sync worker...")
	indexerContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: indexerReq,
		Started:          true,
	})
	if err != nil {
		panic(fmt.Errorf("failed to start indexer container: %w", err))
	}
	in.indexerContainer = indexerContainer
}

func (in *LocalIotaNode) startGraphQLContainer(ctx context.Context, networkName, imagePlatform string) {
	graphqlReq := testcontainers.ContainerRequest{
		Image:           "iotaledger/iota-graphql-rpc:devnet",
		ImagePlatform:   imagePlatform,
		AlwaysPullImage: true,
		ExposedPorts:    []string{"9125/tcp"},
		Networks:      []string{networkName},
		Entrypoint:    []string{"iota-graphql-rpc"},
		Cmd: []string{
			"start-server",
			"--host=0.0.0.0",
			"--port=9125",
			"--db-url=postgres://postgres:postgrespw@postgres:5432/iota_indexer",
			"--node-rpc-url=http://iota-node:9000",
		},
		WaitingFor: wait.ForListeningPort("9125/tcp").WithStartupTimeout(2 * time.Minute),
	}

	in.logf("Starting GraphQL server...")
	graphqlContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: graphqlReq,
		Started:          true,
	})
	if err != nil {
		panic(fmt.Errorf("failed to start graphql container: %w", err))
	}
	in.graphqlContainer = graphqlContainer

	graphqlPort, err := graphqlContainer.MappedPort(ctx, "9125")
	if err != nil {
		panic(fmt.Errorf("failed to get graphql port: %w", err))
	}
	in.config.Ports.GraphQL = graphqlPort.Int()
	in.logf("GraphQL container (ID: %s) mapped to localhost:%d", graphqlContainer.GetContainerID()[:12], graphqlPort.Int())
}

func (in *LocalIotaNode) deployISCContracts(ctx context.Context) {
	in.logf("Deploying ISC contracts...")
	packageID, err := in.L1Client().L2().DeployISCContracts(ctx, ISCPackageOwner)
	if err != nil {
		panic(fmt.Errorf("isc contract deployment failed: %w", err))
	}
	in.iscPackageID = packageID
}

func (in *LocalIotaNode) stop(ctx context.Context) {
	in.logf("Stopping...")
	if in.graphqlContainer != nil {
		err := in.graphqlContainer.Terminate(ctx, testcontainers.StopTimeout(0))
		if err != nil {
			in.logf("Failed to stop graphql container: %s", err)
		}
	}
	if in.indexerContainer != nil {
		err := in.indexerContainer.Terminate(ctx, testcontainers.StopTimeout(0))
		if err != nil {
			in.logf("Failed to stop indexer container: %s", err)
		}
	}
	if in.nodeContainer != nil {
		err := in.nodeContainer.Terminate(ctx, testcontainers.StopTimeout(0))
		if err != nil {
			in.logf("Failed to stop node container: %s", err)
		}
	}
	if in.pgContainer != nil {
		err := in.pgContainer.Terminate(ctx, testcontainers.StopTimeout(0))
		if err != nil {
			in.logf("Failed to stop postgres container: %s", err)
		}
	}
	if in.network != nil {
		if err := in.network.Remove(ctx); err != nil {
			in.logf("Failed to remove network: %s", err)
		}
	}
	instance.Store(nil)
}

func (in *LocalIotaNode) ISCPackageID() iotago.PackageID {
	return in.iscPackageID
}

func (in *LocalIotaNode) APIURL() string {
	return fmt.Sprintf("%s:%d", in.config.Host, in.config.Ports.GraphQL)
}

func (in *LocalIotaNode) FaucetURL() string {
	return fmt.Sprintf("%s:%d/gas", in.config.Host, in.config.Ports.Faucet)
}

func (in *LocalIotaNode) L1Client() clients.L1Client {
	return clients.NewL1Client(clients.L1Config{
		APIURL:    in.APIURL(),
		FaucetURL: in.FaucetURL(),
	}, WaitUntilEffectsVisible)
}

func (in *LocalIotaNode) IsLocal() bool {
	return true
}

func (in *LocalIotaNode) waitAllHealthy(ctx context.Context) {
	ts := time.Now()
	in.logf("Using temporary folder: %s", in.config.TempDir)
	in.logf("Waiting for all IOTA nodes to become healthy...")

	tryLoop := func(f func() bool) {
		for {
			if ctx.Err() != nil {
				panic("nodes didn't become healthy in time")
			}
			if f() {
				return
			}
			in.logf("Waiting until LocalIotaNode becomes ready. Time waiting: %v", time.Since(ts).Truncate(time.Millisecond))
			time.Sleep(500 * time.Millisecond)
		}
	}

	tryLoop(func() bool {
		res, err := in.L1Client().GetLatestIotaSystemState(ctx)
		if err != nil {
			in.logf("StatusLoop: err: %s", err)
		}
		if err != nil || res == nil {
			return false
		}
		if res.Epoch.ValidatorSet.PendingActiveValidatorsSize != 0 {
			return false
		}
		return true
	})

	tryLoop(func() bool {
		err := in.L1Client().RequestFundsFromFaucet(ctx, ISCPackageOwner.Address())
		if err != nil {
			in.logf("FaucetLoop: err: %s", err)
		}
		return err == nil
	})

	in.logf("Waiting for faucet funds to arrive...")
	tryLoop(func() bool {
		balances, err := in.L1Client().GetAllBalances(ctx, ISCPackageOwner.Address())
		if err != nil {
			in.logf("CheckBalanceLoop: err: %s", err)
			return false
		}
		if len(balances) == 0 {
			in.logf("CheckBalanceLoop: no balances yet")
			return false
		}
		// Check if we have any balance with a non-zero amount
		for _, bal := range balances {
			if bal.TotalBalance != nil && bal.TotalBalance.Uint64() > 0 {
				in.logf("CheckBalanceLoop: found balance of %s", bal.TotalBalance.String())
				return true
			}
		}
		in.logf("CheckBalanceLoop: balances exist but all are zero")
		return false
	})

	in.logf("Waiting until LocalIotaNode becomes ready... done! took: %v", time.Since(ts).Truncate(time.Millisecond))
}

func (in *LocalIotaNode) logf(msg string, args ...any) {
	if in.config.Logger != nil {
		in.config.Logger.Printf("Iota Node: "+msg+"\n", args...)
	}
}
