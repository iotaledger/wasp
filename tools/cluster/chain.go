package cluster

import (
	"bytes"
	"fmt"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/client/multiclient"
	"github.com/iotaledger/wasp/client/scclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type Chain struct {
	Description string

	OriginatorKeyPair *cryptolib.KeyPair

	AllPeers       []int
	CommitteeNodes []int
	Quorum         uint16
	StateAddress   iotago.Address

	ChainID *isc.ChainID

	Cluster *Cluster
}

func (ch *Chain) ChainAddress() iotago.Address {
	return ch.ChainID.AsAddress()
}

func (ch *Chain) CommitteeAPIHosts() []string {
	return ch.Cluster.Config.APIHosts(ch.CommitteeNodes)
}

func (ch *Chain) CommitteePeeringHosts() []string {
	return ch.Cluster.Config.PeeringHosts(ch.CommitteeNodes)
}

func (ch *Chain) AllPeeringHosts() []string {
	return ch.Cluster.Config.PeeringHosts(ch.AllPeers)
}

func (ch *Chain) AllAPIHosts() []string {
	return ch.Cluster.Config.APIHosts(ch.AllPeers)
}

func (ch *Chain) OriginatorAddress() iotago.Address {
	addr := ch.OriginatorKeyPair.Address()
	return addr
}

func (ch *Chain) OriginatorID() isc.AgentID {
	ret := isc.NewAgentID(ch.OriginatorAddress())
	return ret
}

func (ch *Chain) OriginatorClient() *chainclient.Client {
	return ch.Client(ch.OriginatorKeyPair)
}

func (ch *Chain) Client(sigScheme *cryptolib.KeyPair, nodeIndex ...int) *chainclient.Client {
	idx := 0
	if len(nodeIndex) == 1 {
		idx = nodeIndex[0]
	}
	return chainclient.New(
		ch.Cluster.L1Client(),
		ch.Cluster.WaspClient(idx),
		ch.ChainID,
		sigScheme,
	)
}

func (ch *Chain) SCClient(contractHname isc.Hname, sigScheme *cryptolib.KeyPair, nodeIndex ...int) *scclient.SCClient {
	return scclient.New(ch.Client(sigScheme, nodeIndex...), contractHname)
}

func (ch *Chain) CommitteeMultiClient() *multiclient.MultiClient {
	return multiclient.New(ch.CommitteeAPIHosts())
}

func (ch *Chain) AllNodesMultiClient() *multiclient.MultiClient {
	return multiclient.New(ch.AllAPIHosts())
}

func (ch *Chain) DeployContract(name, progHashStr, description string, initParams map[string]interface{}) (*iotago.Transaction, error) {
	programHash, err := hashing.HashValueFromHex(progHashStr)
	if err != nil {
		return nil, err
	}

	params := map[string]interface{}{
		root.ParamName:        name,
		root.ParamProgramHash: programHash,
		root.ParamDescription: description,
	}
	for k, v := range initParams {
		params[k] = v
	}
	tx, err := ch.OriginatorClient().Post1Request(
		root.Contract.Hname(),
		root.FuncDeployContract.Hname(),
		chainclient.PostRequestParams{
			Args: codec.MakeDict(params),
		},
	)
	if err != nil {
		return nil, err
	}
	_, err = ch.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(ch.ChainID, tx, 30*time.Second)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (ch *Chain) DeployWasmContract(name, description string, progBinary []byte, initParams map[string]interface{}) (hashing.HashValue, error) {
	blobFieldValues := codec.MakeDict(map[string]interface{}{
		blob.VarFieldVMType:             vmtypes.WasmTime,
		blob.VarFieldProgramBinary:      progBinary,
		blob.VarFieldProgramDescription: description,
	})

	programHash, _, _, err := ch.OriginatorClient().UploadBlob(blobFieldValues)
	if err != nil {
		return hashing.NilHash, err
	}

	progBinaryBack, err := ch.GetBlobFieldValue(programHash, blob.VarFieldProgramBinary)
	if err != nil {
		return hashing.NilHash, err
	}
	if !bytes.Equal(progBinary, progBinaryBack) {
		return hashing.NilHash, fmt.Errorf("!bytes.Equal(progBinary, progBinaryBack)")
	}
	fmt.Printf("---- blob installed correctly len = %d\n", len(progBinaryBack))

	params := make(map[string]interface{})
	for k, v := range initParams {
		params[k] = v
	}
	params[root.ParamName] = name
	params[root.ParamProgramHash] = programHash
	params[root.ParamDescription] = description

	args := codec.MakeDict(params)
	tx, err := ch.OriginatorClient().Post1Request(
		root.Contract.Hname(),
		root.FuncDeployContract.Hname(),
		chainclient.PostRequestParams{
			Args: args,
		},
	)
	if err != nil {
		return hashing.NilHash, err
	}
	_, err = ch.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(ch.ChainID, tx, 30*time.Second)
	if err != nil {
		return hashing.NilHash, err
	}

	return programHash, nil
}

func (ch *Chain) GetBlobFieldValue(blobHash hashing.HashValue, field string) ([]byte, error) {
	v, err := ch.Cluster.WaspClient(0).CallView(
		ch.ChainID, blob.Contract.Hname(), blob.ViewGetBlobField.Name,
		dict.Dict{
			blob.ParamHash:  blobHash[:],
			blob.ParamField: []byte(field),
		})
	if err != nil {
		return nil, err
	}
	if v.IsEmpty() {
		return nil, nil
	}
	ret, err := v.Get(blob.ParamBytes)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (ch *Chain) StartMessageCounter(expectations map[string]int) (*MessageCounter, error) {
	return NewMessageCounter(ch.Cluster, ch.CommitteeNodes, expectations)
}

func (ch *Chain) BlockIndex(nodeIndex ...int) (uint32, error) {
	cl := ch.SCClient(blocklog.Contract.Hname(), nil, nodeIndex...)
	ret, err := cl.CallView(blocklog.ViewGetBlockInfo.Name, nil)
	if err != nil {
		return 0, err
	}
	return codec.DecodeUint32(ret.MustGet(blocklog.ParamBlockIndex), 0)
}

func (ch *Chain) GetAllBlockInfoRecordsReverse(nodeIndex ...int) ([]*blocklog.BlockInfo, error) {
	blockIndex, err := ch.BlockIndex(nodeIndex...)
	if err != nil {
		return nil, err
	}
	cl := ch.SCClient(blocklog.Contract.Hname(), nil, nodeIndex...)
	ret := make([]*blocklog.BlockInfo, 0, blockIndex+1)
	for idx := int(blockIndex); idx >= 0; idx-- {
		res, err := cl.CallView(blocklog.ViewGetBlockInfo.Name, dict.Dict{
			blocklog.ParamBlockIndex: codec.EncodeUint32(uint32(idx)),
		})
		if err != nil {
			return nil, err
		}
		bi, err := blocklog.BlockInfoFromBytes(uint32(idx), res.MustGet(blocklog.ParamBlockInfo))
		if err != nil {
			return nil, err
		}
		ret = append(ret, bi)
	}
	return ret, nil
}

func (ch *Chain) ContractRegistry(nodeIndex ...int) (map[isc.Hname]*root.ContractRecord, error) {
	cl := ch.SCClient(root.Contract.Hname(), nil, nodeIndex...)
	ret, err := cl.CallView(root.ViewGetContractRecords.Name, nil)
	if err != nil {
		return nil, err
	}
	return root.DecodeContractRegistry(collections.NewMapReadOnly(ret, root.StateVarContractRegistry))
}

func (ch *Chain) GetCounterValue(inccounterSCHname isc.Hname, nodeIndex ...int) (int64, error) {
	cl := ch.SCClient(inccounterSCHname, nil, nodeIndex...)
	ret, err := cl.CallView(inccounter.ViewGetCounter.Name, nil)
	if err != nil {
		return 0, err
	}
	return codec.DecodeInt64(ret.MustGet(inccounter.VarCounter), 0)
}

func (ch *Chain) GetStateVariable(contractHname isc.Hname, key string, nodeIndex ...int) ([]byte, error) {
	cl := ch.SCClient(contractHname, nil, nodeIndex...)
	return cl.StateGet(key)
}

func (ch *Chain) GetRequestReceipt(reqID isc.RequestID, nodeIndex ...int) (*isc.Receipt, error) {
	idx := 0
	if len(nodeIndex) > 0 {
		idx = nodeIndex[0]
	}
	rec, err := ch.Cluster.WaspClient(idx).RequestReceipt(ch.ChainID, reqID)
	return rec, err
}

func (ch *Chain) GetRequestReceiptsForBlock(blockIndex *uint32, nodeIndex ...int) ([]*blocklog.RequestReceipt, error) {
	cl := ch.SCClient(blocklog.Contract.Hname(), nil, nodeIndex...)
	params := dict.Dict{}
	if blockIndex != nil {
		params = dict.Dict{
			blocklog.ParamBlockIndex: codec.EncodeUint32(*blockIndex),
		}
	}
	res, err := cl.CallView(blocklog.ViewGetRequestReceiptsForBlock.Name, params)
	if err != nil {
		return nil, err
	}
	returnedBlockIndex := codec.MustDecodeUint32(res.MustGet(blocklog.ParamBlockIndex))
	recs := collections.NewArray16ReadOnly(res, blocklog.ParamRequestRecord)
	ret := make([]*blocklog.RequestReceipt, recs.MustLen())
	for i := range ret {
		data, err := recs.GetAt(uint16(i))
		if err != nil {
			return nil, err
		}
		ret[i], err = blocklog.RequestReceiptFromBytes(data)
		if err != nil {
			return nil, err
		}
		ret[i].WithBlockData(returnedBlockIndex, uint16(i))
	}
	return ret, nil
}
