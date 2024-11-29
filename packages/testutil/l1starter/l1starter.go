// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// l1starter allows starting and stopping the iota validator tool
// for testing purposes.
package l1starter

import (
	"bufio"
	"context"
	"crypto/ed25519"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/packages/util"
)

var (
	ISCPackageOwner iotasigner.Signer
	instance        atomic.Pointer[IotaNode]
)

func init() {
	var seed [ed25519.SeedSize]byte
	copy(seed[:], []byte("iscPackageOwner"))
	ISCPackageOwner = iotasigner.NewSigner(seed[:], iotasigner.KeySchemeFlagDefault)
}

func Instance() *IotaNode {
	in := instance.Load()
	if in == nil {
		panic("IotaNode not started; call Start() first")
	}
	return in
}

var (
	iotaNode   *IotaNode
	serverOnce sync.Once
)

func getIotaNode() *IotaNode {
	serverOnce.Do(func() {
		iotaNode = Start(context.Background(), DefaultConfig)
	})
	return iotaNode
}

func TestMain(m *testing.M) {
	flag.Parse()
	getIotaNode()
	m.Run()
}

func SingleTest(t *testing.T) {
	if iotaNode == nil {
		flag.Parse()
		getIotaNode()
	}

	/*
		TODO: Find better solution than this, because it wont work with multiple tests.
		Maybe just dont worry, IOTA Node will be killed when all tests are done regardless, just not as gentle.
		t.Cleanup(func() {
			iotaNode.Stop()
		})*/
}

func ISCPackageID() iotago.PackageID {
	return Instance().ISCPackageID
}

type Config struct {
	Host       string
	RPCPort    int
	FaucetPort int
	LogFunc    LogFunc
}

var DefaultConfig = Config{
	Host:       "http://localhost",
	RPCPort:    9000,
	FaucetPort: 9123,
	LogFunc: func(format string, args ...interface{}) {
		fmt.Printf(format+"\n", args...)
	},
}

type LogFunc func(format string, args ...interface{})

type IotaNode struct {
	ctx          context.Context
	Config       Config
	Cmd          *exec.Cmd
	ISCPackageID iotago.PackageID
}

func Start(ctx context.Context, cfg Config) *IotaNode {
	in := &IotaNode{Config: cfg}
	if !instance.CompareAndSwap(nil, in) {
		panic("an instance of IotaNode is already running")
	}
	in.start(ctx)
	return in
}

func (in *IotaNode) start(ctx context.Context) {
	in.ctx = ctx
	var ts time.Time
	if runtime.GOOS == "darwin" {
		in.logf("Not run IotaNode by Go on MacOS")
	} else {
		in.logf("Starting IotaNode...")
		ts = time.Now()
		in.execCmd()
	}

	in.logf("Starting IotaNode... done! took: %v", time.Since(ts).Truncate(time.Millisecond))
	in.waitAllHealthy(5 * time.Minute)
	in.logf("Deploying ISC contracts...")

	packageID, err := in.Client().DeployISCContracts(ctx, ISCPackageOwner)
	if err != nil {
		panic(fmt.Errorf("isc contract deployment failed: %w", err))
	}

	in.ISCPackageID = packageID

	in.logf("IotaNode started successfully")
}

func (in *IotaNode) execCmd() {
	// using CommandContext so that the iotago process is killed when the
	// ctx is done
	testValidatorCmd := exec.CommandContext(
		in.ctx,
		"iota",
		"start",
		"--force-regenesis",
		fmt.Sprintf("--epoch-duration-ms=%d", 60000),
		fmt.Sprintf("--fullnode-rpc-port=%d", in.Config.RPCPort),
		fmt.Sprintf("--with-faucet=%d", in.Config.FaucetPort),
		fmt.Sprintf("--faucet-amount=%d", iotaclient.FundsFromFaucetAmount), // We need a lot. For now.
	)

	// also kill the iotago process if the go process dies
	util.TerminateCmdWhenTestStops(testValidatorCmd)

	testValidatorCmd.Env = os.Environ()
	in.Cmd = testValidatorCmd
	in.redirectOutput()

	lo.Must0(testValidatorCmd.Start())
}

func (in *IotaNode) Stop() {
	in.logf("Stopping...")
	if runtime.GOOS == "darwin" {
		return
	}
	if err := in.Cmd.Process.Signal(syscall.SIGTERM); err != nil {
		panic(fmt.Errorf("unable to send TERM signal to IotaNode: %w", err))
	}

	if err := in.Cmd.Wait(); err != nil {
		var errCode *exec.ExitError
		ok := errors.As(err, &errCode)

		if ok && strings.Contains(errCode.Error(), "terminated") {
			in.logf("Stopping... Done")
			return
		}

		panic(fmt.Errorf("IOTA node failed: %s", in.Cmd.ProcessState.String()))
	}

	if !instance.CompareAndSwap(in, nil) {
		panic("should not happen")
	}
}

func (in *IotaNode) Client() clients.L1Client {
	return clients.NewL1Client(clients.L1Config{
		APIURL:    fmt.Sprintf("%s:%d", in.Config.Host, in.Config.RPCPort),
		FaucetURL: fmt.Sprintf("%s:%d/gas", in.Config.Host, in.Config.FaucetPort),
	})
}

func (in *IotaNode) waitAllHealthy(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(in.ctx, timeout)
	defer cancel()

	ts := time.Now()
	in.logf("Waiting for all IOTA nodes to become healthy...")

	tryLoop := func(f func() bool) {
		for {
			if ctx.Err() != nil {
				panic("nodes didn't become healthy in time")
			}
			if f() {
				return
			}
			in.logf("Waiting until IotaNode becomes ready. Time waiting: %v", time.Since(ts).Truncate(time.Millisecond))
			time.Sleep(500 * time.Millisecond)
		}
	}

	tryLoop(func() bool {
		res, err := in.Client().GetLatestIotaSystemState(in.ctx)
		if err != nil {
			in.logf("err: %s", err)
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
		err := iotaclient.RequestFundsFromFaucet(ctx, ISCPackageOwner.Address(), iotaconn.LocalnetFaucetURL)
		if err != nil {
			in.logf("err: %s", err)
		}
		return err == nil
	})

	in.logf("Waiting until IotaNode becomes ready... done! took: %v", time.Since(ts).Truncate(time.Millisecond))
}

func (in *IotaNode) logf(msg string, args ...any) {
	if in.Config.LogFunc != nil {
		in.Config.LogFunc("IotaNode: "+msg, args...)
	}
}

func (in *IotaNode) redirectOutput() {
	stdout := lo.Must(in.Cmd.StdoutPipe())
	stderr := lo.Must(in.Cmd.StderrPipe())

	outFilePath := filepath.Join(os.TempDir(), "iota-validator-stdout.log")
	outFile := lo.Must(os.Create(outFilePath))
	go scanLog(stdout, outFile)

	errFilePath := filepath.Join(os.TempDir(), "iota-validator-stderr.log")
	errFile := lo.Must(os.Create(errFilePath))
	go scanLog(stderr, errFile)
}

func scanLog(reader io.Reader, out *os.File) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		_ = lo.Must(out.WriteString(fmt.Sprintln(line)))
	}
}
