// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

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
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accounts"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/chainlog"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/plugins/wasmtimevm"
	"github.com/stretchr/testify/require"
	"io/ioutil"
)

// String is string representation for main parameters of the chain
//goland:noinspection ALL
func (ch *Chain) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Chain ID: %s\n", ch.ChainID.String())
	fmt.Fprintf(&buf, "Chain address: %s\n", ch.ChainAddress.String())
	fmt.Fprintf(&buf, "State hash: %s\n", ch.State.Hash().String())
	fmt.Fprintf(&buf, "UTXODB genesis address: %s\n", ch.Glb.utxoDB.GetGenesisAddress().String())
	return string(buf.Bytes())
}

// FindContract is a view call to the 'root' smart contract on the chain.
// It returns registry record of the deployed smart contract with the given name
func (ch *Chain) FindContract(scName string) (*root.ContractRecord, error) {
	retDict, err := ch.CallView(root.Interface.Name, root.FuncFindContract,
		root.ParamHname, coretypes.Hn(scName),
	)
	if err != nil {
		return nil, err
	}
	retBin, err := retDict.Get(root.ParamData)
	if err != nil {
		return nil, err
	}
	if retBin == nil {
		return nil, fmt.Errorf("smart contract '%s' not found", scName)
	}
	return root.DecodeContractRecord(retBin)
}

// GetBlobInfo return info about blob with the given hash with existence flag
// The blob information is returned as a map of pairs 'blobFieldName': 'fieldDataLength'
func (ch *Chain) GetBlobInfo(blobHash hashing.HashValue) (map[string]uint32, bool) {
	res, err := ch.CallView(blob.Interface.Name, blob.FuncGetBlobInfo, blob.ParamHash, blobHash)
	require.NoError(ch.Glb.T, err)
	if res.IsEmpty() {
		return nil, false
	}
	ret, err := blob.DecodeSizesMap(res)
	require.NoError(ch.Glb.T, err)
	return ret, true
}

// UploadBlob calls core 'blob' smart contract blob.FuncStoreBlob entry point to upload blob
// data to the chain. It returns hash of the blob, the unique identified of it.
// Takes request token and necessary fees from the 'sigScheme' address (or OriginatorAddress if nil).
//
// The parameters must be in the form of sequence of pairs 'fieldName': 'fieldValue'
func (ch *Chain) UploadBlob(sigScheme signaturescheme.SignatureScheme, params ...interface{}) (ret hashing.HashValue, err error) {
	par := toMap(params...)
	expectedHash := blob.MustGetBlobHash(codec.MakeDict(par))
	if _, ok := ch.GetBlobInfo(expectedHash); ok {
		// blob exists, return hash of existing
		return expectedHash, nil
	}

	feeColor, ownerFee, validatorFee := ch.GetFeeInfo(blob.Interface.Name)
	require.EqualValues(ch.Glb.T, feeColor, balance.ColorIOTA)
	totalFee := ownerFee + validatorFee
	transferMap := map[balance.Color]int64{}
	if totalFee > 0 {
		transferMap[balance.ColorIOTA] = totalFee
	}
	req := NewCall(blob.Interface.Name, blob.FuncStoreBlob, params...).WithTransfer(transferMap)
	var res dict.Dict
	res, err = ch.PostRequest(req, sigScheme)
	if err != nil {
		return
	}
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

// UploadWasm is a syntactic sugar of the UploadBlob used to upload Wasm binary to the chain.
//  parameter 'binaryCode' is the binary of Wasm smart contract program
//
// The blob for the Wasm binary used fixed field names which are statically known by the .
// 'root' smart contract which is responsible for the deployment of contracts on the chain
func (ch *Chain) UploadWasm(sigScheme signaturescheme.SignatureScheme, binaryCode []byte) (ret hashing.HashValue, err error) {
	return ch.UploadBlob(sigScheme,
		blob.VarFieldVMType, wasmtimevm.VMType,
		blob.VarFieldProgramBinary, binaryCode,
	)
}

// UploadWasmFromFile is a syntactic sugar to upload file content as blob data to the chain
func (ch *Chain) UploadWasmFromFile(sigScheme signaturescheme.SignatureScheme, fileName string) (ret hashing.HashValue, err error) {
	var binary []byte
	binary, err = ioutil.ReadFile(fileName)
	if err != nil {
		return
	}
	return ch.UploadWasm(sigScheme, binary)
}

// GetWasmBinary retrieves program binary in the format of Wasm blob from the chain by hash.
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

// DeployContract deploys contract with the given name by its 'programHash'. 'sigScheme' represents
// the private key of the creator (nil defaults to chain originator). The creator becomes the
// initial owner of the chain.
// The parameter 'programHash' can be one of the following:
//   - it is and ID of  the blob stored on the chain in the format of Wasm binary
//   - it can be a hash (ID) of the example smart contract ("hardcoded"). The "hardcoded"
//     smart contact must be made available with the call examples.AddProcessor
func (ch *Chain) DeployContract(sigScheme signaturescheme.SignatureScheme, name string, programHash hashing.HashValue, params ...interface{}) error {
	par := []interface{}{root.ParamProgramHash, programHash, root.ParamName, name}
	par = append(par, params...)
	req := NewCall(root.Interface.Name, root.FuncDeployContract, par...)
	_, err := ch.PostRequest(req, sigScheme)
	return err
}

// DeployWasmContract is syntactic sugar for uploading Wasm binary from file and
// deploying the smart contract in one call
func (ch *Chain) DeployWasmContract(sigScheme signaturescheme.SignatureScheme, name string, fname string, params ...interface{}) error {
	hprog, err := ch.UploadWasmFromFile(sigScheme, fname)
	if err != nil {
		return err
	}
	return ch.DeployContract(sigScheme, name, hprog, params...)
}

type ChainInfo struct {
	ChainID      coretypes.ChainID
	ChainOwnerID coretypes.AgentID
	ChainColor   balance.Color
	ChainAddress address.Address
}

// GetInfo return main parameters of the chain:
//  - chainID
//  - agentID of the chain owner
//  - registry of contract deployed on the chain in the form of map 'contract hname': 'contract record'
func (ch *Chain) GetInfo() (ChainInfo, map[coretypes.Hname]*root.ContractRecord) {
	res, err := ch.CallView(root.Interface.Name, root.FuncGetChainInfo)
	require.NoError(ch.Glb.T, err)

	chainID, ok, err := codec.DecodeChainID(res.MustGet(root.VarChainID))
	require.NoError(ch.Glb.T, err)
	require.True(ch.Glb.T, ok)

	chainColor, ok, err := codec.DecodeColor(res.MustGet(root.VarChainColor))
	require.NoError(ch.Glb.T, err)
	require.True(ch.Glb.T, ok)

	chainAddress, ok, err := codec.DecodeAddress(res.MustGet(root.VarChainAddress))
	require.NoError(ch.Glb.T, err)
	require.True(ch.Glb.T, ok)

	chainOwnerID, ok, err := codec.DecodeAgentID(res.MustGet(root.VarChainOwnerID))
	require.NoError(ch.Glb.T, err)
	require.True(ch.Glb.T, ok)

	contracts, err := root.DecodeContractRegistry(datatypes.NewMustMap(res, root.VarContractRegistry))
	require.NoError(ch.Glb.T, err)
	return ChainInfo{
		ChainID:      chainID,
		ChainOwnerID: chainOwnerID,
		ChainColor:   chainColor,
		ChainAddress: *chainAddress,
	}, contracts
}

// GetUtxodbBalance returns number of tokens of given color contained in the given address
// on the UTXODB ledger
func (glb *Solo) GetUtxodbBalance(addr address.Address, col balance.Color) int64 {
	bals := glb.GetUtxodbBalances(addr)
	ret, _ := bals[col]
	return ret
}

// GetUtxodbBalances returns all colored balances of the address contained in the UTXODB ledger
func (glb *Solo) GetUtxodbBalances(addr address.Address) map[balance.Color]int64 {
	outs := glb.utxoDB.GetAddressOutputs(addr)
	ret, _ := waspconn.OutputBalancesByColor(outs)
	return ret
}

// GetAccounts returns all accounts on the chain with non-zero balances
func (ch *Chain) GetAccounts() []coretypes.AgentID {
	d, err := ch.CallView(accounts.Interface.Name, accounts.FuncAccounts)
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
		ret[col] = val
		return true
	})
	require.NoError(ch.Glb.T, err)
	return cbalances.NewFromMap(ret)
}

// GetAccountBalance return all balances of colored tokens contained in the on-chain
// account controlled by the 'agentID'
func (ch *Chain) GetAccountBalance(agentID coretypes.AgentID) coretypes.ColoredBalances {
	return ch.getAccountBalance(
		ch.CallView(accounts.Interface.Name, accounts.FuncBalance, accounts.ParamAgentID, agentID),
	)
}

// GetTotalAssets return total sum of colored tokens contained in the on-chain accounts
func (ch *Chain) GetTotalAssets() coretypes.ColoredBalances {
	return ch.getAccountBalance(
		ch.CallView(accounts.Interface.Name, accounts.FuncTotalAssets),
	)
}

// GetFeeInfo returns the fee info for the specific chain and smart contract
//  - color of the fee tokens in the chain
//  - chain owner part of the fee (number of tokens)
//  - validator part of the fee (number of tokens)
// Total fee is sum of owner fee and validator fee
func (ch *Chain) GetFeeInfo(contactName string) (balance.Color, int64, int64) {
	hname := coretypes.Hn(contactName)
	ret, err := ch.CallView(root.Interface.Name, root.FuncGetFeeInfo, root.ParamHname, hname)
	require.NoError(ch.Glb.T, err)
	require.NotEqualValues(ch.Glb.T, 0, len(ret))

	feeColor, ok, err := codec.DecodeColor(ret.MustGet(root.ParamFeeColor))
	require.NoError(ch.Glb.T, err)
	require.True(ch.Glb.T, ok)
	require.NotNil(ch.Glb.T, feeColor)

	validatorFee, ok, err := codec.DecodeInt64(ret.MustGet(root.ParamValidatorFee))
	require.NoError(ch.Glb.T, err)
	require.True(ch.Glb.T, ok)
	require.True(ch.Glb.T, validatorFee >= 0)

	ownerFee, ok, err := codec.DecodeInt64(ret.MustGet(root.ParamOwnerFee))
	require.NoError(ch.Glb.T, err)
	require.True(ch.Glb.T, ok)
	require.True(ch.Glb.T, ownerFee >= 0)

	return feeColor, ownerFee, validatorFee
}

// GetChainLogRecords calls the view in the  'chainlog' core smart contract to retrieve
// latest up to 50 records for a given smart contract.
// It returns records as array in time-descending order.
//
// More than 50 records may be retrieved by calling the view directly
func (ch *Chain) GetChainLogRecords(name string) ([]datatypes.TimestampedLogRecord, error) {
	res, err := ch.CallView(chainlog.Interface.Name, chainlog.FuncGetLogRecords,
		chainlog.ParamContractHname, coretypes.Hn(name),
	)
	if err != nil {
		return nil, err
	}
	recs := datatypes.NewMustArray(res, chainlog.ParamRecords)
	ret := make([]datatypes.TimestampedLogRecord, recs.Len())
	for i := uint16(0); i < recs.Len(); i++ {
		data := recs.GetAt(i)
		rec, err := datatypes.ParseRawLogRecord(data)
		require.NoError(ch.Glb.T, err)
		ret[i] = *rec
	}
	return ret, nil
}

// GetChainLogRecordsString return stringified response from GetChainLogRecords
func (ch *Chain) GetChainLogRecordsString(name string) (string, error) {
	recs, err := ch.GetChainLogRecords(name)
	if err != nil {
		return "", err
	}
	ret := fmt.Sprintf("log records for '%s':", name)
	for _, r := range recs {
		ret += fmt.Sprintf("\n%d: %s", r.Timestamp, string(r.Data))
	}
	return ret, nil
}

// GetChainLogNumRecords returns total number of chainlog records for the given contact.
func (ch *Chain) GetChainLogNumRecords(name string) int {
	res, err := ch.CallView(chainlog.Interface.Name, chainlog.FuncGetNumRecords,
		chainlog.ParamContractHname, coretypes.Hn(name),
	)
	require.NoError(ch.Glb.T, err)
	ret, ok, err := codec.DecodeInt64(res.MustGet(chainlog.ParamNumRecords))
	require.NoError(ch.Glb.T, err)
	require.True(ch.Glb.T, ok)
	return int(ret)
}
