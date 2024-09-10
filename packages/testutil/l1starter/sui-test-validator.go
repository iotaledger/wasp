// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// l1starter allows starting and stopping the sui-test-validator tool
// for testing purposes.
package l1starter

import (
	"bufio"
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/sui-go/contracts"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suiconn"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
	"github.com/iotaledger/wasp/sui-go/suisigner"
)

var (
	ISCPackageOwner suisigner.Signer
	instance        atomic.Pointer[SuiTestValidator]
)

func init() {
	var seed [ed25519.SeedSize]byte
	copy(seed[:], []byte("iscPackageOwner"))
	ISCPackageOwner = suisigner.NewSigner(seed[:], suisigner.KeySchemeFlagDefault)
}

func Instance() *SuiTestValidator {
	stv := instance.Load()
	if stv == nil {
		panic("SuiTestValidator not started; call Start() first")
	}
	return stv
}

func ISCPackageID() sui.PackageID {
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

type SuiTestValidator struct {
	ctx          context.Context
	Config       Config
	Cmd          *exec.Cmd
	ISCPackageID sui.PackageID
}

func Start(ctx context.Context, cfg Config) *SuiTestValidator {
	stv := &SuiTestValidator{Config: cfg}
	if !instance.CompareAndSwap(nil, stv) {
		panic("an instance of sui-test-validator is already running")
	}
	stv.start(ctx)
	return stv
}

func (stv *SuiTestValidator) start(ctx context.Context) {
	stv.ctx = ctx
	stv.logf("Starting sui-test-validator...")
	ts := time.Now()
	stv.execCmd()
	stv.logf("Starting sui-test-validator... done! took: %v", time.Since(ts).Truncate(time.Millisecond))
	stv.waitAllHealthy(5 * time.Minute)
	stv.logf("Deploying ISC contracts...")
	stv.ISCPackageID = stv.deployISCContracts()
	stv.logf("SuiTestValidator started successfully")
}

func (stv *SuiTestValidator) execCmd() {
	// using CommandContext so that the sui process is killed when the
	// ctx is done
	testValidatorCmd := exec.CommandContext(
		stv.ctx,
		"sui-test-validator",
		fmt.Sprintf("--epoch-duration-ms=%d", 60000),
		fmt.Sprintf("--fullnode-rpc-port=%d", stv.Config.RPCPort),
		fmt.Sprintf("--faucet-port=%d", stv.Config.FaucetPort),
	)
	// also kill the sui process if the go process dies
	util.TerminateCmdWhenTestStops(testValidatorCmd)

	testValidatorCmd.Env = os.Environ()
	stv.Cmd = testValidatorCmd
	stv.redirectOutput()
	lo.Must0(testValidatorCmd.Start())
}

func (stv *SuiTestValidator) Stop() {
	stv.logf("Stopping...")

	if err := stv.Cmd.Process.Signal(syscall.SIGTERM); err != nil {
		panic(fmt.Errorf("unable to send TERM signal to sui-test-validator: %w", err))
	}

	if err := stv.Cmd.Wait(); err != nil {
		var errCode *exec.ExitError
		ok := errors.As(err, &errCode)

		if ok && strings.Contains(errCode.Error(), "terminated") {
			stv.logf("Stopping... Done")
			return
		}

		panic(fmt.Errorf("SUI node failed: %s", stv.Cmd.ProcessState.String()))
	}

	if !instance.CompareAndSwap(stv, nil) {
		panic("should not happen")
	}
}

func (stv *SuiTestValidator) Client() *suiclient.Client {
	return suiclient.NewHTTP(fmt.Sprintf("%s:%d", stv.Config.Host, stv.Config.RPCPort))
}

func (stv *SuiTestValidator) waitAllHealthy(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(stv.ctx, timeout)
	defer cancel()

	ts := time.Now()
	stv.logf("Waiting for all SUI nodes to become healthy...")

	tryLoop := func(f func() bool) {
		for {
			if ctx.Err() != nil {
				panic("nodes didn't become healthy in time")
			}
			if f() {
				return
			}
			stv.logf("Waiting until sui-test-validator becomes ready. Time waiting: %v", time.Since(ts).Truncate(time.Millisecond))
			time.Sleep(100 * time.Millisecond)
		}
	}

	tryLoop(func() bool {
		res, err := stv.Client().GetLatestSuiSystemState(stv.ctx)
		if err != nil || res == nil {
			return false
		}
		if res.PendingActiveValidatorsSize.Uint64() != 0 {
			return false
		}
		return true
	})

	tryLoop(func() bool {
		err := suiclient.RequestFundsFromFaucet(ctx, ISCPackageOwner.Address(), suiconn.LocalnetFaucetURL)
		return err == nil
	})

	stv.logf("Waiting until sui-test-validator becomes ready... done! took: %v", time.Since(ts).Truncate(time.Millisecond))
}

func (stv *SuiTestValidator) logf(msg string, args ...any) {
	if stv.Config.LogFunc != nil {
		stv.Config.LogFunc("SuiTestValidator: "+msg, args...)
	}
}

func (stv *SuiTestValidator) redirectOutput() {
	stdout := lo.Must(stv.Cmd.StdoutPipe())
	stderr := lo.Must(stv.Cmd.StderrPipe())

	outFilePath := filepath.Join(os.TempDir(), "sui-test-validator-stdout.log")
	outFile := lo.Must(os.Create(outFilePath))
	go scanLog(stdout, outFile)

	errFilePath := filepath.Join(os.TempDir(), "sui-test-validator-stderr.log")
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

func (stv *SuiTestValidator) deployISCContracts() sui.PackageID {
	client := stv.Client()
	iscBytecode := contracts.ISC()
	txnBytes := lo.Must(client.Publish(context.Background(), suiclient.PublishRequest{
		Sender:          ISCPackageOwner.Address(),
		CompiledModules: iscBytecode.Modules,
		Dependencies:    iscBytecode.Dependencies,
		GasBudget:       suijsonrpc.NewBigInt(suiclient.DefaultGasBudget * 10),
	}))
	txnResponse := lo.Must(client.SignAndExecuteTransaction(
		context.Background(),
		ISCPackageOwner,
		txnBytes.TxBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	))
	if !txnResponse.Effects.Data.IsSuccess() {
		panic("publish ISC contracts failed")
	}
	packageID := lo.Must(txnResponse.GetPublishedPackageID())
	return *packageID
}
