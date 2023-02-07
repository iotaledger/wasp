package cluster

import (
	"bytes"
	"context"
	"fmt"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/clients/multiclient"
	"github.com/iotaledger/wasp/clients/scclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
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

	ChainID isc.ChainID

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
	var resolver multiclient.ClientResolver = func(apiHost string) *apiclient.APIClient {
		return ch.Cluster.WaspClientFromHostName(apiHost)
	}

	return multiclient.New(resolver, ch.CommitteeAPIHosts()) //.WithLogFunc(ch.Cluster.t.Logf)
}

func (ch *Chain) AllNodesMultiClient() *multiclient.MultiClient {
	var resolver multiclient.ClientResolver = func(apiHost string) *apiclient.APIClient {
		return ch.Cluster.WaspClientFromHostName(apiHost)
	}

	return multiclient.New(resolver, ch.AllAPIHosts()) //.WithLogFunc(ch.Cluster.t.Logf)
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

	programHash, _, _, err := ch.OriginatorClient().UploadBlob(context.Background(), blobFieldValues)
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
	v, _, err := ch.Cluster.WaspClient(0).CorecontractsApi.
		BlobsGetBlobValue(context.Background(), ch.ChainID.String(), blobHash.Hex(), field).
		Execute()

	if err != nil {
		return nil, err
	}

	decodedValue, err := iotago.DecodeHex(v.ValueData)
	if err != nil {
		return nil, err
	}

	return decodedValue, nil
}

func (ch *Chain) BlockIndex(nodeIndex ...int) (uint32, error) {
	blockInfo, _, err := ch.Cluster.
		WaspClient(nodeIndex...).CorecontractsApi.BlocklogGetLatestBlockInfo(context.Background(), ch.ChainID.String()).
		Execute()

	if err != nil {
		return 0, err
	}

	return blockInfo.BlockIndex, nil
}

func (ch *Chain) GetAllBlockInfoRecordsReverse(nodeIndex ...int) ([]*apiclient.BlockInfoResponse, error) {
	blockIndex, err := ch.BlockIndex(nodeIndex...)
	if err != nil {
		return nil, err
	}
	ret := make([]*apiclient.BlockInfoResponse, 0, blockIndex+1)
	for idx := int(blockIndex); idx >= 0; idx-- {
		blockInfo, _, err := ch.Cluster.
			WaspClient(nodeIndex...).CorecontractsApi.BlocklogGetBlockInfo(context.Background(), ch.ChainID.String(), uint32(idx)).
			Execute()

		if err != nil {
			return nil, err
		}

		ret = append(ret, blockInfo)
	}
	return ret, nil
}

func (ch *Chain) ContractRegistry(nodeIndex ...int) ([]apiclient.ContractInfoResponse, error) {
	contracts, _, err := ch.Cluster.
		WaspClient(nodeIndex...).ChainsApi.GetContracts(context.Background(), ch.ChainID.String()).
		Execute()

	if err != nil {
		return nil, err
	}

	return contracts, nil
}

func (ch *Chain) GetCounterValue(inccounterSCHname isc.Hname, nodeIndex ...int) (int64, error) {
	result, _, err := ch.Cluster.
		WaspClient(nodeIndex...).RequestsApi.CallView(context.Background()).ContractCallViewRequest(apiclient.ContractCallViewRequest{
		ChainId:       ch.ChainID.String(),
		ContractHName: inccounterSCHname.String(),
		FunctionHName: inccounter.ViewGetCounter.Hname().String(),
	}).Execute()
	if err != nil {
		return 0, err
	}

	parsedDict, err := apiextensions.APIJsonDictToDict(*result)
	if err != nil {
		return 0, err
	}

	return codec.DecodeInt64(parsedDict.MustGet(inccounter.VarCounter), 0)
}

func (ch *Chain) GetStateVariable(contractHname isc.Hname, key string, nodeIndex ...int) ([]byte, error) {
	cl := ch.SCClient(contractHname, nil, nodeIndex...)

	return cl.StateGet(context.Background(), key)
}

func (ch *Chain) GetRequestReceipt(reqID isc.RequestID, nodeIndex ...int) (*apiclient.ReceiptResponse, error) {
	receipt, _, err := ch.Cluster.WaspClient(nodeIndex...).RequestsApi.GetReceipt(context.Background(), ch.ChainID.String(), reqID.String()).
		Execute()

	return receipt, err
}

func (ch *Chain) GetRequestReceiptsForBlock(blockIndex *uint32, nodeIndex ...int) ([]apiclient.RequestReceiptResponse, error) {
	var err error
	var receipts *apiclient.BlockReceiptsResponse

	if blockIndex != nil {
		receipts, _, err = ch.Cluster.WaspClient(nodeIndex...).CorecontractsApi.BlocklogGetRequestReceiptsOfBlock(context.Background(), ch.ChainID.String(), *blockIndex).
			Execute()
	} else {
		receipts, _, err = ch.Cluster.WaspClient(nodeIndex...).CorecontractsApi.BlocklogGetRequestReceiptsOfLatestBlock(context.Background(), ch.ChainID.String()).
			Execute()
	}

	if err != nil {
		return nil, err
	}

	return receipts.Receipts, nil
}
