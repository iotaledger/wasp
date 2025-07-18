package iotaclient

import (
	"context"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
)

type BatchTransactionRequest struct {
	Signer    *iotago.Address
	TxnParams []map[string]interface{}
	Gas       *iotago.ObjectID // optional
	GasBudget uint64
	// txnBuilderMode // optional // FIXME IotaTransactionBlockBuilderMode
}

// TODO: execution_mode : <IotaTransactionBlockBuilderMode>
func (c *Client) BatchTransaction(
	ctx context.Context,
	req BatchTransactionRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	resp := iotajsonrpc.TransactionBytes{}
	return &resp, c.transport.Call(ctx, &resp, batchTransaction, req.Signer, req.TxnParams, req.Gas, req.GasBudget)
}

type MergeCoinsRequest struct {
	Signer      *iotago.Address
	PrimaryCoin *iotago.ObjectID
	CoinToMerge *iotago.ObjectID
	Gas         *iotago.ObjectID // optional
	GasBudget   *iotajsonrpc.BigInt
}

// MergeCoins Create an unsigned transaction to merge multiple coins into one coin.
func (c *Client) MergeCoins(
	ctx context.Context,
	req MergeCoinsRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	resp := iotajsonrpc.TransactionBytes{}
	return &resp, c.transport.Call(
		ctx,
		&resp,
		mergeCoins,
		req.Signer,
		req.PrimaryCoin,
		req.CoinToMerge,
		req.Gas,
		req.GasBudget,
	)
}

type MoveCallRequest struct {
	Signer    *iotago.Address
	PackageID *iotago.PackageID
	Module    string
	Function  string
	TypeArgs  []string
	Arguments []any
	Gas       *iotago.ObjectID // optional
	GasBudget *iotajsonrpc.BigInt
	// txnBuilderMode // optional // FIXME IotaTransactionBlockBuilderMode
}

// MoveCall Create an unsigned transaction to execute a Move call on the network, by calling the specified function in the module of a given package.
// TODO: execution_mode : <IotaTransactionBlockBuilderMode>
// `arguments: []any` *IotaAddress can be arguments here, it will automatically convert to Address in hex string.
// [][]byte can't be passed. User should encode array of hex string.
func (c *Client) MoveCall(
	ctx context.Context,
	req MoveCallRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	resp := iotajsonrpc.TransactionBytes{}
	return &resp, c.transport.Call(
		ctx,
		&resp,
		moveCall,
		req.Signer,
		req.PackageID,
		req.Module,
		req.Function,
		req.TypeArgs,
		req.Arguments,
		req.Gas,
		req.GasBudget,
	)
}

type PayRequest struct {
	Signer     *iotago.Address
	InputCoins []*iotago.ObjectID
	Recipients []*iotago.Address
	Amount     []*iotajsonrpc.BigInt
	Gas        *iotago.ObjectID // optional
	GasBudget  *iotajsonrpc.BigInt
}

func (c *Client) Pay(
	ctx context.Context,
	req PayRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	resp := iotajsonrpc.TransactionBytes{}
	return &resp, c.transport.Call(
		ctx,
		&resp,
		pay,
		req.Signer,
		req.InputCoins,
		req.Recipients,
		req.Amount,
		req.Gas,
		req.GasBudget,
	)
}

type PayAllIotaRequest struct {
	Signer     *iotago.Address
	Recipient  *iotago.Address
	InputCoins []*iotago.ObjectID
	GasBudget  *iotajsonrpc.BigInt
}

// PayAllIota Create an unsigned transaction to send all IOTA coins to one recipient.
func (c *Client) PayAllIota(
	ctx context.Context,
	req PayAllIotaRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	resp := iotajsonrpc.TransactionBytes{}
	return &resp, c.transport.Call(ctx, &resp, payAllIota, req.Signer, req.InputCoins, req.Recipient, req.GasBudget)
}

type PayIotaRequest struct {
	Signer     *iotago.Address
	InputCoins []*iotago.ObjectID
	Recipients []*iotago.Address
	Amount     []*iotajsonrpc.BigInt
	GasBudget  *iotajsonrpc.BigInt
}

// see explanation in https://forums.iota.io/t/how-to-use-the-iota-payiota-method/2282
func (c *Client) PayIota(
	ctx context.Context,
	req PayIotaRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	resp := iotajsonrpc.TransactionBytes{}
	return &resp, c.transport.Call(
		ctx,
		&resp,
		payIota,
		req.Signer,
		req.InputCoins,
		req.Recipients,
		req.Amount,
		req.GasBudget,
	)
}

type PublishRequest struct {
	Sender          *iotago.Address
	CompiledModules []*iotago.Base64Data
	Dependencies    []*iotago.ObjectID
	Gas             *iotago.ObjectID // optional
	GasBudget       *iotajsonrpc.BigInt
}

func (c *Client) Publish(
	ctx context.Context,
	req PublishRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	var resp iotajsonrpc.TransactionBytes
	return &resp, c.transport.Call(
		ctx,
		&resp,
		publish,
		req.Sender,
		req.CompiledModules,
		req.Dependencies,
		req.Gas,
		req.GasBudget,
	)
}

type RequestAddStakeRequest struct {
	Signer    *iotago.Address
	Coins     []*iotago.ObjectID
	Amount    *iotajsonrpc.BigInt // optional
	Validator *iotago.Address
	Gas       *iotago.ObjectID // optional
	GasBudget *iotajsonrpc.BigInt
}

func (c *Client) RequestAddStake(
	ctx context.Context,
	req RequestAddStakeRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	var resp iotajsonrpc.TransactionBytes
	return &resp, c.transport.Call(
		ctx,
		&resp,
		requestAddStake,
		req.Signer,
		req.Coins,
		req.Amount,
		req.Validator,
		req.Gas,
		req.GasBudget,
	)
}

type RequestWithdrawStakeRequest struct {
	Signer       *iotago.Address
	StakedIotaID *iotago.ObjectID
	Gas          *iotago.ObjectID // optional
	GasBudget    *iotajsonrpc.BigInt
}

func (c *Client) RequestWithdrawStake(
	ctx context.Context,
	req RequestWithdrawStakeRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	var resp iotajsonrpc.TransactionBytes
	return &resp, c.transport.Call(
		ctx,
		&resp,
		requestWithdrawStake,
		req.Signer,
		req.StakedIotaID,
		req.Gas,
		req.GasBudget,
	)
}

type SplitCoinRequest struct {
	Signer       *iotago.Address
	Coin         *iotago.ObjectID
	SplitAmounts []*iotajsonrpc.BigInt
	Gas          *iotago.ObjectID // optional
	GasBudget    *iotajsonrpc.BigInt
}

// SplitCoin Creates an unsigned transaction to split a coin object into multiple coins.
// better to replace with unsafe_pay API which consumes less gas
func (c *Client) SplitCoin(
	ctx context.Context,
	req SplitCoinRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	resp := iotajsonrpc.TransactionBytes{}
	return &resp, c.transport.Call(
		ctx,
		&resp,
		splitCoin,
		req.Signer,
		req.Coin,
		req.SplitAmounts,
		req.Gas,
		req.GasBudget,
	)
}

type SplitCoinEqualRequest struct {
	Signer     *iotago.Address
	Coin       *iotago.ObjectID
	SplitCount *iotajsonrpc.BigInt
	Gas        *iotago.ObjectID // optional
	GasBudget  *iotajsonrpc.BigInt
}

// SplitCoinEqual Creates an unsigned transaction to split a coin object into multiple equal-size coins.
// better to replace with unsafe_pay API which consumes less gas
func (c *Client) SplitCoinEqual(
	ctx context.Context,
	req SplitCoinEqualRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	resp := iotajsonrpc.TransactionBytes{}
	return &resp, c.transport.Call(
		ctx,
		&resp,
		splitCoinEqual,
		req.Signer,
		req.Coin,
		req.SplitCount,
		req.Gas,
		req.GasBudget,
	)
}

type TransferObjectRequest struct {
	Signer    *iotago.Address
	ObjectID  *iotago.ObjectID
	Gas       *iotago.ObjectID // optional
	GasBudget *iotajsonrpc.BigInt
	Recipient *iotago.Address
}

// TransferObject Create an unsigned transaction to transfer an object from one address to another. The object's type must allow public transfers
func (c *Client) TransferObject(
	ctx context.Context,
	req TransferObjectRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	resp := iotajsonrpc.TransactionBytes{}
	return &resp, c.transport.Call(
		ctx,
		&resp,
		transferObject,
		req.Signer,
		req.ObjectID,
		req.Gas,
		req.GasBudget,
		req.Recipient,
	)
}

type TransferIotaRequest struct {
	Signer    *iotago.Address
	ObjectID  *iotago.ObjectID
	GasBudget *iotajsonrpc.BigInt
	Recipient *iotago.Address
	Amount    *iotajsonrpc.BigInt // optional
}

// TransferIota Create an unsigned transaction to send IOTA coin object to a Iota address.
// The IOTA object is also used as the gas object.
func (c *Client) TransferIota(
	ctx context.Context,
	req TransferIotaRequest,
) (*iotajsonrpc.TransactionBytes, error) {
	resp := iotajsonrpc.TransactionBytes{}
	return &resp, c.transport.Call(
		ctx,
		&resp,
		transferIota,
		req.Signer,
		req.ObjectID,
		req.GasBudget,
		req.Recipient,
		req.Amount,
	)
}
