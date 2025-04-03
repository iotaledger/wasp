package tests

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/tools/cluster"
)

type WaspCLITest struct {
	T              *testing.T
	Cluster        *cluster.Cluster
	dir            string
	WaspCliAddress *cryptolib.Address
}

func newWaspCLITest(t *testing.T, opt ...waspClusterOpts) *WaspCLITest {
	clu := newCluster(t, opt...)

	dir, err := os.MkdirTemp(os.TempDir(), "wasp-cli-test-*")
	t.Logf("Using temporary directory %s", dir)
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	w := &WaspCLITest{
		T:       t,
		Cluster: clu,
		dir:     dir,
	}
	w.MustRun("wallet-provider", "unsafe_inmemory_testing_seed")
	w.MustRun("init")

	// FIXME make them into parameters
	w.MustRun("set", "l1.apiAddress", clu.Config.L1APIAddress())
	w.MustRun("set", "l1.faucetAddress", clu.Config.L1FaucetAddress())
	w.MustRun("set", "l1.packageId", clu.Config.ISCPackageID().String())
	for _, node := range clu.Config.AllNodes() {
		w.MustRun("wasp", "add", fmt.Sprintf("%d", node), clu.Config.APIHost(node))
	}

	requestFundstext := w.MustRun("request-funds")
	// regex example: Request funds for address atoi1qqqrqtn44e0563utwau9aaygt824qznjkhvr6836eratglg3cp2n6ydplqx: success
	expectedRegexp := regexp.MustCompile(`(?i:Request funds for address)\s*(0x[a-fA-F0-9]{40}).*(?i:success)`)
	rs := expectedRegexp.FindStringSubmatch(requestFundstext[len(requestFundstext)-1])
	require.Len(t, rs, 2)
	addr, err := cryptolib.AddressFromHex(rs[1])
	require.NoError(t, err)
	w.WaspCliAddress = addr
	return w
}

func (w *WaspCLITest) runCmd(args []string, f func(*exec.Cmd)) ([]string, error) {
	w.T.Helper()
	// -w: wait for requests
	// -d: debug output
	cmd := exec.Command("wasp-cli", append([]string{"-c", w.dir + "/wasp-cli.json", "-w", "-d"}, args...)...) //nolint:gosec
	cmd.Dir = w.dir

	stdout := new(bytes.Buffer)
	cmd.Stdout = stdout
	stderr := new(bytes.Buffer)
	cmd.Stderr = stderr

	if f != nil {
		f(cmd)
	}

	w.T.Logf("Running: %s", strings.Join(cmd.Args, " "))
	err := cmd.Run()

	outStr, errStr := stdout.String(), stderr.String()
	if err != nil {
		return nil, fmt.Errorf(
			"cmd `wasp-cli %s` failed\n%w\noutput:\n%s",
			strings.Join(args, " "),
			err,
			outStr+errStr,
		)
	}
	outStr = strings.Replace(outStr, "\r", "", -1)
	outStr = strings.TrimRight(outStr, "\n")
	outLines := strings.Split(outStr, "\n")
	w.T.Logf("OUTPUT #lines=%v", len(outLines))
	for _, outLine := range outLines {
		w.T.Logf("OUTPUT: %v", outLine)
	}
	return outLines, nil
}

func (w *WaspCLITest) Run(args ...string) ([]string, error) {
	w.T.Helper()
	return w.runCmd(args, nil)
}

func (w *WaspCLITest) MustRun(args ...string) []string {
	w.T.Helper()
	lines, err := w.Run(args...)
	if err != nil {
		panic(err)
	}
	return lines
}

func (w *WaspCLITest) PostRequestGetReceipt(args ...string) []string {
	runArgs := []string{"chain", "post-request", "-s"}
	runArgs = append(runArgs, args...)
	out := w.MustRun(runArgs...)
	return w.GetReceiptFromRunPostRequestOutput(out)
}

func (w *WaspCLITest) GetReceiptFromRunPostRequestOutput(out []string) []string {
	r := regexp.MustCompile(`(.*)\(check result with:\s*wasp-cli (.*?)\)`).
		FindStringSubmatch(strings.Join(out, ""))
	checkReceiptCommand := strings.Split(r[2], " ")
	checkReceiptCommand = append(checkReceiptCommand, "--node=0")
	return w.MustRun(checkReceiptCommand...)
}

func (w *WaspCLITest) Pipe(in []string, args ...string) ([]string, error) {
	return w.runCmd(args, func(cmd *exec.Cmd) {
		cmd.Stdin = bytes.NewReader([]byte(strings.Join(in, "\n")))
	})
}

func (w *WaspCLITest) MustPipe(in []string, args ...string) []string {
	lines, err := w.Pipe(in, args...)
	if err != nil {
		panic(err)
	}
	return lines
}

// CopyFile copies the given file into the temp directory
func (w *WaspCLITest) CopyFile(srcFile string) {
	source, err := os.Open(srcFile)
	require.NoError(w.T, err)
	defer source.Close()

	dst := path.Join(w.dir, path.Base(srcFile))
	destination, err := os.Create(dst)
	require.NoError(w.T, err)
	defer destination.Close()

	_, err = io.Copy(destination, source)
	require.NoError(w.T, err)
}

func (w *WaspCLITest) ArgAllNodesExcept(idx int) string {
	var nodes []string
	for i := 0; i < len(w.Cluster.Config.Wasp); i++ {
		if i != idx {
			nodes = append(nodes, fmt.Sprintf("%d", i))
		}
	}
	return "--peers=" + strings.Join(nodes, ",")
}

func (w *WaspCLITest) ArgCommitteeConfig(initiatorIndex int) (string, string) {
	quorum := 3 * len(w.Cluster.Config.Wasp) / 4
	if quorum < 1 {
		quorum = 1
	}

	return w.ArgAllNodesExcept(initiatorIndex), fmt.Sprintf("--quorum=%d", quorum)
}

func (w *WaspCLITest) Address() iotago.Address {
	out := w.MustRun("address")
	s := regexp.MustCompile(`(?m)Address:[[:space:]]+([[:alnum:]]+)$`).FindStringSubmatch(out[1])[1] //nolint:gocritic
	addr, err := iotago.AddressFromHex(s)
	require.NoError(w.T, err)
	return *addr
}

// TODO there is a small issue if we try to activate the chain twice (deploy command also activates the chain)
// if this happens, the node will return an error on `getChainInfo` because there is no state yet.
// as a temporary fix, we add `skipOnNodes`, so to not run the activate command on that node
func (w *WaspCLITest) ActivateChainOnAllNodes(chainName string, skipOnNodes ...int) {
	for _, idx := range w.Cluster.AllNodes() {
		if !slices.Contains(skipOnNodes, idx) {
			w.MustRun("chain", "activate", "--chain="+chainName, fmt.Sprintf("--node=%d", idx))
		}
	}

	// Hack to get the chainID that was deployed

	chainIsUpAndRunning := func(t *testing.T, nodeIndex int) bool {
		_, _, err := w.Cluster.WaspClient(nodeIndex).ChainsAPI.
			CallView(context.Background()).
			ContractCallViewRequest(apiclient.ContractCallViewRequest{
				ContractHName: governance.Contract.Hname().String(),
				FunctionHName: governance.ViewGetChainInfo.Hname().String(),
			}).
			Execute()
		return err == nil
	}
	// wait until the chain is synced on the nodes, otherwise we get a race condition on the next test commands
	waitUntil(w.T, chainIsUpAndRunning, w.Cluster.AllNodes(), 30*time.Second)
}

func (w *WaspCLITest) CreateL2NativeToken(tokenScheme isc.SimpleTokenScheme, tokenName string, tokenSymbol string, tokenDecimals uint8) {
	panic("refactor me: support native token creation")
}

func (w *WaspCLITest) ChainID(idx int) string {
	out := w.MustRun("chain", "info", fmt.Sprintf("--node=%d", idx))
	return regexp.MustCompile(`(?m)Chain ID:\s+([[:alnum:]]+)$`).FindStringSubmatch(out[0])[1]
}
