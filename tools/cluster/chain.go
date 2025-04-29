package cluster

import (
	"context"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/clients/multiclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/contracts/inccounter"
)

type Chain struct {
	OriginatorKeyPair *cryptolib.KeyPair

	AllPeers       []int
	CommitteeNodes []int
	Quorum         uint16
	StateAddress   *cryptolib.Address

	ChainID isc.ChainID

	Cluster *Cluster
}

func (ch *Chain) ChainAddress() *cryptolib.Address {
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

func (ch *Chain) OriginatorAddress() *cryptolib.Address {
	addr := ch.OriginatorKeyPair.Address()
	return addr
}

func (ch *Chain) OriginatorID() isc.AgentID {
	ret := isc.NewAddressAgentID(ch.OriginatorAddress())
	return ret
}

func (ch *Chain) OriginatorClient() *chainclient.Client {
	return ch.Client(ch.OriginatorKeyPair)
}

func (ch *Chain) Client(keyPair cryptolib.Signer, nodeIndex ...int) *chainclient.Client {
	idx := 0
	if len(nodeIndex) == 1 {
		idx = nodeIndex[0]
	}
	return chainclient.New(
		ch.Cluster.L1Client(),
		ch.Cluster.WaspClient(idx),
		ch.ChainID,
		ch.Cluster.Config.ISCPackageID(),
		keyPair,
	)
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

func (ch *Chain) BlockIndex(nodeIndex ...int) (uint32, error) {
	blockInfo, _, err := ch.Cluster.
		WaspClient(nodeIndex...).CorecontractsAPI.BlocklogGetLatestBlockInfo(context.Background()).
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
			WaspClient(nodeIndex...).CorecontractsAPI.BlocklogGetBlockInfo(context.Background(), uint32(idx)).
			Execute() //nolint:bodyclose // false positive
		if err != nil {
			return nil, err
		}

		ret = append(ret, blockInfo)
	}
	return ret, nil
}

func (ch *Chain) GetCounterValue(nodeIndex ...int) (int64, error) {
	result, _, err := ch.Cluster.
		WaspClient(nodeIndex...).ChainsAPI.CallView(context.Background()).
		ContractCallViewRequest(apiclient.ContractCallViewRequest{
			ContractHName: inccounter.Contract.Hname().String(),
			FunctionHName: inccounter.ViewGetCounter.Hname().String(),
		}).Execute() //nolint:bodyclose // false positive
	if err != nil {
		return 0, err
	}

	parsedDict, err := apiextensions.APIResultToCallArgs(result)
	if err != nil {
		return 0, err
	}

	return inccounter.ViewGetCounter.DecodeOutput(parsedDict)
}

func (ch *Chain) GetStateVariable(contractHname isc.Hname, key string, nodeIndex ...int) ([]byte, error) {
	cl := ch.Client(nil, nodeIndex...)
	return cl.ContractStateGet(context.Background(), contractHname, key)
}

func (ch *Chain) GetRequestReceipt(reqID isc.RequestID, nodeIndex ...int) (*apiclient.ReceiptResponse, error) {
	receipt, _, err := ch.Cluster.WaspClient(nodeIndex...).ChainsAPI.GetReceipt(context.Background(), reqID.String()).
		Execute() //nolint:bodyclose // false positive

	return receipt, err
}

func (ch *Chain) GetRequestReceiptsForBlock(blockIndex *uint32, nodeIndex ...int) ([]apiclient.ReceiptResponse, error) {
	var err error
	var receipts []apiclient.ReceiptResponse

	if blockIndex != nil {
		receipts, _, err = ch.Cluster.WaspClient(nodeIndex...).CorecontractsAPI.BlocklogGetRequestReceiptsOfBlock(context.Background(), *blockIndex).
			Execute() //nolint:bodyclose // false positive
	} else {
		receipts, _, err = ch.Cluster.WaspClient(nodeIndex...).CorecontractsAPI.BlocklogGetRequestReceiptsOfLatestBlock(context.Background()).
			Execute() //nolint:bodyclose // false positive
	}

	if err != nil {
		return nil, err
	}

	return receipts, nil
}
