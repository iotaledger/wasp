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
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/plugins/wasmtimevm"
	"github.com/stretchr/testify/require"
	"io/ioutil"
)

//goland:noinspection ALL
func (e *AloneEnvironment) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Chain ID: %s\n", e.ChainID.String())
	fmt.Fprintf(&buf, "Chain address: %s\n", e.ChainAddress.String())
	fmt.Fprintf(&buf, "State hash: %s\n", e.State.Hash().String())
	fmt.Fprintf(&buf, "UTXODB genesis address: %s\n", e.UtxoDB.GetGenesisAddress().String())
	return string(buf.Bytes())
}

func (e *AloneEnvironment) Infof(format string, args ...interface{}) {
	e.Log.Infof(format, args...)
}

// NewSigScheme generates new ed25519 sigscheme and requests funds from the faucet
func (e *AloneEnvironment) NewSigScheme() signaturescheme.SignatureScheme {
	ret := signaturescheme.ED25519(ed25519.GenerateKeyPair())
	_, err := e.UtxoDB.RequestFunds(ret.Address())
	require.NoError(e.T, err)
	e.CheckUtxodbBalance(ret.Address(), balance.ColorIOTA, testutil.RequestFundsAmount)
	return ret
}

func (e *AloneEnvironment) FindContract(name string) (*root.ContractRecord, error) {
	req := NewCall(root.Interface.Name, root.FuncFindContract, root.ParamHname, coretypes.Hn(name))
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

func (e *AloneEnvironment) UploadBlob(sigScheme signaturescheme.SignatureScheme, params ...interface{}) (ret hashing.HashValue, err error) {
	par := toMap(params...)
	expectedHash := blob.MustGetBlobHash(codec.MakeDict(par))

	req := NewCall(blob.Interface.Name, blob.FuncStoreBlob, params...)
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

func (e *AloneEnvironment) UploadWasm(sigScheme signaturescheme.SignatureScheme, binaryCode []byte) (ret hashing.HashValue, err error) {
	return e.UploadBlob(sigScheme,
		blob.VarFieldVMType, wasmtimevm.VMType,
		blob.VarFieldProgramBinary, binaryCode,
	)
}

func (e *AloneEnvironment) UploadWasmFromFile(sigScheme signaturescheme.SignatureScheme, fname string) (ret hashing.HashValue, err error) {
	var binary []byte
	binary, err = ioutil.ReadFile(fname)
	if err != nil {
		return
	}
	return e.UploadWasm(sigScheme, binary)
}

func (e *AloneEnvironment) GetWasmBinary(progHash hashing.HashValue) ([]byte, error) {
	reqVmtype := NewCall(blob.Interface.Name, blob.FuncGetBlobField,
		blob.ParamHash, progHash,
		blob.ParamField, blob.VarFieldVMType,
	)
	res, err := e.CallView(reqVmtype)
	if err != nil {
		return nil, err
	}
	require.EqualValues(e.T, wasmtimevm.VMType, string(res.MustGet(blob.ParamBytes)))

	reqBin := NewCall(blob.Interface.Name, blob.FuncGetBlobField,
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

func (e *AloneEnvironment) DeployContract(sigScheme signaturescheme.SignatureScheme, name string, progHash hashing.HashValue, params ...interface{}) error {
	par := []interface{}{root.ParamProgramHash, progHash, root.ParamName, name}
	par = append(par, params...)
	req := NewCall(root.Interface.Name, root.FuncDeployContract, par...)
	_, err := e.PostRequest(req, sigScheme)
	return err
}

func (e *AloneEnvironment) DeployWasmContract(sigScheme signaturescheme.SignatureScheme, name string, fname string, params ...interface{}) error {
	hprog, err := e.UploadWasmFromFile(sigScheme, fname)
	if err != nil {
		return err
	}
	return e.DeployContract(sigScheme, name, hprog, params...)
}

func (e *AloneEnvironment) GetInfo() (coretypes.ChainID, coretypes.AgentID, map[coretypes.Hname]*root.ContractRecord) {
	req := NewCall(root.Interface.Name, root.FuncGetInfo)
	res, err := e.CallView(req)
	require.NoError(e.T, err)

	chainID, ok, err := codec.DecodeChainID(res.MustGet(root.VarChainID))
	require.NoError(e.T, err)
	require.True(e.T, ok)

	chainOwnerID, ok, err := codec.DecodeAgentID(res.MustGet(root.VarChainOwnerID))
	require.NoError(e.T, err)
	require.True(e.T, ok)

	contracts, err := root.DecodeContractRegistry(datatypes.NewMustMap(res, root.VarContractRegistry))
	require.NoError(e.T, err)
	return chainID, chainOwnerID, contracts
}

func (e *AloneEnvironment) GetUtxodbBalance(addr address.Address, col balance.Color) int64 {
	bals := e.GetUtxodbBalances(addr)
	ret, _ := bals[col]
	return ret
}

func (e *AloneEnvironment) GetUtxodbBalances(addr address.Address) map[balance.Color]int64 {
	outs := e.UtxoDB.GetAddressOutputs(addr)
	ret, _ := waspconn.OutputBalancesByColor(outs)
	return ret
}

func (e *AloneEnvironment) GetAccounts() []coretypes.AgentID {
	req := NewCall(accountsc.Interface.Name, accountsc.FuncAccounts)
	d, err := e.CallView(req)
	require.NoError(e.T, err)
	keys := d.KeysSorted()
	ret := make([]coretypes.AgentID, 0, len(keys)-1)
	for _, key := range keys {
		aid, ok, err := codec.DecodeAgentID([]byte(key))
		require.NoError(e.T, err)
		require.True(e.T, ok)
		if aid == accountsc.TotalAssetsAccountID {
			continue
		}
		ret = append(ret, aid)
	}
	return ret
}

func (e *AloneEnvironment) GetAccountBalance(agentID coretypes.AgentID) coretypes.ColoredBalances {
	req := NewCall(accountsc.Interface.Name, accountsc.FuncBalance, accountsc.ParamAgentID, agentID)
	d, err := e.CallView(req)
	require.NoError(e.T, err)
	if d.IsEmpty() {
		return cbalances.Nil
	}
	ret := make(map[balance.Color]int64)
	err = d.Iterate("", func(key kv.Key, value []byte) bool {
		col, _, err := codec.DecodeColor([]byte(key))
		require.NoError(e.T, err)
		val, _, err := codec.DecodeInt64(value)
		require.NoError(e.T, err)
		ret[*col] = val
		return true
	})
	require.NoError(e.T, err)
	return cbalances.NewFromMap(ret)
}

func (e *AloneEnvironment) GetTotalAssets() coretypes.ColoredBalances {
	return e.GetAccountBalance(accountsc.TotalAssetsAccountID)
}
