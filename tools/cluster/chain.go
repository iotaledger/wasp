package cluster

import (
	"bytes"
	"context"
	"fmt"
	"time"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/hexutil"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/clients/multiclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type Chain struct {
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

func (ch *Chain) Client(keyPair cryptolib.VariantKeyPair, nodeIndex ...int) *chainclient.Client {
	idx := 0
	if len(nodeIndex) == 1 {
		idx = nodeIndex[0]
	}
	return chainclient.New(
		ch.Cluster.L1Client(),
		ch.Cluster.WaspClient(idx),
		ch.ChainID,
		keyPair,
	)
}

func (ch *Chain) CommitteeMultiClient() *multiclient.MultiClient {
	var resolver multiclient.ClientResolver = func(apiHost string) *apiclient.APIClient {
		return ch.Cluster.WaspClientFromHostName(apiHost)
	}

	return multiclient.New(resolver, ch.CommitteeAPIHosts(), ch.Cluster.l1.APIProvider().LatestAPI()) //.WithLogFunc(ch.Cluster.t.Logf)
}

func (ch *Chain) AllNodesMultiClient() *multiclient.MultiClient {
	var resolver multiclient.ClientResolver = func(apiHost string) *apiclient.APIClient {
		return ch.Cluster.WaspClientFromHostName(apiHost)
	}

	return multiclient.New(resolver, ch.AllAPIHosts(), ch.Cluster.l1.APIProvider().LatestAPI()) //.WithLogFunc(ch.Cluster.t.Logf)
}

func (ch *Chain) DeployContract(name, progHashStr string, initParams dict.Dict) (*iotago.SignedTransaction, error) {
	programHash, err := hashing.HashValueFromHex(progHashStr)
	if err != nil {
		return nil, err
	}

	tx, err := ch.OriginatorClient().PostRequest(
		root.FuncDeployContract.Message(name, programHash, initParams),
	)
	if err != nil {
		return nil, err
	}
	_, err = ch.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(ch.ChainID, tx, false, 30*time.Second)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (ch *Chain) DeployWasmContract(name string, progBinary []byte, initParams dict.Dict) (hashing.HashValue, error) {
	blobFieldValues := codec.MakeDict(map[string]interface{}{
		blob.VarFieldVMType:        vmtypes.WasmTime,
		blob.VarFieldProgramBinary: progBinary,
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

	tx, err := ch.OriginatorClient().PostRequest(
		root.FuncDeployContract.Message(name, programHash, initParams),
	)
	if err != nil {
		return hashing.NilHash, err
	}
	_, err = ch.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(ch.ChainID, tx, false, 30*time.Second)
	if err != nil {
		return hashing.NilHash, err
	}

	return programHash, nil
}

func (ch *Chain) GetBlobFieldValue(blobHash hashing.HashValue, field string) ([]byte, error) {
	v, _, err := ch.Cluster.WaspClient(0).CorecontractsApi.
		BlobsGetBlobValue(context.Background(), ch.ChainID.Bech32(ch.Cluster.l1.Bech32HRP()), blobHash.Hex(), field).
		Execute() //nolint:bodyclose // false positive
	if err != nil {
		return nil, err
	}

	decodedValue, err := hexutil.DecodeHex(v.ValueData)
	if err != nil {
		return nil, err
	}

	return decodedValue, nil
}

func (ch *Chain) BlockIndex(nodeIndex ...int) (uint32, error) {
	blockInfo, _, err := ch.Cluster.
		WaspClient(nodeIndex...).CorecontractsApi.BlocklogGetLatestBlockInfo(context.Background(), ch.ChainID.Bech32(ch.Cluster.l1.Bech32HRP())).
		Execute() //nolint:bodyclose // false positive
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
			WaspClient(nodeIndex...).CorecontractsApi.BlocklogGetBlockInfo(context.Background(), ch.ChainID.Bech32(ch.Cluster.l1.Bech32HRP()), uint32(idx)).
			Execute() //nolint:bodyclose // false positive
		if err != nil {
			return nil, err
		}

		ret = append(ret, blockInfo)
	}
	return ret, nil
}

func (ch *Chain) ContractRegistry(nodeIndex ...int) ([]apiclient.ContractInfoResponse, error) {
	contracts, _, err := ch.Cluster.
		WaspClient(nodeIndex...).ChainsApi.GetContracts(context.Background(), ch.ChainID.Bech32(ch.Cluster.l1.Bech32HRP())).
		Execute() //nolint:bodyclose // false positive
	if err != nil {
		return nil, err
	}

	return contracts, nil
}

func (ch *Chain) GetCounterValue(nodeIndex ...int) (int64, error) {
	result, _, err := ch.Cluster.
		WaspClient(nodeIndex...).ChainsApi.CallView(context.Background(), ch.ChainID.Bech32(ch.Cluster.l1.Bech32HRP())).
		ContractCallViewRequest(apiclient.ContractCallViewRequest{
			ContractHName: inccounter.Contract.Hname().String(),
			FunctionHName: inccounter.ViewGetCounter.Hname().String(),
		}).Execute() //nolint:bodyclose // false positive
	if err != nil {
		return 0, err
	}

	parsedDict, err := apiextensions.APIJsonDictToDict(*result)
	if err != nil {
		return 0, err
	}

	return inccounter.ViewGetCounter.Output.Decode(parsedDict)
}

func (ch *Chain) GetStateVariable(contractHname isc.Hname, key string, nodeIndex ...int) ([]byte, error) {
	cl := ch.Client(nil, nodeIndex...)
	return cl.ContractStateGet(context.Background(), contractHname, key)
}

func (ch *Chain) GetRequestReceipt(reqID isc.RequestID, nodeIndex ...int) (*apiclient.ReceiptResponse, error) {
	receipt, _, err := ch.Cluster.WaspClient(nodeIndex...).ChainsApi.GetReceipt(context.Background(), ch.ChainID.Bech32(ch.Cluster.l1.Bech32HRP()), reqID.String()).
		Execute() //nolint:bodyclose // false positive

	return receipt, err
}

func (ch *Chain) GetRequestReceiptsForBlock(blockIndex *uint32, nodeIndex ...int) ([]apiclient.ReceiptResponse, error) {
	var err error
	var receipts []apiclient.ReceiptResponse

	if blockIndex != nil {
		receipts, _, err = ch.Cluster.WaspClient(nodeIndex...).CorecontractsApi.BlocklogGetRequestReceiptsOfBlock(context.Background(), ch.ChainID.Bech32(ch.Cluster.l1.Bech32HRP()), *blockIndex).
			Execute() //nolint:bodyclose // false positive
	} else {
		receipts, _, err = ch.Cluster.WaspClient(nodeIndex...).CorecontractsApi.BlocklogGetRequestReceiptsOfLatestBlock(context.Background(), ch.ChainID.Bech32(ch.Cluster.l1.Bech32HRP())).
			Execute() //nolint:bodyclose // false positive
	}

	if err != nil {
		return nil, err
	}

	return receipts, nil
}
