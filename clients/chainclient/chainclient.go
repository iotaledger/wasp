package chainclient

import (
	"context"
	"math"

	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

// Client allows to interact with a specific chain in the node, for example to send on-ledger or off-ledger requests
type Client struct {
	L2Client     clients.L2Client
	WaspClient   *apiclient.APIClient
	ChainID      isc.ChainID
	IscPackageID iotago.PackageID
	KeyPair      cryptolib.Signer
}

// New creates a new chainclient.Client
func New(
	l2Client clients.L2Client,
	waspClient *apiclient.APIClient,
	chainID isc.ChainID,
	iscPackageID iotago.PackageID,
	keyPair cryptolib.Signer,
) *Client {
	return &Client{
		L2Client:     l2Client,
		WaspClient:   waspClient,
		ChainID:      chainID,
		IscPackageID: iscPackageID,
		KeyPair:      keyPair,
	}
}

type PostRequestParams struct {
	Transfer  *isc.Assets
	Nonce     uint64
	NFT       *isc.NFT
	Allowance *isc.Assets
	gasBudget uint64
}

func (par *PostRequestParams) GasBudget() uint64 {
	if par.gasBudget == 0 {
		return math.MaxUint64
	}
	return par.gasBudget
}

func defaultParams(params ...PostRequestParams) PostRequestParams {
	if len(params) > 0 {
		return params[0]
	}
	return PostRequestParams{}
}

// PostRequest sends an on-ledger transaction with one request on it to the chain
func (c *Client) PostRequest(
	ctx context.Context,
	msg isc.Message,
	param PostRequestParams,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	return c.postSingleRequest(ctx, msg, param)
}

// PostNRequest sends n consecutive on-ledger transactions with one request on each, to the chain
func (c *Client) PostMultipleRequests(
	ctx context.Context,
	msg isc.Message,
	requestsCount int,
	params ...PostRequestParams,
) ([]*iotajsonrpc.IotaTransactionBlockResponse, error) {
	var err error
	txRes := make([]*iotajsonrpc.IotaTransactionBlockResponse, requestsCount)
	for i := 0; i < requestsCount; i++ {
		txRes[i], err = c.postSingleRequest(ctx, msg, params[i])
		if err != nil {
			return nil, err
		}
	}
	return txRes, nil
}

func (c *Client) postSingleRequest(
	ctx context.Context,
	iscmsg isc.Message,
	params PostRequestParams,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	assets := iscmove.NewAssets(0)
	for cointype, coinbal := range params.Transfer.Coins {
		assets.AddCoin(cointype.String(), iotajsonrpc.CoinValue(coinbal.Uint64()))
	}
	msg := &iscmove.Message{
		Contract: uint32(iscmsg.Target.Contract),
		Function: uint32(iscmsg.Target.EntryPoint),
		Args:     iscmsg.Params,
	}
	allowances := iscmove.NewAssets(0)
	for cointype, coinbal := range params.Allowance.Coins {
		allowances.AddCoin(cointype.String(), iotajsonrpc.CoinValue(coinbal))
	}
	return c.L2Client.CreateAndSendRequestWithAssets(
		ctx,
		c.KeyPair,
		c.IscPackageID,
		c.ChainID.AsAddress().AsIotaAddress(),
		assets,
		msg,
		allowances,
		params.gasBudget,
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
		false,
	)
}

func (c *Client) ISCNonce(ctx context.Context) (uint64, error) {
	var agentID isc.AgentID = isc.NewAddressAgentID(c.KeyPair.Address())

	msg := accounts.ViewGetAccountNonce.Message(&agentID)
	msg.Target.Contract.String()
	result, _, err := c.WaspClient.ChainsApi.CallView(ctx, c.ChainID.String()).
		ContractCallViewRequest(apiextensions.CallViewReq(accounts.ViewGetAccountNonce.Message(&agentID))).
		Execute()
	if err != nil {
		return 0, err
	}
	resultDict, err := apiextensions.APIResultToCallArgs(result)
	if err != nil {
		return 0, err
	}

	nonce, err := accounts.ViewGetAccountNonce.DecodeOutput(resultDict)
	if err != nil {
		return 0, err
	}

	return nonce, nil
}

// PostOffLedgerRequest sends an off-ledger tx via the wasp node web api
func (c *Client) PostOffLedgerRequest(
	ctx context.Context,
	msg isc.Message,
	params ...PostRequestParams,
) (isc.OffLedgerRequest, error) {
	par := defaultParams(params...)
	if par.Nonce == 0 {
		nonce, err := c.ISCNonce(ctx)
		if err != nil {
			return nil, err
		}
		par.Nonce = nonce
	}
	req := isc.NewOffLedgerRequest(c.ChainID, msg, par.Nonce, par.GasBudget())
	req.WithAllowance(par.Allowance)
	req.WithNonce(par.Nonce)
	signed := req.Sign(c.KeyPair)

	request := cryptolib.EncodeHex(signed.Bytes())

	offLedgerRequest := apiclient.OffLedgerRequest{
		ChainId: c.ChainID.String(),
		Request: request,
	}
	_, err := c.WaspClient.RequestsApi.
		OffLedger(ctx).
		OffLedgerRequest(offLedgerRequest).
		Execute()

	return signed, err
}

func (c *Client) DepositFunds(n coin.Value) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	return c.PostRequest(context.Background(), accounts.FuncDeposit.Message(), PostRequestParams{
		Transfer: isc.NewAssets(n),
	})
}

func NewPostRequestParams() *PostRequestParams {
	return &PostRequestParams{
		Transfer:  isc.NewEmptyAssets(),
		Allowance: isc.NewEmptyAssets(),
	}
}

func (par *PostRequestParams) WithTransfer(transfer *isc.Assets) *PostRequestParams {
	par.Transfer = transfer
	return par
}

func (par *PostRequestParams) WithBaseTokens(i coin.Value) *PostRequestParams {
	par.Transfer.AddBaseTokens(i)
	return par
}

func (par *PostRequestParams) WithGasBudget(budget uint64) *PostRequestParams {
	par.gasBudget = budget
	return par
}
