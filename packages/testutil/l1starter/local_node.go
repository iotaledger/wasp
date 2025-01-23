package l1starter

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/juju/fslock"
	"github.com/samber/lo"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
)

var WaitUntilEffectsVisible = &iotaclient.WaitParams{
	Attempts:             10,
	DelayBetweenAttempts: 1 * time.Second,
}

type LocalIotaNode struct {
	ctx             context.Context
	config          Config
	iscPackageOwner iotasigner.Signer
	container       testcontainers.Container
	Logger          testcontainers.Logging
}

type sharedLocalNodeInfo struct {
	UseCount int
	Config   Config
	RunID    string
}

type sharedLocalNodeUsageStats struct {
	MaxUseCount int
	TotalUses   int
	Startups    int
}

func NewLocalIotaNode(iscPackageOwner iotasigner.Signer) *LocalIotaNode {
	return &LocalIotaNode{
		iscPackageOwner: iscPackageOwner,
		config: Config{
			Host:  "http://localhost",
			Ports: Ports{},
		},
		Logger: Logger{},
	}
}

// Returns hash string based for all of the input parameters of the node.
func (in *LocalIotaNode) configHash() string {
	configHashBytes := md5.Sum(in.iscPackageOwner.Address().Bytes())
	configHash := hex.EncodeToString(configHashBytes[:])
	return configHash
}

var tmpTestFilesDir = os.TempDir() + "/wasp/testing"

func (in *LocalIotaNode) localNodeInfoFilePath() string {
	configHash := in.configHash()

	localNodeInfoFilePath := tmpTestFilesDir + "/shared-local-iota-node-" + configHash

	return localNodeInfoFilePath
}

func (in *LocalIotaNode) localNodeUsageStatsFilePath() string {
	return tmpTestFilesDir + "/shared-local-iota-node-usage-stats"
}

func (in *LocalIotaNode) nodeInfoHasExpired() bool {
	infoFileStat := lo.Must(os.Stat(in.localNodeInfoFilePath()))
	lastKeepaliveTime := infoFileStat.ModTime()

	keepaliveExpires := time.Since(lastKeepaliveTime) > 2*keepaliveInterval

	return keepaliveExpires
}

func (in *LocalIotaNode) lockAndModifyLocalNodeInfo() (_ *sharedLocalNodeInfo, _ *sharedLocalNodeUsageStats, unlockLocalNodeInfo func()) {
	in.logf("Locking shared local node info file...")
	info, unlockInfo := openJSONFile[sharedLocalNodeInfo](in, "local node info", in.localNodeInfoFilePath())

	if in.nodeInfoHasExpired() {
		in.logf("Local node info file is too old - ignoring old container if exists")
		*info = sharedLocalNodeInfo{}
	}

	if info.RunID == "" {
		info.RunID = strconv.Itoa(rand.Int())
	}

	in.logf("Locking shared local node usage stats file...")
	stats, unlockStats := openJSONFile[sharedLocalNodeUsageStats](in, "local node stats", in.localNodeUsageStatsFilePath())

	return info, stats, sync.OnceFunc(func() {
		unlockStats()
		unlockInfo()
	})
}

func openJSONFile[Data any](in *LocalIotaNode, hmName, filePath string) (_ *Data, saveAndReleaseFile func()) {
	var data Data

	lo.Must0(os.MkdirAll(path.Dir(filePath), 0755))

	fLock := fslock.New(filePath + ".lock")
	fLock.Lock()

	f, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(fmt.Errorf("failed to open %v file at %v: %w", hmName, filePath, err))
		}

		in.logf("%v file does not exist - creating: %s", hmName, filePath)
		f = lo.Must(os.Create(filePath))
		lo.Must0(json.NewEncoder(f).Encode(lo.Empty[Data]()))
		lo.Must(f.Seek(0, 0))
	} else {
		in.logf("Reading %v file: %s", hmName, filePath)
		lo.Must0(json.NewDecoder(f).Decode(&data))
	}

	return &data, sync.OnceFunc(func() {
		in.logf("Writing %v file...", hmName)
		lo.Must0(f.Truncate(0))
		lo.Must(f.Seek(0, 0))
		lo.Must0(json.NewEncoder(f).Encode(data))
		lo.Must0(f.Sync())
		lo.Must0(f.Close())
		fLock.Unlock()
	})
}

func (in *LocalIotaNode) deleteLocalNodeInfo() {
	localNodeInfoFilePath := in.localNodeInfoFilePath()

	in.logf("Deleting shared local node info file: %s", localNodeInfoFilePath)
	lo.Must0(os.Remove(localNodeInfoFilePath))
	lo.Must0(os.Remove(localNodeInfoFilePath + ".lock"))
}

func (in *LocalIotaNode) start(ctx context.Context) {
	in.ctx = ctx

	localNodeInfo, localNodeStats, unlockLocalNodeInfo := in.lockAndModifyLocalNodeInfo()
	defer unlockLocalNodeInfo()

	if localNodeInfo.UseCount == 0 {
		in.logf("Starting LocalIotaNode...")
	} else {
		in.logf("LocalIotaNode already started, reusing...")
	}

	contName := "wasp-iota-node-" + in.configHash() + "-" + localNodeInfo.RunID

	imagePlatform := "linux/amd64"
	if runtime.GOARCH == "arm64" {
		imagePlatform = "linux/arm64"
	}

	req := testcontainers.ContainerRequest{
		Name:          contName,
		Image:         "iotaledger/iota-tools:v0.9.0-alpha",
		ImagePlatform: imagePlatform,
		ExposedPorts:  []string{"9000/tcp", "9123/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("9000/tcp"),
			wait.ForListeningPort("9123/tcp"),
		).WithDeadline(30 * time.Second),
		Cmd: []string{
			"iota",
			"start",
			"--force-regenesis",
			fmt.Sprintf("--epoch-duration-ms=%d", 60000),
			"--with-faucet",
			fmt.Sprintf("--faucet-amount=%d", iotaclient.SingleCoinFundsFromFaucetAmount),
		},
	}

	now := time.Now()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          localNodeInfo.UseCount == 0,
		Reuse:            localNodeInfo.UseCount != 0,
	})
	if err != nil {
		panic(fmt.Errorf("failed to start/reuse container: %w", err))
	}

	container.SessionID()
	in.container = container

	if localNodeInfo.UseCount == 0 {
		webAPIPort, err := container.MappedPort(ctx, "9000")
		if err != nil {
			container.Terminate(ctx)
			panic(fmt.Errorf("failed to get web API port: %w", err))
		}

		faucetPort, err := container.MappedPort(ctx, "9123")
		if err != nil {
			container.Terminate(ctx)
			panic(fmt.Errorf("failed to get faucet port: %w", err))
		}

		in.config.Ports.RPC = webAPIPort.Int()
		in.config.Ports.Faucet = faucetPort.Int()

		in.logf("Starting LocalIotaNode... done! took: %v", time.Since(now).Truncate(time.Millisecond))
		in.waitAllHealthy(5 * time.Minute)

		// TODO: should we deploy separate ISC contracts for each "user" of the node?
		in.logf("Deploying ISC contracts...")
		packageID, err := in.L1Client().DeployISCContracts(ctx, in.iscPackageOwner)
		if err != nil {
			panic(fmt.Errorf("isc contract deployment failed: %w", err))
		}

		in.config.IscPackageID = packageID

		localNodeInfo.Config = in.config
		localNodeStats.Startups++

		in.logf("LocalIotaNode started successfully")
	} else {
		in.config = localNodeInfo.Config
		in.logf("Connected to existing LocalIotaNode container: current users count = %v", localNodeInfo.UseCount+1)
	}

	localNodeInfo.UseCount++
	localNodeStats.TotalUses++
	localNodeStats.MaxUseCount = max(localNodeInfo.UseCount, localNodeStats.MaxUseCount)

	startLocalNodeKeepalive(in.localNodeInfoFilePath())
}

func (in *LocalIotaNode) stop() {
	stopLocalNodeKeepalive(in.localNodeInfoFilePath())

	localNodeInfo, _, unlockLocalNodeInfo := in.lockAndModifyLocalNodeInfo()
	defer unlockLocalNodeInfo()

	localNodeInfo.UseCount--

	if localNodeInfo.UseCount == 0 {
		in.logf("Stopping LocalIotaNode...")
		in.container.Terminate(context.Background(), testcontainers.StopTimeout(0))
		in.deleteLocalNodeInfo()
	} else {
		in.logf("LocalIotaNode still used by %v users, not stopping", localNodeInfo.UseCount)
	}

	instance.Store(nil)
}

func (in *LocalIotaNode) ISCPackageID() iotago.PackageID {
	return in.config.IscPackageID
}

func (in *LocalIotaNode) APIURL() string {
	return fmt.Sprintf("%s:%d", in.config.Host, in.config.Ports.RPC)
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

func (in *LocalIotaNode) waitAllHealthy(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(in.ctx, timeout)
	defer cancel()

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
		res, err := in.L1Client().GetLatestIotaSystemState(in.ctx)
		if err != nil {
			in.logf("StatusLoop: err: %s", err)
		}
		if err != nil || res == nil {
			return false
		}
		if res.PendingActiveValidatorsSize.Uint64() != 0 {
			return false
		}
		return true
	})

	tryLoop(func() bool {
		err := iotaclient.RequestFundsFromFaucet(ctx, in.iscPackageOwner.Address(), in.FaucetURL())
		if err != nil {
			in.logf("FaucetLoop: err: %s", err)
		}
		return err == nil
	})

	in.logf("Waiting until LocalIotaNode becomes ready... done! took: %v", time.Since(ts).Truncate(time.Millisecond))
}

func (in *LocalIotaNode) logf(msg string, args ...any) {
	if in.Logger != nil {
		in.Logger.Printf("Iota Node: "+msg+"\n", args...)
	}
}

var (
	runningKeepalives    = make(map[string]struct{})
	runningKeepaliveLock sync.Mutex
)

const (
	keepaliveInterval = time.Second
)

func startLocalNodeKeepalive(nodeInfoFile string) {
	runningKeepaliveLock.Lock()
	defer runningKeepaliveLock.Unlock()

	if _, ok := runningKeepalives[nodeInfoFile]; ok {
		return
	}

	runningKeepalives[nodeInfoFile] = struct{}{}

	go func() {
		ticker := time.NewTicker(keepaliveInterval)
		defer ticker.Stop()

		for {
			func() {
				<-ticker.C

				runningKeepaliveLock.Lock()
				defer runningKeepaliveLock.Unlock()

				if _, ok := runningKeepalives[nodeInfoFile]; !ok {
					return
				}

				lo.Must0(os.Chtimes(nodeInfoFile, time.Time{}, time.Now()))
			}()
		}
	}()
}

func stopLocalNodeKeepalive(nodeInfoFile string) {
	runningKeepaliveLock.Lock()
	defer runningKeepaliveLock.Unlock()

	delete(runningKeepalives, nodeInfoFile)
}
