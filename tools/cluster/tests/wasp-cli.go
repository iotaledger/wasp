package tests

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/tools/cluster"
)

type WaspCLITest struct {
	T              *testing.T
	Cluster        *cluster.Cluster
	dir            string
	WaspCliAddress iotago.Address
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
	w.MustRun("init")

	w.MustRun("set", "l1.apiAddress", clu.Config.L1.APIAddress)
	w.MustRun("set", "l1.faucetAddress", clu.Config.L1.FaucetAddress)
	for _, node := range clu.Config.AllNodes() {
		w.MustRun("wasp", "add", fmt.Sprintf("%d", node), clu.Config.APIHost(node))
	}

	requestFundstext := w.MustRun("request-funds")
	// regex example: Request funds for address atoi1qqqrqtn44e0563utwau9aaygt824qznjkhvr6836eratglg3cp2n6ydplqx: success
	expectedRegexp := regexp.MustCompile(`(?i:Request funds for address)\s*([a-z]{1,4}1[a-z0-9]{59}).*(?i:success)`)
	rs := expectedRegexp.FindStringSubmatch(requestFundstext[len(requestFundstext)-1])
	require.Len(t, rs, 2)
	_, addr, err := iotago.ParseBech32(rs[1])
	require.NoError(t, err)
	w.WaspCliAddress = addr
	// requested funds will take some time to be available
	for {
		outputs, err := clu.L1Client().OutputMap(addr)
		require.NoError(t, err)
		if len(outputs) > 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	return w
}

func (w *WaspCLITest) runCmd(args []string, f func(*exec.Cmd)) ([]string, error) {
	// -w: wait for requests
	// -d: debug output
	cmd := exec.Command("wasp-cli", append([]string{"-w", "-d"}, args...)...) //nolint:gosec
	cmd.Dir = w.dir

	stdout := &bytes.Buffer{}
	cmd.Stdout = stdout
	stderr := &bytes.Buffer{}
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
	return w.runCmd(args, nil)
}

func (w *WaspCLITest) MustRun(args ...string) []string {
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
	r := regexp.MustCompile(`(.*)\(check result with:\s*wasp-cli (.*)\).*$`).
		FindStringSubmatch(strings.Join(out, ""))
	command := r[2]
	return w.MustRun(strings.Split(command, " ")...)
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

func (w *WaspCLITest) AllNodesArg() string {
	var nodes []string
	for i := 0; i < len(w.Cluster.Config.Wasp); i++ {
		nodes = append(nodes, fmt.Sprintf("%d", i))
	}
	return "--nodes=" + strings.Join(nodes, ",")
}

func (w *WaspCLITest) CommitteeConfigArgs() (string, string) {
	quorum := 3 * len(w.Cluster.Config.Wasp) / 4
	if quorum < 1 {
		quorum = 1
	}

	return w.AllNodesArg(), fmt.Sprintf("--quorum=%d", quorum)
}

func (w *WaspCLITest) Address() iotago.Address {
	out := w.MustRun("address")
	s := regexp.MustCompile(`(?m)Address:[[:space:]]+([[:alnum:]]+)$`).FindStringSubmatch(out[1])[1] //nolint:gocritic
	_, addr, err := iotago.ParseBech32(s)
	require.NoError(w.T, err)
	return addr
}
