package suiclient

import (
	"context"

	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

type BatchTransactionRequest struct {
	Signer    *sui.Address
	TxnParams []map[string]interface{}
	Gas       *sui.ObjectID // optional
	GasBudget uint64
	// txnBuilderMode // optional // FIXME SuiTransactionBlockBuilderMode
}

// TODO: execution_mode : <SuiTransactionBlockBuilderMode>
func (s *Client) BatchTransaction(
	ctx context.Context,
	req BatchTransactionRequest,
) (*suijsonrpc.TransactionBytes, error) {
	resp := suijsonrpc.TransactionBytes{}
	return &resp, s.http.CallContext(ctx, &resp, batchTransaction, req.Signer, req.TxnParams, req.Gas, req.GasBudget)
}

type MergeCoinsRequest struct {
	Signer      *sui.Address
	PrimaryCoin *sui.ObjectID
	CoinToMerge *sui.ObjectID
	Gas         *sui.ObjectID // optional
	GasBudget   *suijsonrpc.BigInt
}

// MergeCoins Create an unsigned transaction to merge multiple coins into one coin.
func (s *Client) MergeCoins(
	ctx context.Context,
	req MergeCoinsRequest,
) (*suijsonrpc.TransactionBytes, error) {
	resp := suijsonrpc.TransactionBytes{}
	return &resp, s.http.CallContext(ctx, &resp, mergeCoins, req.Signer, req.PrimaryCoin, req.CoinToMerge, req.Gas, req.GasBudget)
}

type MoveCallRequest struct {
	Signer    *sui.Address
	PackageID *sui.PackageID
	Module    string
	Function  string
	TypeArgs  []string
	Arguments []any
	Gas       *sui.ObjectID // optional
	GasBudget *suijsonrpc.BigInt
	// txnBuilderMode // optional // FIXME SuiTransactionBlockBuilderMode
}

// MoveCall Create an unsigned transaction to execute a Move call on the network, by calling the specified function in the module of a given package.
// TODO: execution_mode : <SuiTransactionBlockBuilderMode>
// `arguments: []any` *SuiAddress can be arguments here, it will automatically convert to Address in hex string.
// [][]byte can't be passed. User should encode array of hex string.
func (s *Client) MoveCall(
	ctx context.Context,
	req MoveCallRequest,
) (*suijsonrpc.TransactionBytes, error) {
	resp := suijsonrpc.TransactionBytes{}
	return &resp, s.http.CallContext(
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
	Signer     *sui.Address
	InputCoins []*sui.ObjectID
	Recipients []*sui.Address
	Amount     []*suijsonrpc.BigInt
	Gas        *sui.ObjectID // optional
	GasBudget  *suijsonrpc.BigInt
}

func (s *Client) Pay(
	ctx context.Context,
	req PayRequest,
) (*suijsonrpc.TransactionBytes, error) {
	resp := suijsonrpc.TransactionBytes{}
	return &resp, s.http.CallContext(ctx, &resp, pay, req.Signer, req.InputCoins, req.Recipients, req.Amount, req.Gas, req.GasBudget)
}

type PayAllSuiRequest struct {
	Signer     *sui.Address
	Recipient  *sui.Address
	InputCoins []*sui.ObjectID
	GasBudget  *suijsonrpc.BigInt
}

// PayAllSui Create an unsigned transaction to send all SUI coins to one recipient.
func (s *Client) PayAllSui(
	ctx context.Context,
	req PayAllSuiRequest,
) (*suijsonrpc.TransactionBytes, error) {
	resp := suijsonrpc.TransactionBytes{}
	return &resp, s.http.CallContext(ctx, &resp, payAllSui, req.Signer, req.InputCoins, req.Recipient, req.GasBudget)
}

type PaySuiRequest struct {
	Signer     *sui.Address
	InputCoins []*sui.ObjectID
	Recipients []*sui.Address
	Amount     []*suijsonrpc.BigInt
	GasBudget  *suijsonrpc.BigInt
}

// see explanation in https://forums.sui.io/t/how-to-use-the-sui-paysui-method/2282
func (s *Client) PaySui(
	ctx context.Context,
	req PaySuiRequest,
) (*suijsonrpc.TransactionBytes, error) {
	resp := suijsonrpc.TransactionBytes{}
	return &resp, s.http.CallContext(ctx, &resp, paySui, req.Signer, req.InputCoins, req.Recipients, req.Amount, req.GasBudget)
}

type PublishRequest struct {
	Sender          *sui.Address
	CompiledModules []*sui.Base64Data
	Dependencies    []*sui.ObjectID
	Gas             *sui.ObjectID // optional
	GasBudget       *suijsonrpc.BigInt
}

func (s *Client) Publish(
	ctx context.Context,
	req PublishRequest,
) (*suijsonrpc.TransactionBytes, error) {
	var resp suijsonrpc.TransactionBytes
	return &resp, s.http.CallContext(ctx, &resp, publish, req.Sender, req.CompiledModules, req.Dependencies, req.Gas, req.GasBudget)
}

type RequestAddStakeRequest struct {
	Signer    *sui.Address
	Coins     []*sui.ObjectID
	Amount    *suijsonrpc.BigInt // optional
	Validator *sui.Address
	Gas       *sui.ObjectID // optional
	GasBudget *suijsonrpc.BigInt
}

func (s *Client) RequestAddStake(
	ctx context.Context,
	req RequestAddStakeRequest,
) (*suijsonrpc.TransactionBytes, error) {
	var resp suijsonrpc.TransactionBytes
	return &resp, s.http.CallContext(ctx, &resp, requestAddStake, req.Signer, req.Coins, req.Amount, req.Validator, req.Gas, req.GasBudget)
}

type RequestWithdrawStakeRequest struct {
	Signer      *sui.Address
	StakedSuiId *sui.ObjectID
	Gas         *sui.ObjectID // optional
	GasBudget   *suijsonrpc.BigInt
}

func (s *Client) RequestWithdrawStake(
	ctx context.Context,
	req RequestWithdrawStakeRequest,
) (*suijsonrpc.TransactionBytes, error) {
	var resp suijsonrpc.TransactionBytes
	return &resp, s.http.CallContext(ctx, &resp, requestWithdrawStake, req.Signer, req.StakedSuiId, req.Gas, req.GasBudget)
}

type SplitCoinRequest struct {
	Signer       *sui.Address
	Coin         *sui.ObjectID
	SplitAmounts []*suijsonrpc.BigInt
	Gas          *sui.ObjectID // optional
	GasBudget    *suijsonrpc.BigInt
}

// SplitCoin Creates an unsigned transaction to split a coin object into multiple coins.
// better to replace with unsafe_pay API which consumes less gas
func (s *Client) SplitCoin(
	ctx context.Context,
	req SplitCoinRequest,
) (*suijsonrpc.TransactionBytes, error) {
	resp := suijsonrpc.TransactionBytes{}
	return &resp, s.http.CallContext(ctx, &resp, splitCoin, req.Signer, req.Coin, req.SplitAmounts, req.Gas, req.GasBudget)
}

type SplitCoinEqualRequest struct {
	Signer     *sui.Address
	Coin       *sui.ObjectID
	SplitCount *suijsonrpc.BigInt
	Gas        *sui.ObjectID // optional
	GasBudget  *suijsonrpc.BigInt
}

// SplitCoinEqual Creates an unsigned transaction to split a coin object into multiple equal-size coins.
// better to replace with unsafe_pay API which consumes less gas
func (s *Client) SplitCoinEqual(
	ctx context.Context,
	req SplitCoinEqualRequest,
) (*suijsonrpc.TransactionBytes, error) {
	resp := suijsonrpc.TransactionBytes{}
	return &resp, s.http.CallContext(ctx, &resp, splitCoinEqual, req.Signer, req.Coin, req.SplitCount, req.Gas, req.GasBudget)
}

type TransferObjectRequest struct {
	Signer    *sui.Address
	ObjectID  *sui.ObjectID
	Gas       *sui.ObjectID // optional
	GasBudget *suijsonrpc.BigInt
	Recipient *sui.Address
}

// TransferObject Create an unsigned transaction to transfer an object from one address to another. The object's type must allow public transfers
func (s *Client) TransferObject(
	ctx context.Context,
	req TransferObjectRequest,
) (*suijsonrpc.TransactionBytes, error) {
	resp := suijsonrpc.TransactionBytes{}
	return &resp, s.http.CallContext(ctx, &resp, transferObject, req.Signer, req.ObjectID, req.Gas, req.GasBudget, req.Recipient)
}

type TransferSuiRequest struct {
	Signer    *sui.Address
	ObjectID  *sui.ObjectID
	GasBudget *suijsonrpc.BigInt
	Recipient *sui.Address
	Amount    *suijsonrpc.BigInt // optional
}

// TransferSui Create an unsigned transaction to send SUI coin object to a Sui address. The SUI object is also used as the gas object.
func (s *Client) TransferSui(
	ctx context.Context,
	req TransferSuiRequest,
) (*suijsonrpc.TransactionBytes, error) {
	resp := suijsonrpc.TransactionBytes{}
	return &resp, s.http.CallContext(ctx, &resp, transferSui, req.Signer, req.ObjectID, req.GasBudget, req.Recipient, req.Amount)
}
