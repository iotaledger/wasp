package testutil

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

type WaspCLITest struct {
	t   *testing.T
	clu *cluster.Cluster
	dir string
}

func NewWaspCLITest(t *testing.T) *WaspCLITest {
	clu := NewCluster(t)

	dir, err := ioutil.TempDir(os.TempDir(), "wasp-cli-test-*")
	t.Logf("Using temporary directory %s", dir)
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	w := &WaspCLITest{
		t:   t,
		clu: clu,
		dir: dir,
	}
	w.Run("set", "utxodb", "true")
	return w
}

func (w *WaspCLITest) runCmd(args []string, f func(*exec.Cmd)) []string {
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

func (w *WaspCLITest) Run(args ...string) []string {
	return w.runCmd(args, nil)
}

func (w *WaspCLITest) Pipe(in []string, args ...string) []string {
	return w.runCmd(args, func(cmd *exec.Cmd) {
		cmd.Stdin = bytes.NewReader([]byte(strings.Join(in, "\n")))
	})
}

// CopyFile copies the given file into the temp directory
func (w *WaspCLITest) CopyFile(srcFile string) {
	source, err := os.Open(srcFile)
	require.NoError(w.t, err)
	defer source.Close()

	dst := path.Join(w.dir, path.Base(srcFile))
	destination, err := os.Create(dst)
	require.NoError(w.t, err)
	defer destination.Close()

	_, err = io.Copy(destination, source)
	require.NoError(w.t, err)
}

func (w *WaspCLITest) CommitteeConfig() (string, string) {
	var committee []string
	for i := 0; i < w.clu.Config.Wasp.NumNodes; i++ {
		committee = append(committee, fmt.Sprintf("%d", i))
	}

	quorum := 3 * w.clu.Config.Wasp.NumNodes / 4
	if quorum < 1 {
		quorum = 1
	}

	return "--committee=" + strings.Join(committee, ","), fmt.Sprintf("--quorum=%d", quorum)
}
