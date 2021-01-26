package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/iotaledger/wasp/tools/cluster/testutil"
	"github.com/stretchr/testify/require"
)

type WaspCliTest struct {
	t   *testing.T
	clu *cluster.Cluster
	dir string
}

func NewWaspCliTest(t *testing.T) *WaspCliTest {
	clu := testutil.NewCluster(t)

	dir, err := ioutil.TempDir(os.TempDir(), "wasp-cli-test")
	t.Logf("Using temporary directory %s", dir)
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	w := &WaspCliTest{
		t:   t,
		clu: clu,
		dir: dir,
	}
	w.Run("set", "utxodb", "true")
	return w
}

func (w *WaspCliTest) runCmd(args []string, f func(*exec.Cmd)) []string {
	// -w: wait for requests
	// -d: debug output
	cmd := exec.Command("wasp-cli", append([]string{"-w", "-d"}, args...)...)
	cmd.Dir = w.dir

	stdout := &bytes.Buffer{}
	cmd.Stdout = stdout
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr

	if f != nil {
		f(cmd)
	}

	err := cmd.Run()

	outStr, errStr := stdout.String(), stderr.String()
	if err != nil {
		require.NoError(w.t, fmt.Errorf(
			"cmd `wasp-cli %s` failed\n%w\noutput:\n%s",
			strings.Join(args, " "),
			err,
			outStr+errStr,
		))
	}
	outStr = strings.Replace(outStr, "\r", "", -1)
	outStr = strings.TrimRight(outStr, "\n")
	return strings.Split(outStr, "\n")
}

func (w *WaspCliTest) Run(args ...string) []string {
	return w.runCmd(args, nil)
}

func (w *WaspCliTest) Pipe(in []string, args ...string) []string {
	return w.runCmd(args, func(cmd *exec.Cmd) {
		cmd.Stdin = bytes.NewReader([]byte(strings.Join(in, "\n")))
	})
}

func TestNoChains(t *testing.T) {
	w := NewWaspCliTest(t)

	w.Run("init")
	w.Run("request-funds")

	out := w.Run("address")

	ownerAddr := regexp.MustCompile(`(?m)Address:[[:space:]]+([[:alnum:]]+)$`).FindStringSubmatch(out[1])[1]
	require.NotEmpty(t, ownerAddr)

	out = w.Run("chain", "list")
	require.Contains(t, out[0], "Total 0 chain(s)")
}

func Test1Chain(t *testing.T) {
	w := NewWaspCliTest(t)

	w.Run("init")
	w.Run("request-funds")

	out := w.Run("address")
	ownerAddr := regexp.MustCompile(`(?m)Address:[[:space:]]+([[:alnum:]]+)$`).FindStringSubmatch(out[1])[1]
	require.NotEmpty(t, ownerAddr)
	t.Logf("Owner address: %s", ownerAddr)

	alias := "chain1"

	// test chain deploy command
	w.Run("chain", "deploy", "--chain="+alias, "--committee=0,1,2,3", "--quorum=3")

	// test chain info command
	out = w.Run("chain", "info")
	chainID := regexp.MustCompile(`(?m)Chain ID:[[:space:]]+([[:alnum:]]+)$`).FindStringSubmatch(out[0])[1]
	require.NotEmpty(t, chainID)
	t.Logf("Chain ID: %s", chainID)

	// test chain list command
	out = w.Run("chain", "list")
	require.Contains(t, out[0], "Total 1 chain(s)")
	require.Contains(t, out[4], chainID)

	// unnecessary, since it is the latest deployed chain
	w.Run("set", "chain", alias)

	// test chain list-contracts command
	out = w.Run("chain", "list-contracts")
	require.Contains(t, out[0], "Total 4 contracts")

	// test chain list-accounts command
	out = w.Run("chain", "list-accounts")
	require.Contains(t, out[0], "Total 1 account(s)")
	agentID := strings.TrimSpace(out[4])
	require.NotEmpty(t, agentID)
	t.Logf("Agent ID: %s", agentID)

	// test chain balance command
	out = w.Run("chain", "balance", agentID)
	// check that the chain balance of owner is 1 IOTA
	require.Regexp(t, "(?m)IOTA[[:space:]]+1$", out[3])

	// same test, this time calling the view function manually
	out = w.Run("chain", "call-view", "accounts", "balance", "string", "a", "agentid", agentID)
	out = w.Pipe(out, "decode", "color", "int")
	require.Regexp(t, "(?m)IOTA:[[:space:]]+1$", out[0])

	// test the chainlog
	out = w.Run("chain", "log", "root")
	require.Len(t, out, 1)
}
