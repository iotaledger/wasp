package alone

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBlob(t *testing.T) {
	al := New(t, false, true)
	binary := []byte("supposed to be wasm")
	hwasm, err := al.UploadWasm(nil, binary)
	require.NoError(t, err)

	binBack, err := al.GetWasmBinary(hwasm)
	require.NoError(t, err)

	require.EqualValues(t, binary, binBack)
}

const wasmFile = "../../../tools/cluster/tests/wasptest_new/wasm/inccounter_bg.wasm"

func TestDeploy(t *testing.T) {
	al := New(t, false, true)
	hwasm, err := al.UploadWasmFromFile(nil, wasmFile)
	require.NoError(t, err)

	err = al.DeployContract(nil, "testInccounter", hwasm)
	require.NoError(t, err)
}

func TestDeployWasm(t *testing.T) {
	al := New(t, false, true)
	err := al.DeployWasmContract(nil, "testInccounter", wasmFile)
	require.NoError(t, err)
}
