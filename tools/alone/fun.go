package alone

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/waspconn"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/plugins/wasmtimevm"
	"github.com/stretchr/testify/require"
	"io/ioutil"
)

//goland:noinspection ALL
func (e *aloneEnvironment) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Chain ID: %s\n", e.ChainID.String())
	fmt.Fprintf(&buf, "Chain address: %s\n", e.ChainAddress.String())
	fmt.Fprintf(&buf, "State hash: %s\n", e.State.Hash().String())
	fmt.Fprintf(&buf, "UTXODB genesis address: %s\n", e.UtxoDB.GetGenesisAddress().String())
	return string(buf.Bytes())
}

func (e *aloneEnvironment) Infof(format string, args ...interface{}) {
	e.Log.Infof(format, args...)
}

// NewSigScheme generates new ed25519 sigscheme and requests funds from the faucet
func (e *aloneEnvironment) NewSigScheme() signaturescheme.SignatureScheme {
	ret := signaturescheme.ED25519(ed25519.GenerateKeyPair())
	_, err := e.UtxoDB.RequestFunds(ret.Address())
	require.NoError(e.T, err)
	e.CheckBalance(ret.Address(), balance.ColorIOTA, testutil.RequestFundsAmount)
	return ret
}

func (e *aloneEnvironment) GetBalance(addr address.Address, col balance.Color) int64 {
	bals := e.GetColoredBalances(addr)
	ret, _ := bals[col]
	return ret
}

func (e *aloneEnvironment) GetColoredBalances(addr address.Address) map[balance.Color]int64 {
	outs := e.UtxoDB.GetAddressOutputs(addr)
	ret, _ := waspconn.OutputBalancesByColor(outs)
	return ret
}

func (e *aloneEnvironment) FindContract(name string) (*root.ContractRecord, error) {
	req := NewCall(root.Interface.Name, root.FuncFindContract).
		WithParams(root.ParamHname, coretypes.Hn(name))
	retDict, err := e.CallView(req)
	if err != nil {
		return nil, err
	}
	retBin, err := retDict.Get(root.ParamData)
	if err != nil {
		return nil, err
	}
	if retBin == nil {
		return nil, fmt.Errorf("conract '%s' not found", name)
	}
	return root.DecodeContractRecord(retBin)
}

func (e *aloneEnvironment) UploadBlob(sigScheme signaturescheme.SignatureScheme, params ...interface{}) (ret hashing.HashValue, err error) {
	par := toMap(params...)
	expectedHash := blob.MustGetBlobHash(codec.MakeDict(par))

	req := NewCall(blob.Interface.Name, blob.FuncStoreBlob).WithParams(params...)
	var res dict.Dict
	var resBin []byte
	res, err = e.PostRequest(req, sigScheme)
	if err != nil {
		return
	}
	resBin = res.MustGet(blob.ParamHash)
	var r *hashing.HashValue
	var ok bool
	r, ok, err = codec.DecodeHashValue(resBin)
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("interbal error: no hash returned")
		return
	}
	ret = *r
	require.EqualValues(e.T, expectedHash, ret)
	return
}

func (e *aloneEnvironment) UploadWasm(sigScheme signaturescheme.SignatureScheme, binaryCode []byte) (ret hashing.HashValue, err error) {
	return e.UploadBlob(sigScheme,
		blob.VarFieldVMType, wasmtimevm.VMType,
		blob.VarFieldProgramBinary, binaryCode,
	)
}

func (e *aloneEnvironment) UploadWasmFromFile(sigScheme signaturescheme.SignatureScheme, fname string) (ret hashing.HashValue, err error) {
	var binary []byte
	binary, err = ioutil.ReadFile(fname)
	if err != nil {
		return
	}
	return e.UploadWasm(sigScheme, binary)
}

func (e *aloneEnvironment) GetWasmBinary(progHash hashing.HashValue) ([]byte, error) {
	reqVmtype := NewCall(blob.Interface.Name, blob.FuncGetBlobField).
		WithParams(
			blob.ParamHash, progHash,
			blob.ParamField, blob.VarFieldVMType,
		)
	res, err := e.CallView(reqVmtype)
	if err != nil {
		return nil, err
	}
	require.EqualValues(e.T, wasmtimevm.VMType, string(res.MustGet(blob.ParamBytes)))

	reqBin := NewCall(blob.Interface.Name, blob.FuncGetBlobField).
		WithParams(
			blob.ParamHash, progHash,
			blob.ParamField, blob.VarFieldProgramBinary,
		)
	res, err = e.CallView(reqBin)
	if err != nil {
		return nil, err
	}
	binary := res.MustGet(blob.ParamBytes)
	return binary, nil
}

func (e *aloneEnvironment) DeployContract(sigScheme signaturescheme.SignatureScheme, name string, progHash hashing.HashValue) error {
	req := NewCall(root.Interface.Name, root.FuncDeployContract).
		WithParams(
			root.ParamProgramHash, progHash,
			root.ParamName, name,
		)
	_, err := e.PostRequest(req, sigScheme)
	return err
}

func (e *aloneEnvironment) DeployWasmContract(sigScheme signaturescheme.SignatureScheme, name string, fname string) error {
	hprog, err := e.UploadWasmFromFile(sigScheme, fname)
	if err != nil {
		return err
	}
	return e.DeployContract(sigScheme, name, hprog)
}
