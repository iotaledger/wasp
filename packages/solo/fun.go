// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/chainid"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/eventlog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/plugins/wasmtimevm"
	"github.com/stretchr/testify/require"
)

// String is string representation for main parameters of the chain
//goland:noinspection ALL
func (ch *Chain) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Chain ID: %s\n", ch.ChainID)
	fmt.Fprintf(&buf, "Chain state controller: %s\n", ch.StateControllerAddress)
	fmt.Fprintf(&buf, "State hash: %s\n", ch.State.Hash().String())
	fmt.Fprintf(&buf, "UTXODB genesis address: %s\n", ch.Env.utxoDB.GetGenesisAddress())
	return string(buf.Bytes())
}

// DumpAccounts dumps all account balances into the human readable string
func (ch *Chain) DumpAccounts() string {
	_, chainOwnerID, _ := ch.GetInfo()
	ret := fmt.Sprintf("ChainID: %s\nChain owner: %s\n", ch.ChainID.String(), chainOwnerID.String())
	acc := ch.GetAccounts()
	for _, aid := range acc {
		ret += fmt.Sprintf("  %s:\n", aid.String())
		bals := ch.GetAccountBalance(&aid)
		bals.ForEach(func(col ledgerstate.Color, bal uint64) bool {
			ret += fmt.Sprintf("       %s: %d\n", col, bal)
			return true
		})
	}
	return ret
}

// FindContract is a view call to the 'root' smart contract on the chain.
// It returns blobCache record of the deployed smart contract with the given name
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
	record, err := root.DecodeContractRecord(retBin)
	if err != nil {
		return nil, err
	}
	if record.Name != scName {
		return nil, fmt.Errorf("smart contract '%s' not found", scName)
	}
	return record, err
}

// GetBlobInfo return info about blob with the given hash with existence flag
// The blob information is returned as a map of pairs 'blobFieldName': 'fieldDataLength'
func (ch *Chain) GetBlobInfo(blobHash hashing.HashValue) (map[string]uint32, bool) {
	res, err := ch.CallView(blob.Interface.Name, blob.FuncGetBlobInfo, blob.ParamHash, blobHash)
	require.NoError(ch.Env.T, err)
	if res.IsEmpty() {
		return nil, false
	}
	ret, err := blob.DecodeSizesMap(res)
	require.NoError(ch.Env.T, err)
	return ret, true
}

// UploadBlob calls core 'blob' smart contract blob.FuncStoreBlob entry point to upload blob
// data to the chain. It returns hash of the blob, the unique identified of it.
// Takes request token and necessary fees from the 'sigScheme' address (or OriginatorAddress if nil).
//
// The parameters must be in the form of sequence of pairs 'fieldName': 'fieldValue'
func (ch *Chain) UploadBlob(keyPair *ed25519.KeyPair, params ...interface{}) (ret hashing.HashValue, err error) {
	par := toMap(params...)
	expectedHash := blob.MustGetBlobHash(codec.MakeDict(par))
	if _, ok := ch.GetBlobInfo(expectedHash); ok {
		// blob exists, return hash of existing
		return expectedHash, nil
	}

	req := NewCallParams(blob.Interface.Name, blob.FuncStoreBlob, params...)
	feeColor, ownerFee, validatorFee := ch.GetFeeInfo(blob.Interface.Name)
	require.EqualValues(ch.Env.T, feeColor, ledgerstate.ColorIOTA)
	totalFee := ownerFee + validatorFee
	if totalFee > 0 {
		req.WithIotas(totalFee)
	} else {
		req.WithIotas(1)
	}
	res, err := ch.PostRequestSync(req, keyPair)
	if err != nil {
		return
	}
	resBin := res.MustGet(blob.ParamHash)
	ret, ok, err := codec.DecodeHashValue(resBin)
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("internal error: no hash returned")
		return
	}
	require.EqualValues(ch.Env.T, expectedHash, ret)
	return
}

// UploadBlobOptimized does the same as UploadBlob, only better but more complicated
// It allows  big data chunks to bypass the request transaction. Instead, in transaction only hash of the data is put
// The data itself must be uploaded to the node (in this case into Solo environment, separately
// Before running the request in VM, the hash references contained in the request transaction are resolved with
// the real data, previously uploaded directly.
func (ch *Chain) UploadBlobOptimized(optimalSize int, keyPair *ed25519.KeyPair, params ...interface{}) (ret hashing.HashValue, err error) {
	par := toMap(params...)
	expectedHash := blob.MustGetBlobHash(codec.MakeDict(par))
	if _, ok := ch.GetBlobInfo(expectedHash); ok {
		// blob exists, return hash of existing
		return expectedHash, nil
	}
	// creates call parameters by optimizing big data chunks, the ones larger than optimalSize.
	// The call returns map of keys/value pairs which were replaced by hashes. These data must be uploaded
	// separately
	req, toUpload := NewCallParamsOptimized(blob.Interface.Name, blob.FuncStoreBlob, optimalSize, params...)
	req.WithIotas(1)
	// the too big data we first upload into the blobCache
	for _, v := range toUpload {
		ch.Env.PutBlobDataIntoRegistry(v)
	}
	feeColor, ownerFee, validatorFee := ch.GetFeeInfo(blob.Interface.Name)
	require.EqualValues(ch.Env.T, feeColor, ledgerstate.ColorIOTA)
	totalFee := ownerFee + validatorFee
	if totalFee > 0 {
		req.WithTransfer(ledgerstate.ColorIOTA, totalFee)
	}
	res, err := ch.PostRequestSync(req, keyPair)
	if err != nil {
		return
	}
	resBin := res.MustGet(blob.ParamHash)
	ret, ok, err := codec.DecodeHashValue(resBin)
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("internal error: no hash returned")
		return
	}
	require.EqualValues(ch.Env.T, expectedHash, ret)
	return
}

const (
	OptimizeUpload  = true
	OptimalBlobSize = 512
)

// UploadWasm is a syntactic sugar of the UploadBlob used to upload Wasm binary to the chain.
//  parameter 'binaryCode' is the binary of Wasm smart contract program
//
// The blob for the Wasm binary used fixed field names which are statically known by the .
// 'root' smart contract which is responsible for the deployment of contracts on the chain
func (ch *Chain) UploadWasm(keyPair *ed25519.KeyPair, binaryCode []byte) (ret hashing.HashValue, err error) {
	if OptimizeUpload {
		return ch.UploadBlobOptimized(OptimalBlobSize, keyPair,
			blob.VarFieldVMType, wasmtimevm.VMType,
			blob.VarFieldProgramBinary, binaryCode,
		)
	}
	return ch.UploadBlob(keyPair,
		blob.VarFieldVMType, wasmtimevm.VMType,
		blob.VarFieldProgramBinary, binaryCode,
	)
}

// UploadWasmFromFile is a syntactic sugar to upload file content as blob data to the chain
func (ch *Chain) UploadWasmFromFile(keyPair *ed25519.KeyPair, fileName string) (ret hashing.HashValue, err error) {
	var binary []byte
	binary, err = ioutil.ReadFile(fileName)
	if err != nil {
		return
	}
	return ch.UploadWasm(keyPair, binary)
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
	require.EqualValues(ch.Env.T, wasmtimevm.VMType, string(res.MustGet(blob.ParamBytes)))

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
// the private key of the creator (nil defaults to chain originator). The 'creator' becomes an immutable
// property of the contract instance.
// The parameter 'programHash' can be one of the following:
//   - it is and ID of  the blob stored on the chain in the format of Wasm binary
//   - it can be a hash (ID) of the example smart contract ("hardcoded"). The "hardcoded"
//     smart contact must be made available with the call examples.AddProcessor
func (ch *Chain) DeployContract(keyPair *ed25519.KeyPair, name string, programHash hashing.HashValue, params ...interface{}) error {
	par := []interface{}{root.ParamProgramHash, programHash, root.ParamName, name}
	par = append(par, params...)
	req := NewCallParams(root.Interface.Name, root.FuncDeployContract, par...).WithIotas(1)
	_, err := ch.PostRequestSync(req, keyPair)
	return err
}

// DeployWasmContract is syntactic sugar for uploading Wasm binary from file and
// deploying the smart contract in one call
func (ch *Chain) DeployWasmContract(keyPair *ed25519.KeyPair, name string, fname string, params ...interface{}) error {
	hprog, err := ch.UploadWasmFromFile(keyPair, fname)
	if err != nil {
		return err
	}
	return ch.DeployContract(keyPair, name, hprog, params...)
}

// GetInfo return main parameters of the chain:
//  - chainID
//  - agentID of the chain owner
//  - blobCache of contract deployed on the chain in the form of map 'contract hname': 'contract record'
func (ch *Chain) GetInfo() (chainid.ChainID, coretypes.AgentID, map[coretypes.Hname]*root.ContractRecord) {
	res, err := ch.CallView(root.Interface.Name, root.FuncGetChainInfo)
	require.NoError(ch.Env.T, err)

	chainID, ok, err := codec.DecodeChainID(res.MustGet(root.VarChainID))
	require.NoError(ch.Env.T, err)
	require.True(ch.Env.T, ok)

	chainOwnerID, ok, err := codec.DecodeAgentID(res.MustGet(root.VarChainOwnerID))
	require.NoError(ch.Env.T, err)
	require.True(ch.Env.T, ok)

	contracts, err := root.DecodeContractRegistry(collections.NewMapReadOnly(res, root.VarContractRegistry))
	require.NoError(ch.Env.T, err)
	return chainID, chainOwnerID, contracts
}

// GetAddressBalance returns number of tokens of given color contained in the given address
// on the UTXODB ledger
func (env *Solo) GetAddressBalance(addr ledgerstate.Address, col ledgerstate.Color) uint64 {
	bals := env.GetAddressBalances(addr)
	ret, _ := bals[col]
	return ret
}

// GetAddressBalances returns all colored balances of the address contained in the UTXODB ledger
func (env *Solo) GetAddressBalances(addr ledgerstate.Address) map[ledgerstate.Color]uint64 {
	return env.utxoDB.GetAddressBalances(addr)
}

// GetAccounts returns all accounts on the chain with non-zero balances
func (ch *Chain) GetAccounts() []coretypes.AgentID {
	d, err := ch.CallView(accounts.Interface.Name, accounts.FuncAccounts)
	require.NoError(ch.Env.T, err)
	keys := d.KeysSorted()
	ret := make([]coretypes.AgentID, 0, len(keys)-1)
	for _, key := range keys {
		aid, ok, err := codec.DecodeAgentID([]byte(key))
		require.NoError(ch.Env.T, err)
		require.True(ch.Env.T, ok)
		ret = append(ret, aid)
	}
	return ret
}

func (ch *Chain) getAccountBalance(d dict.Dict, err error) *ledgerstate.ColoredBalances {
	require.NoError(ch.Env.T, err)
	if d.IsEmpty() {
		return ledgerstate.NewColoredBalances(nil)
	}
	ret := make(map[ledgerstate.Color]uint64)
	err = d.Iterate("", func(key kv.Key, value []byte) bool {
		col, _, err := codec.DecodeColor([]byte(key))
		require.NoError(ch.Env.T, err)
		val, _, err := codec.DecodeUint64(value)
		require.NoError(ch.Env.T, err)
		ret[col] = val
		return true
	})
	require.NoError(ch.Env.T, err)
	return ledgerstate.NewColoredBalances(ret)
}

// GetAccountBalance return all balances of colored tokens contained in the on-chain
// account controlled by the 'agentID'
func (ch *Chain) GetAccountBalance(agentID *coretypes.AgentID) *ledgerstate.ColoredBalances {
	return ch.getAccountBalance(
		ch.CallView(accounts.Interface.Name, accounts.FuncBalance, accounts.ParamAgentID, agentID),
	)
}

func (ch *Chain) GetOwnersBalance() *ledgerstate.ColoredBalances {
	return ch.GetAccountBalance(ch.OwnersAccount())
}

func (ch *Chain) GetOwnersIotas() uint64 {
	ret, _ := ch.GetAccountBalance(ch.OwnersAccount()).Get(ledgerstate.ColorIOTA)
	return ret
}

// GetTotalAssets return total sum of colored tokens contained in the on-chain accounts
func (ch *Chain) GetTotalAssets() *ledgerstate.ColoredBalances {
	return ch.getAccountBalance(
		ch.CallView(accounts.Interface.Name, accounts.FuncTotalAssets),
	)
}

// GetTotalIotas return total sum of iotas
func (ch *Chain) GetTotalIotas() uint64 {
	ret, _ := ch.GetTotalAssets().Get(ledgerstate.ColorIOTA)
	return ret
}

// GetFeeInfo returns the fee info for the specific chain and smart contract
//  - color of the fee tokens in the chain
//  - chain owner part of the fee (number of tokens)
//  - validator part of the fee (number of tokens)
// Total fee is sum of owner fee and validator fee
func (ch *Chain) GetFeeInfo(contactName string) (ledgerstate.Color, uint64, uint64) {
	hname := coretypes.Hn(contactName)
	ret, err := ch.CallView(root.Interface.Name, root.FuncGetFeeInfo, root.ParamHname, hname)
	require.NoError(ch.Env.T, err)
	require.NotEqualValues(ch.Env.T, 0, len(ret))

	feeColor, ok, err := codec.DecodeColor(ret.MustGet(root.ParamFeeColor))
	require.NoError(ch.Env.T, err)
	require.True(ch.Env.T, ok)
	require.NotNil(ch.Env.T, feeColor)

	validatorFee, ok, err := codec.DecodeUint64(ret.MustGet(root.ParamValidatorFee))
	require.NoError(ch.Env.T, err)
	require.True(ch.Env.T, ok)
	require.True(ch.Env.T, validatorFee >= 0)

	ownerFee, ok, err := codec.DecodeUint64(ret.MustGet(root.ParamOwnerFee))
	require.NoError(ch.Env.T, err)
	require.True(ch.Env.T, ok)
	require.True(ch.Env.T, ownerFee >= 0)

	return feeColor, ownerFee, validatorFee
}

// GetEventLogRecords calls the view in the  'eventlog' core smart contract to retrieve
// latest up to 50 records for a given smart contract.
// It returns records as array in time-descending order.
// More than 50 records may be retrieved by calling the view directly
func (ch *Chain) GetEventLogRecords(name string) ([]collections.TimestampedLogRecord, error) {
	res, err := ch.CallView(eventlog.Interface.Name, eventlog.FuncGetRecords,
		eventlog.ParamContractHname, coretypes.Hn(name),
	)
	if err != nil {
		return nil, err
	}
	recs := collections.NewArray16ReadOnly(res, eventlog.ParamRecords)
	ret := make([]collections.TimestampedLogRecord, recs.MustLen())
	for i := uint16(0); i < recs.MustLen(); i++ {
		data := recs.MustGetAt(i)
		rec, err := collections.ParseRawLogRecord(data)
		require.NoError(ch.Env.T, err)
		ret[i] = *rec
	}
	return ret, nil
}

// GetEventLogRecordsString return stringified response from GetEventLogRecords
// Returns latest 50 records from the log
func (ch *Chain) GetEventLogRecordsString(name string) (string, error) {
	recs, err := ch.GetEventLogRecords(name)
	if err != nil {
		return "", err
	}
	ret := fmt.Sprintf("log records for '%s':", name)
	for _, r := range recs {
		ret += fmt.Sprintf("\n%d: %s", r.Timestamp, string(r.Data))
	}
	return ret, nil
}

// GetEventLogNumRecords returns total number of eventlog records for the given contact.
func (ch *Chain) GetEventLogNumRecords(name string) int {
	res, err := ch.CallView(eventlog.Interface.Name, eventlog.FuncGetNumRecords,
		eventlog.ParamContractHname, coretypes.Hn(name),
	)
	require.NoError(ch.Env.T, err)
	ret, ok, err := codec.DecodeInt64(res.MustGet(eventlog.ParamNumRecords))
	require.NoError(ch.Env.T, err)
	require.True(ch.Env.T, ok)
	return int(ret)
}

// OwnersAccount return the agentID of the common account (controlled by the owner)
func (ch *Chain) OwnersAccount() *coretypes.AgentID {
	return coretypes.NewAgentID(ch.ChainID.AsAddress(), 0)
}

// GetLatestBlockInfo return BlockInfo for the latest block in the chain
func (ch *Chain) GetLatestBlockInfo() *blocklog.BlockInfo {
	ret, err := ch.CallView(blocklog.Interface.Name, blocklog.FuncGetLatestBlockInfo)
	require.NoError(ch.Env.T, err)
	resultDecoder := kvdecoder.New(ret, ch.Log)
	blockIndex := uint32(resultDecoder.MustGetUint64(blocklog.ParamBlockIndex))
	blockInfoBin := resultDecoder.MustGetBytes(blocklog.ParamBlockInfo)

	blockInfo, err := blocklog.BlockInfoFromBytes(blockIndex, blockInfoBin)
	require.NoError(ch.Env.T, err)
	return blockInfo
}

// GetBlockInfo return BlockInfo for the particular block index in the chain
func (ch *Chain) GetBlockInfo(blockIndex uint32) (*blocklog.BlockInfo, error) {
	ret, err := ch.CallView(blocklog.Interface.Name, blocklog.FuncGetBlockInfo,
		blocklog.ParamBlockIndex, blockIndex)
	if err != nil {
		return nil, err
	}
	resultDecoder := kvdecoder.New(ret, ch.Log)
	blockInfoBin := resultDecoder.MustGetBytes(blocklog.ParamBlockInfo)

	blockInfo, err := blocklog.BlockInfoFromBytes(blockIndex, blockInfoBin)
	require.NoError(ch.Env.T, err)
	return blockInfo, nil
}

// IsRequestProcessed checks if the request is booked on the chain as processed
func (ch *Chain) IsRequestProcessed(reqID coretypes.RequestID) bool {
	ret, err := ch.CallView(blocklog.Interface.Name, blocklog.FuncIsRequestProcessed,
		blocklog.ParamRequestID, reqID)
	require.NoError(ch.Env.T, err)
	resultDecoder := kvdecoder.New(ret, ch.Log)
	bin, err := resultDecoder.GetBytes(blocklog.ParamRequestProcessed)
	require.NoError(ch.Env.T, err)
	return bin != nil
}

// GetRequestLogRecord gets the log records for a particular request, the block index and request index in the block
func (ch *Chain) GetRequestLogRecord(reqID coretypes.RequestID) (*blocklog.RequestLogRecord, uint32, uint16, bool) {
	ret, err := ch.CallView(blocklog.Interface.Name, blocklog.FuncGetRequestLogRecord,
		blocklog.ParamRequestID, reqID)
	require.NoError(ch.Env.T, err)
	resultDecoder := kvdecoder.New(ret, ch.Log)
	binRec, err := resultDecoder.GetBytes(blocklog.ParamRequestRecord)
	if err != nil || binRec == nil {
		return nil, 0, 0, false
	}
	ret1, err := blocklog.RequestLogRecordFromBytes(binRec)
	require.NoError(ch.Env.T, err)
	blockIndex := uint32(resultDecoder.MustGetUint64(blocklog.ParamBlockIndex))
	requestIndex := uint16(resultDecoder.MustGetUint64(blocklog.ParamRequestIndex))

	return ret1, blockIndex, requestIndex, true
}

// GetRequestLogRecordsForBlock returns all request log records for a particular block
func (ch *Chain) GetRequestLogRecordsForBlock(blockIndex uint32) []*blocklog.RequestLogRecord {
	res, err := ch.CallView(blocklog.Interface.Name, blocklog.FuncGetRequestLogRecordsForBlock,
		blocklog.ParamBlockIndex, blockIndex)
	if err != nil {
		return nil
	}
	recs := collections.NewArray16ReadOnly(res, blocklog.ParamRequestRecord)
	ret := make([]*blocklog.RequestLogRecord, recs.MustLen())
	for i := range ret {
		data, err := recs.GetAt(uint16(i))
		require.NoError(ch.Env.T, err)
		ret[i], err = blocklog.RequestLogRecordFromBytes(data)
		require.NoError(ch.Env.T, err)
		ret[i].WithBlockData(blockIndex, uint16(i))
	}
	return ret
}

// GetRequestIDsForBlock returns return the list of requestIDs settled in a particular block
func (ch *Chain) GetRequestIDsForBlock(blockIndex uint32) []coretypes.RequestID {
	res, err := ch.CallView(blocklog.Interface.Name, blocklog.FuncGetRequestIDsForBlock,
		blocklog.ParamBlockIndex, blockIndex)
	if err != nil {
		ch.Log.Warnf("GetRequestIDsForBlock: %v", err)
		return nil
	}
	recs := collections.NewArray16ReadOnly(res, blocklog.ParamRequestID)
	ret := make([]coretypes.RequestID, recs.MustLen())
	for i := range ret {
		reqIDBin, err := recs.GetAt(uint16(i))
		require.NoError(ch.Env.T, err)
		ret[i], err = coretypes.RequestIDFromBytes(reqIDBin)
		require.NoError(ch.Env.T, err)
	}
	return ret
}

// GetLogRecordsForBlockRange returns all request log records for range of blocks, inclusively.
// Upper bound is 'latest block' is set to 0
func (ch *Chain) GetLogRecordsForBlockRange(fromBlockIndex, toBlockIndex uint32) []*blocklog.RequestLogRecord {
	if toBlockIndex == 0 {
		toBlockIndex = ch.GetLatestBlockInfo().BlockIndex
	}
	if fromBlockIndex > toBlockIndex {
		return nil
	}
	ret := make([]*blocklog.RequestLogRecord, 0)
	for i := fromBlockIndex; i <= toBlockIndex; i++ {
		recs := ch.GetRequestLogRecordsForBlock(i)
		require.True(ch.Env.T, i == 0 || len(recs) != 0)
		ret = append(ret, recs...)
	}
	return ret
}

func (ch *Chain) GetLogRecordsForBlockRangeAsStrings(fromBlockIndex, toBlockIndex uint32) []string {
	recs := ch.GetLogRecordsForBlockRange(fromBlockIndex, toBlockIndex)
	ret := make([]string, len(recs))
	for i := range ret {
		ret[i] = recs[i].String()
	}
	return ret
}

func (ch *Chain) GetControlAddresses() *blocklog.ControlAddresses {
	res, err := ch.CallView(blocklog.Interface.Name, blocklog.FuncControlAddresses)
	require.NoError(ch.Env.T, err)
	par := kvdecoder.New(res, ch.Log)
	ret := &blocklog.ControlAddresses{
		StateAddress:     par.MustGetAddress(blocklog.ParamStateAddress),
		GoverningAddress: par.MustGetAddress(blocklog.ParamGoverningAddress),
		SinceBlockIndex:  uint32(par.MustGetUint64(blocklog.ParamBlockIndex)),
	}
	return ret
}
