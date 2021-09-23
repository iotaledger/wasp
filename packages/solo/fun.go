// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"sort"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/accounts/commonaccount"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/stretchr/testify/require"
)

// String is string representation for main parameters of the chain
//goland:noinspection ALL
func (ch *Chain) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Chain ID: %s\n", ch.ChainID)
	fmt.Fprintf(&buf, "Chain state controller: %s\n", ch.StateControllerAddress)
	fmt.Fprintf(&buf, "State hash: %s\n", ch.State.StateCommitment().String())
	fmt.Fprintf(&buf, "UTXODB genesis address: %s\n", ch.Env.utxoDB.GetGenesisAddress())
	return buf.String()
}

// DumpAccounts dumps all account balances into the human readable string
func (ch *Chain) DumpAccounts() string {
	_, chainOwnerID, _ := ch.GetInfo()
	ret := fmt.Sprintf("ChainID: %s\nChain owner: %s\n", ch.ChainID.String(), chainOwnerID.String())
	acc := ch.GetAccounts()
	for i := range acc {
		aid := acc[i]
		ret += fmt.Sprintf("  %s:\n", aid.String())
		bals := ch.GetAccountBalance(&aid)
		bals.ForEachRandomly(func(col colored.Color, bal uint64) bool {
			ret += fmt.Sprintf("       %s: %d\n", col, bal)
			return true
		})
	}
	return ret
}

// FindContract is a view call to the 'root' smart contract on the chain.
// It returns blobCache record of the deployed smart contract with the given name
func (ch *Chain) FindContract(scName string) (*root.ContractRecord, error) {
	retDict, err := ch.CallView(root.Contract.Name, root.FuncFindContract.Name,
		root.ParamHname, iscp.Hn(scName),
	)
	if err != nil {
		return nil, err
	}
	retBin, err := retDict.Get(root.ParamContractRecData)
	if err != nil {
		return nil, err
	}
	if retBin == nil {
		return nil, fmt.Errorf("smart contract '%s' not found", scName)
	}
	record, err := root.ContractRecordFromBytes(retBin)
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
	res, err := ch.CallView(blob.Contract.Name, blob.FuncGetBlobInfo.Name, blob.ParamHash, blobHash)
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
// The parameters must be either a dict.Dict, or a sequence of pairs 'fieldName': 'fieldValue'
func (ch *Chain) UploadBlob(keyPair *ed25519.KeyPair, params ...interface{}) (ret hashing.HashValue, err error) {
	expectedHash := blob.MustGetBlobHash(parseParams(params))
	if _, ok := ch.GetBlobInfo(expectedHash); ok {
		// blob exists, return hash of existing
		return expectedHash, nil
	}

	req := NewCallParams(blob.Contract.Name, blob.FuncStoreBlob.Name, params...)
	feeColor, ownerFee, validatorFee := ch.GetFeeInfo(blob.Contract.Name)
	require.EqualValues(ch.Env.T, feeColor, colored.IOTA)
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
	return ret, err
}

// UploadBlobOptimized does the same as UploadBlob, only better but more complicated
// It allows  big data chunks to bypass the request transaction. Instead, in transaction only hash of the data is put
// The data itself must be uploaded to the node (in this case into Solo environment, separately
// Before running the request in VM, the hash references contained in the request transaction are resolved with
// the real data, previously uploaded directly.
func (ch *Chain) UploadBlobOptimized(optimalSize int, keyPair *ed25519.KeyPair, params ...interface{}) (ret hashing.HashValue, err error) {
	expectedHash := blob.MustGetBlobHash(parseParams(params))
	if _, ok := ch.GetBlobInfo(expectedHash); ok {
		// blob exists, return hash of existing
		return expectedHash, nil
	}
	// creates call parameters by optimizing big data chunks, the ones larger than optimalSize.
	// The call returns map of keys/value pairs which were replaced by hashes. These data must be uploaded
	// separately
	req, toUpload := NewCallParamsOptimized(blob.Contract.Name, blob.FuncStoreBlob.Name, optimalSize, params...)
	req.WithIotas(1)
	// the too big data we first upload into the blobCache
	for _, v := range toUpload {
		ch.Env.PutBlobDataIntoRegistry(v)
	}
	feeColor, ownerFee, validatorFee := ch.GetFeeInfo(blob.Contract.Name)
	require.EqualValues(ch.Env.T, feeColor, colored.IOTA)
	totalFee := ownerFee + validatorFee
	if totalFee > 0 {
		req.WithIotas(totalFee)
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
	return ret, err
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
			blob.VarFieldVMType, vmtypes.WasmTime,
			blob.VarFieldProgramBinary, binaryCode,
		)
	}
	return ch.UploadBlob(keyPair,
		blob.VarFieldVMType, vmtypes.WasmTime,
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
	res, err := ch.CallView(blob.Contract.Name, blob.FuncGetBlobField.Name,
		blob.ParamHash, progHash,
		blob.ParamField, blob.VarFieldVMType,
	)
	if err != nil {
		return nil, err
	}
	require.EqualValues(ch.Env.T, vmtypes.WasmTime, string(res.MustGet(blob.ParamBytes)))

	res, err = ch.CallView(blob.Contract.Name, blob.FuncGetBlobField.Name,
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
	req := NewCallParams(root.Contract.Name, root.FuncDeployContract.Name, par...).WithIotas(1)
	_, err := ch.PostRequestSync(req, keyPair)
	return err
}

// DeployWasmContract is syntactic sugar for uploading Wasm binary from file and
// deploying the smart contract in one call
func (ch *Chain) DeployWasmContract(keyPair *ed25519.KeyPair, name, fname string, params ...interface{}) error {
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
func (ch *Chain) GetInfo() (iscp.ChainID, iscp.AgentID, map[iscp.Hname]*root.ContractRecord) {
	res, err := ch.CallView(governance.Contract.Name, governance.FuncGetChainInfo.Name)
	require.NoError(ch.Env.T, err)

	chainID, ok, err := codec.DecodeChainID(res.MustGet(governance.VarChainID))
	require.NoError(ch.Env.T, err)
	require.True(ch.Env.T, ok)

	chainOwnerID, ok, err := codec.DecodeAgentID(res.MustGet(governance.VarChainOwnerID))
	require.NoError(ch.Env.T, err)
	require.True(ch.Env.T, ok)

	res, err = ch.CallView(root.Contract.Name, root.FuncGetContractRecords.Name)
	require.NoError(ch.Env.T, err)

	contracts, err := root.DecodeContractRegistry(collections.NewMapReadOnly(res, root.VarContractRegistry))
	require.NoError(ch.Env.T, err)
	return chainID, chainOwnerID, contracts
}

// GetAddressBalance returns number of tokens of given color contained in the given address
// on the UTXODB ledger
func (env *Solo) GetAddressBalance(addr ledgerstate.Address, col colored.Color) uint64 {
	bals := env.GetAddressBalances(addr)
	ret := bals[col]
	return ret
}

// GetAddressBalances returns all colored balances of the address contained in the UTXODB ledger
func (env *Solo) GetAddressBalances(addr ledgerstate.Address) colored.Balances {
	return colored.BalancesFromL1Map(env.utxoDB.GetAddressBalances(addr))
}

// GetAccounts returns all accounts on the chain with non-zero balances
func (ch *Chain) GetAccounts() []iscp.AgentID {
	d, err := ch.CallView(accounts.Contract.Name, accounts.FuncViewAccounts.Name)
	require.NoError(ch.Env.T, err)
	keys := d.KeysSorted()
	ret := make([]iscp.AgentID, 0, len(keys)-1)
	for _, key := range keys {
		aid, ok, err := codec.DecodeAgentID([]byte(key))
		require.NoError(ch.Env.T, err)
		require.True(ch.Env.T, ok)
		ret = append(ret, aid)
	}
	return ret
}

func (ch *Chain) parseAccountBalance(d dict.Dict, err error) colored.Balances {
	require.NoError(ch.Env.T, err)
	if d.IsEmpty() {
		return colored.NewBalances()
	}
	ret := colored.NewBalances()
	err = d.Iterate("", func(key kv.Key, value []byte) bool {
		col, _, err := codec.DecodeColor([]byte(key))
		require.NoError(ch.Env.T, err)
		val, _, err := codec.DecodeUint64(value)
		require.NoError(ch.Env.T, err)
		ret.Set(col, val)
		return true
	})
	require.NoError(ch.Env.T, err)
	return ret
}

func (ch *Chain) GetOnChainLedger() map[string]colored.Balances {
	accs := ch.GetAccounts()
	ret := make(map[string]colored.Balances)
	for i := range accs {
		ret[accs[i].String()] = ch.GetAccountBalance(&accs[i])
	}
	return ret
}

func (ch *Chain) GetOnChainLedgerString() string {
	l := ch.GetOnChainLedger()
	keys := make([]string, 0, len(l))
	for aid := range l {
		keys = append(keys, aid)
	}
	sort.Strings(keys)
	ret := ""
	for _, aid := range keys {
		ret += aid + "\n"
		ret += "        " + l[aid].String() + "\n"
	}
	return ret
}

// GetAccountBalance return all balances of colored tokens contained in the on-chain
// account controlled by the 'agentID'
func (ch *Chain) GetAccountBalance(agentID *iscp.AgentID) colored.Balances {
	return ch.parseAccountBalance(
		ch.CallView(accounts.Contract.Name, accounts.FuncViewBalance.Name, accounts.ParamAgentID, agentID),
	)
}

func (ch *Chain) GetCommonAccountBalance() colored.Balances {
	return ch.GetAccountBalance(ch.CommonAccount())
}

func (ch *Chain) GetCommonAccountIotas() uint64 {
	ret := ch.GetAccountBalance(ch.CommonAccount()).Get(colored.IOTA)
	return ret
}

// GetTotalAssets return total sum of colored tokens contained in the on-chain accounts
func (ch *Chain) GetTotalAssets() colored.Balances {
	return ch.parseAccountBalance(
		ch.CallView(accounts.Contract.Name, accounts.FuncViewTotalAssets.Name),
	)
}

// GetTotalIotas return total sum of iotas
func (ch *Chain) GetTotalIotas() uint64 {
	ret := ch.GetTotalAssets().Get(colored.IOTA)
	return ret
}

// GetFeeInfo returns the fee info for the specific chain and smart contract
//  - color of the fee tokens in the chain
//  - chain owner part of the fee (number of tokens)
//  - validator part of the fee (number of tokens)
// Total fee is sum of owner fee and validator fee
func (ch *Chain) GetFeeInfo(contactName string) (colored.Color, uint64, uint64) {
	hname := iscp.Hn(contactName)
	ret, err := ch.CallView(governance.Contract.Name, governance.FuncGetFeeInfo.Name, governance.ParamHname, hname)
	require.NoError(ch.Env.T, err)
	require.NotEqualValues(ch.Env.T, 0, len(ret))

	feeColor, ok, err := codec.DecodeColor(ret.MustGet(governance.ParamFeeColor))
	require.NoError(ch.Env.T, err)
	require.True(ch.Env.T, ok)
	require.NotNil(ch.Env.T, feeColor)

	validatorFee, ok, err := codec.DecodeUint64(ret.MustGet(governance.ParamValidatorFee))
	require.NoError(ch.Env.T, err)
	require.True(ch.Env.T, ok)

	ownerFee, ok, err := codec.DecodeUint64(ret.MustGet(governance.ParamOwnerFee))
	require.NoError(ch.Env.T, err)
	require.True(ch.Env.T, ok)

	return feeColor, ownerFee, validatorFee
}

func eventsFromViewResult(t TestContext, viewResult dict.Dict) []string {
	recs := collections.NewArray16ReadOnly(viewResult, blocklog.ParamEvent)
	ret := make([]string, recs.MustLen())
	for i := range ret {
		data, err := recs.GetAt(uint16(i))
		require.NoError(t, err)
		ret[i] = string(data)
	}
	return ret
}

// GetEventsForContract calls the view in the  'blocklog' core smart contract to retrieve events for a given smart contract.
func (ch *Chain) GetEventsForContract(name string) ([]string, error) {
	viewResult, err := ch.CallView(
		blocklog.Contract.Name, blocklog.FuncGetEventsForContract.Name,
		blocklog.ParamContractHname, iscp.Hn(name),
	)
	if err != nil {
		return nil, err
	}

	return eventsFromViewResult(ch.Env.T, viewResult), nil
}

// GetEventsForRequest calls the view in the  'blocklog' core smart contract to retrieve events for a given request.
func (ch *Chain) GetEventsForRequest(reqID iscp.RequestID) ([]string, error) {
	viewResult, err := ch.CallView(
		blocklog.Contract.Name, blocklog.FuncGetEventsForRequest.Name,
		blocklog.ParamRequestID, reqID,
	)
	if err != nil {
		return nil, err
	}
	return eventsFromViewResult(ch.Env.T, viewResult), nil
}

// GetEventsForBlock calls the view in the 'blocklog' core smart contract to retrieve events for a given block.
func (ch *Chain) GetEventsForBlock(blockIndex uint32) ([]string, error) {
	viewResult, err := ch.CallView(
		blocklog.Contract.Name, blocklog.FuncGetEventsForBlock.Name,
		blocklog.ParamBlockIndex, blockIndex,
	)
	if err != nil {
		return nil, err
	}
	return eventsFromViewResult(ch.Env.T, viewResult), nil
}

// CommonAccount return the agentID of the common account (controlled by the owner)
func (ch *Chain) CommonAccount() *iscp.AgentID {
	return commonaccount.Get(&ch.ChainID)
}

// GetLatestBlockInfo return BlockInfo for the latest block in the chain
func (ch *Chain) GetLatestBlockInfo() *blocklog.BlockInfo {
	ret, err := ch.CallView(blocklog.Contract.Name, blocklog.FuncGetLatestBlockInfo.Name)
	require.NoError(ch.Env.T, err)
	resultDecoder := kvdecoder.New(ret, ch.Log)
	blockIndex := resultDecoder.MustGetUint32(blocklog.ParamBlockIndex)
	blockInfoBin := resultDecoder.MustGetBytes(blocklog.ParamBlockInfo)

	blockInfo, err := blocklog.BlockInfoFromBytes(blockIndex, blockInfoBin)
	require.NoError(ch.Env.T, err)
	return blockInfo
}

// GetBlockInfo return BlockInfo for the particular block index in the chain
func (ch *Chain) GetBlockInfo(blockIndex uint32) (*blocklog.BlockInfo, error) {
	ret, err := ch.CallView(blocklog.Contract.Name, blocklog.FuncGetBlockInfo.Name,
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
func (ch *Chain) IsRequestProcessed(reqID iscp.RequestID) bool {
	ret, err := ch.CallView(blocklog.Contract.Name, blocklog.FuncIsRequestProcessed.Name,
		blocklog.ParamRequestID, reqID)
	require.NoError(ch.Env.T, err)
	resultDecoder := kvdecoder.New(ret, ch.Log)
	bin, err := resultDecoder.GetBytes(blocklog.ParamRequestProcessed)
	require.NoError(ch.Env.T, err)
	return bin != nil
}

// GetRequestReceipt gets the log records for a particular request, the block index and request index in the block
func (ch *Chain) GetRequestReceipt(reqID iscp.RequestID) (*blocklog.RequestReceipt, uint32, uint16, bool) {
	ret, err := ch.CallView(blocklog.Contract.Name, blocklog.FuncGetRequestReceipt.Name,
		blocklog.ParamRequestID, reqID)
	require.NoError(ch.Env.T, err)
	resultDecoder := kvdecoder.New(ret, ch.Log)
	binRec, err := resultDecoder.GetBytes(blocklog.ParamRequestRecord)
	if err != nil || binRec == nil {
		return nil, 0, 0, false
	}
	ret1, err := blocklog.RequestReceiptFromBytes(binRec)
	require.NoError(ch.Env.T, err)
	blockIndex := resultDecoder.MustGetUint32(blocklog.ParamBlockIndex)
	requestIndex := resultDecoder.MustGetUint16(blocklog.ParamRequestIndex)

	return ret1, blockIndex, requestIndex, true
}

// GetRequestReceiptsForBlock returns all request log records for a particular block
func (ch *Chain) GetRequestReceiptsForBlock(blockIndex uint32) []*blocklog.RequestReceipt {
	res, err := ch.CallView(blocklog.Contract.Name, blocklog.FuncGetRequestReceiptsForBlock.Name,
		blocklog.ParamBlockIndex, blockIndex)
	if err != nil {
		return nil
	}
	recs := collections.NewArray16ReadOnly(res, blocklog.ParamRequestRecord)
	ret := make([]*blocklog.RequestReceipt, recs.MustLen())
	for i := range ret {
		data, err := recs.GetAt(uint16(i))
		require.NoError(ch.Env.T, err)
		ret[i], err = blocklog.RequestReceiptFromBytes(data)
		require.NoError(ch.Env.T, err)
		ret[i].WithBlockData(blockIndex, uint16(i))
	}
	return ret
}

// GetRequestIDsForBlock returns return the list of requestIDs settled in a particular block
func (ch *Chain) GetRequestIDsForBlock(blockIndex uint32) []iscp.RequestID {
	res, err := ch.CallView(blocklog.Contract.Name, blocklog.FuncGetRequestIDsForBlock.Name,
		blocklog.ParamBlockIndex, blockIndex)
	if err != nil {
		ch.Log.Warnf("GetRequestIDsForBlock: %v", err)
		return nil
	}
	recs := collections.NewArray16ReadOnly(res, blocklog.ParamRequestID)
	ret := make([]iscp.RequestID, recs.MustLen())
	for i := range ret {
		reqIDBin, err := recs.GetAt(uint16(i))
		require.NoError(ch.Env.T, err)
		ret[i], err = iscp.RequestIDFromBytes(reqIDBin)
		require.NoError(ch.Env.T, err)
	}
	return ret
}

// GetRequestReceiptsForBlockRange returns all request log records for range of blocks, inclusively.
// Upper bound is 'latest block' is set to 0
func (ch *Chain) GetRequestReceiptsForBlockRange(fromBlockIndex, toBlockIndex uint32) []*blocklog.RequestReceipt {
	if toBlockIndex == 0 {
		toBlockIndex = ch.GetLatestBlockInfo().BlockIndex
	}
	if fromBlockIndex > toBlockIndex {
		return nil
	}
	ret := make([]*blocklog.RequestReceipt, 0)
	for i := fromBlockIndex; i <= toBlockIndex; i++ {
		recs := ch.GetRequestReceiptsForBlock(i)
		require.True(ch.Env.T, i == 0 || len(recs) != 0)
		ret = append(ret, recs...)
	}
	return ret
}

func (ch *Chain) GetRequestReceiptsForBlockRangeAsStrings(fromBlockIndex, toBlockIndex uint32) []string {
	recs := ch.GetRequestReceiptsForBlockRange(fromBlockIndex, toBlockIndex)
	ret := make([]string, len(recs))
	for i := range ret {
		ret[i] = recs[i].String()
	}
	return ret
}

func (ch *Chain) GetControlAddresses() *blocklog.ControlAddresses {
	res, err := ch.CallView(blocklog.Contract.Name, blocklog.FuncControlAddresses.Name)
	require.NoError(ch.Env.T, err)
	par := kvdecoder.New(res, ch.Log)
	ret := &blocklog.ControlAddresses{
		StateAddress:     par.MustGetAddress(blocklog.ParamStateControllerAddress),
		GoverningAddress: par.MustGetAddress(blocklog.ParamGoverningAddress),
		SinceBlockIndex:  par.MustGetUint32(blocklog.ParamBlockIndex),
	}
	return ret
}

// AddAllowedStateController adds the address to the allowed state controlled address list
func (ch *Chain) AddAllowedStateController(addr ledgerstate.Address, keyPair *ed25519.KeyPair) error {
	req := NewCallParams(coreutil.CoreContractGovernance, governance.FuncAddAllowedStateControllerAddress.Name,
		governance.ParamStateControllerAddress, addr,
	).WithIotas(1)
	_, err := ch.PostRequestSync(req, keyPair)
	return err
}

// AddAllowedStateController adds the address to the allowed state controlled address list
func (ch *Chain) RemoveAllowedStateController(addr ledgerstate.Address, keyPair *ed25519.KeyPair) error {
	req := NewCallParams(coreutil.CoreContractGovernance, governance.FuncRemoveAllowedStateControllerAddress.Name,
		governance.ParamStateControllerAddress, addr,
	).WithIotas(1)
	_, err := ch.PostRequestSync(req, keyPair)
	return err
}

// AddAllowedStateController adds the address to the allowed state controlled address list
func (ch *Chain) GetAllowedStateControllerAddresses() []ledgerstate.Address {
	res, err := ch.CallView(coreutil.CoreContractGovernance, governance.FuncGetAllowedStateControllerAddresses.Name)
	require.NoError(ch.Env.T, err)
	if len(res) == 0 {
		return nil
	}
	ret := make([]ledgerstate.Address, 0)
	arr := collections.NewArray16ReadOnly(res, governance.ParamAllowedStateControllerAddresses)
	for i := uint16(0); i < arr.MustLen(); i++ {
		a, ok, err := codec.DecodeAddress(arr.MustGetAt(i))
		require.NoError(ch.Env.T, err)
		require.True(ch.Env.T, ok)
		ret = append(ret, a)
	}
	return ret
}

// RotateStateController rotates the chain to the new controller address.
// We assume self-governed chain here.
// Mostly use for the testinng of committee rotation logic, otherwise not much needed for smart contract testing
func (ch *Chain) RotateStateController(newStateAddr ledgerstate.Address, newStateKeyPair, ownerKeyPair *ed25519.KeyPair) error {
	req := NewCallParams(coreutil.CoreContractGovernance, coreutil.CoreEPRotateStateController,
		coreutil.ParamStateControllerAddress, newStateAddr,
	).WithIotas(1)
	err := ch.postRequestSyncTxSpecial(req, ownerKeyPair)
	if err == nil {
		ch.StateControllerAddress = newStateAddr
		ch.StateControllerKeyPair = newStateKeyPair
	}
	return err
}

func (ch *Chain) postRequestSyncTxSpecial(req *CallParams, keyPair *ed25519.KeyPair) error {
	tx, _, err := ch.RequestFromParamsToLedger(req, keyPair)
	if err != nil {
		return err
	}
	reqs, err := ch.Env.RequestsForChain(tx, ch.ChainID)
	require.NoError(ch.Env.T, err)
	_, err = ch.runRequestsSync(reqs, "postSpecial")
	return err
}
