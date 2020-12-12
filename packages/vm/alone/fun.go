// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0
package alone

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/waspconn"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/plugins/wasmtimevm"
	"github.com/stretchr/testify/require"
	"io/ioutil"
)

//goland:noinspection ALL
func (ch *Chain) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Chain ID: %s\n", ch.ChainID.String())
	fmt.Fprintf(&buf, "Chain address: %s\n", ch.ChainAddress.String())
	fmt.Fprintf(&buf, "State hash: %s\n", ch.State.Hash().String())
	fmt.Fprintf(&buf, "UTXODB genesis address: %s\n", ch.Glb.utxoDB.GetGenesisAddress().String())
	return string(buf.Bytes())
}

func (ch *Chain) Infof(format string, args ...interface{}) {
	ch.Log.Infof(format, args...)
}

func (ch *Chain) FindContract(name string) (*root.ContractRecord, error) {
	retDict, err := ch.CallView(root.Interface.Name, root.FuncFindContract, root.ParamHname, coretypes.Hn(name))
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

func (ch *Chain) GetBlobInfo(blobHash hashing.HashValue) (map[string]int, bool) {
	res, err := ch.CallView(blob.Interface.Name, blob.FuncGetBlobInfo, blob.ParamHash, blobHash)
	require.NoError(ch.Glb.T, err)
	if res.IsEmpty() {
		return nil, false
	}
	ret := make(map[string]int)
	res.ForEach(func(key kv.Key, value []byte) bool {
		v, ok, err := codec.DecodeInt64(value)
		require.NoError(ch.Glb.T, err)
		require.True(ch.Glb.T, ok)
		ret[string(key)] = int(v)
		return true
	})
	return ret, true
}

func (ch *Chain) UploadBlob(sigScheme signaturescheme.SignatureScheme, params ...interface{}) (ret hashing.HashValue, err error) {
	par := toMap(params...)
	expectedHash := blob.MustGetBlobHash(codec.MakeDict(par))
	if _, ok := ch.GetBlobInfo(expectedHash); ok {
		// blob exists, return hash of existing
		return expectedHash, nil
	}
	var res dict.Dict
	res, err = ch.PostRequest(NewCall(blob.Interface.Name, blob.FuncStoreBlob, params...), sigScheme)
	require.NoError(ch.Glb.T, err)
	resBin := res.MustGet(blob.ParamHash)
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
	require.EqualValues(ch.Glb.T, expectedHash, ret)
	return
}

func (ch *Chain) UploadWasm(sigScheme signaturescheme.SignatureScheme, binaryCode []byte) (ret hashing.HashValue, err error) {
	return ch.UploadBlob(sigScheme,
		blob.VarFieldVMType, wasmtimevm.VMType,
		blob.VarFieldProgramBinary, binaryCode,
	)
}

func (ch *Chain) UploadWasmFromFile(sigScheme signaturescheme.SignatureScheme, fname string) (ret hashing.HashValue, err error) {
	var binary []byte
	binary, err = ioutil.ReadFile(fname)
	if err != nil {
		return
	}
	return ch.UploadWasm(sigScheme, binary)
}

func (ch *Chain) GetWasmBinary(progHash hashing.HashValue) ([]byte, error) {
	res, err := ch.CallView(blob.Interface.Name, blob.FuncGetBlobField,
		blob.ParamHash, progHash,
		blob.ParamField, blob.VarFieldVMType,
	)
	if err != nil {
		return nil, err
	}
	require.EqualValues(ch.Glb.T, wasmtimevm.VMType, string(res.MustGet(blob.ParamBytes)))

	res, err = ch.CallView(blob.Interface.Name, blob.FuncGetBlobField,
		blob.ParamHash, progHash,
		blob.ParamField, blob.VarFieldProgramBinary,
	)
	if err != nil {
		return nil, err
	}
	binary := res.MustGet(blob.ParamBytes)
	return binary, nil
}

func (ch *Chain) DeployContract(sigScheme signaturescheme.SignatureScheme, name string, progHash hashing.HashValue, params ...interface{}) error {
	par := []interface{}{root.ParamProgramHash, progHash, root.ParamName, name}
	par = append(par, params...)
	req := NewCall(root.Interface.Name, root.FuncDeployContract, par...)
	_, err := ch.PostRequest(req, sigScheme)
	return err
}

func (ch *Chain) DeployWasmContract(sigScheme signaturescheme.SignatureScheme, name string, fname string, params ...interface{}) error {
	hprog, err := ch.UploadWasmFromFile(sigScheme, fname)
	if err != nil {
		return err
	}
	return ch.DeployContract(sigScheme, name, hprog, params...)
}

func (ch *Chain) GetInfo() (coretypes.ChainID, coretypes.AgentID, map[coretypes.Hname]*root.ContractRecord) {
	res, err := ch.CallView(root.Interface.Name, root.FuncGetInfo)
	require.NoError(ch.Glb.T, err)

	chainID, ok, err := codec.DecodeChainID(res.MustGet(root.VarChainID))
	require.NoError(ch.Glb.T, err)
	require.True(ch.Glb.T, ok)

	chainOwnerID, ok, err := codec.DecodeAgentID(res.MustGet(root.VarChainOwnerID))
	require.NoError(ch.Glb.T, err)
	require.True(ch.Glb.T, ok)

	contracts, err := root.DecodeContractRegistry(datatypes.NewMustMap(res, root.VarContractRegistry))
	require.NoError(ch.Glb.T, err)
	return chainID, chainOwnerID, contracts
}

func (glb *Glb) GetUtxodbBalance(addr address.Address, col balance.Color) int64 {
	bals := glb.GetUtxodbBalances(addr)
	ret, _ := bals[col]
	return ret
}

func (glb *Glb) GetUtxodbBalances(addr address.Address) map[balance.Color]int64 {
	outs := glb.utxoDB.GetAddressOutputs(addr)
	ret, _ := waspconn.OutputBalancesByColor(outs)
	return ret
}

func (ch *Chain) GetAccounts() []coretypes.AgentID {
	d, err := ch.CallView(accountsc.Interface.Name, accountsc.FuncAccounts)
	require.NoError(ch.Glb.T, err)
	keys := d.KeysSorted()
	ret := make([]coretypes.AgentID, 0, len(keys)-1)
	for _, key := range keys {
		aid, ok, err := codec.DecodeAgentID([]byte(key))
		require.NoError(ch.Glb.T, err)
		require.True(ch.Glb.T, ok)
		ret = append(ret, aid)
	}
	return ret
}

func (ch *Chain) getAccountBalance(d dict.Dict, err error) coretypes.ColoredBalances {
	require.NoError(ch.Glb.T, err)
	if d.IsEmpty() {
		return cbalances.Nil
	}
	ret := make(map[balance.Color]int64)
	err = d.Iterate("", func(key kv.Key, value []byte) bool {
		col, _, err := codec.DecodeColor([]byte(key))
		require.NoError(ch.Glb.T, err)
		val, _, err := codec.DecodeInt64(value)
		require.NoError(ch.Glb.T, err)
		ret[*col] = val
		return true
	})
	require.NoError(ch.Glb.T, err)
	return cbalances.NewFromMap(ret)
}

func (ch *Chain) GetAccountBalance(agentID coretypes.AgentID) coretypes.ColoredBalances {
	return ch.getAccountBalance(
		ch.CallView(accountsc.Interface.Name, accountsc.FuncBalance, accountsc.ParamAgentID, agentID),
	)
}

func (ch *Chain) GetTotalAssets() coretypes.ColoredBalances {
	return ch.getAccountBalance(
		ch.CallView(accountsc.Interface.Name, accountsc.FuncTotalBalance),
	)
}

func (ch *Chain) GetFeeInfo(hname coretypes.Hname) (balance.Color, int64) {
	ret, err := ch.CallView(root.Interface.Name, root.FuncGetFeeInfo, root.ParamHname, hname)
	require.NoError(ch.Glb.T, err)
	require.NotEqualValues(ch.Glb.T, 0, len(ret))

	feeColor, ok, err := codec.DecodeColor(ret.MustGet(root.ParamFeeColor))
	require.NoError(ch.Glb.T, err)
	require.True(ch.Glb.T, ok)
	require.NotNil(ch.Glb.T, feeColor)

	fee, ok, err := codec.DecodeInt64(ret.MustGet(root.ParamContractFee))
	require.NoError(ch.Glb.T, err)
	require.True(ch.Glb.T, ok)
	require.True(ch.Glb.T, fee >= 0)

	return *feeColor, fee
}
