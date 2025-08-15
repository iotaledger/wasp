package solo

import (
	"math"
	"sort"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
)

// L2Accounts returns all accounts on the chain with non-zero balances
func (ch *Chain) L2Accounts() []isc.AgentID {
	d := accounts.NewStateReaderFromChainState(
		ch.migrationScheme.LatestSchemaVersion(),
		lo.Must(ch.Store().LatestState()),
	).AllAccountsAsDict()
	keys := d.KeysSorted()
	ret := make([]isc.AgentID, 0, len(keys))
	for _, key := range keys {
		aid, err := accounts.AgentIDFromKey(key)
		require.NoError(ch.Env.T, err)
		ret = append(ret, aid)
	}
	return ret
}

func (ch *Chain) L2Ledger() map[string]*isc.Assets {
	accs := ch.L2Accounts()
	ret := make(map[string]*isc.Assets)
	for i := range accs {
		ret[string(accs[i].Bytes())] = ch.L2Assets(accs[i])
	}
	return ret
}

func (ch *Chain) L2LedgerString() string {
	l := ch.L2Ledger()
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

// L2Assets return all tokens contained in the on-chain account controlled by the 'agentID'
func (ch *Chain) L2Assets(agentID isc.AgentID) *isc.Assets {
	return ch.L2AssetsAtStateIndex(agentID, ch.LatestBlockIndex())
}

func (ch *Chain) L2AssetsAtStateIndex(agentID isc.AgentID, stateIndex uint32) *isc.Assets {
	chainState, err := ch.store.StateByIndex(stateIndex)
	require.NoError(ch.Env.T, err)

	res, err := ch.CallViewAtState(chainState, accounts.ViewBalance.Message(&agentID))
	require.NoError(ch.Env.T, err)

	cb := lo.Must(accounts.ViewBalance.DecodeOutput(res))

	assets := isc.NewEmptyAssets()
	for coinType, balance := range cb.Iterate() {
		assets.AddCoin(coinType, balance)
	}

	return assets
}

func (ch *Chain) L2BaseTokens(agentID isc.AgentID) coin.Value {
	return ch.L2Assets(agentID).BaseTokens()
}

func (ch *Chain) L2BaseTokensAtStateIndex(agentID isc.AgentID, stateIndex uint32) coin.Value {
	return ch.L2AssetsAtStateIndex(agentID, stateIndex).BaseTokens()
}

func (ch *Chain) L2Objects(agentID isc.AgentID) []isc.IotaObject {
	res, err := ch.CallView(accounts.ViewAccountObjects.Message(&agentID))
	require.NoError(ch.Env.T, err)
	return lo.Must(accounts.ViewAccountObjects.DecodeOutput(res))
}

func (ch *Chain) L2CoinBalance(agentID isc.AgentID, coinType coin.Type) coin.Value {
	return ch.L2Assets(agentID).CoinBalance(coinType)
}

func (ch *Chain) L2CommonAccountAssets() *isc.Assets {
	return ch.L2Assets(accounts.CommonAccount())
}

func (ch *Chain) L2CommonAccountBaseTokens() coin.Value {
	return ch.L2Assets(accounts.CommonAccount()).BaseTokens()
}

func (ch *Chain) L2CommonAccountNativeTokens(coinType coin.Type) coin.Value {
	return ch.L2Assets(accounts.CommonAccount()).CoinBalance(coinType)
}

// L2TotalAssets return total sum of ftokens contained in the on-chain accounts
func (ch *Chain) L2TotalAssets() *isc.Assets {
	r, err := ch.CallView(accounts.ViewTotalAssets.Message())
	require.NoError(ch.Env.T, err)
	cb := lo.Must(accounts.ViewTotalAssets.DecodeOutput(r))

	assets := isc.NewEmptyAssets()
	for coinType, balance := range cb.Iterate() {
		assets.AddCoin(coinType, balance)
	}

	return assets
}

// L2TotalBaseTokens return total sum of base tokens in L2 (all accounts)
func (ch *Chain) L2TotalBaseTokens() coin.Value {
	return ch.L2TotalAssets().BaseTokens()
}

// CreateFoundryGasBudgetBaseTokens always takes 100000 base tokens as gas budget and ftokens for the call
const (
	DestroyTokensGasBudgetBaseTokens       = 1 * isc.Million
	SendToL2AccountGasBudgetBaseTokens     = 1 * isc.Million
	DestroyFoundryGasBudgetBaseTokens      = 1 * isc.Million
	TransferAllowanceToGasBudgetBaseTokens = 1 * isc.Million
)

// DepositAssetsToL2 deposits ftokens on user's on-chain account, if user is
// nil, then chain admin is assigned
func (ch *Chain) DepositAssetsToL2(assets *isc.Assets, user *cryptolib.KeyPair) error {
	_, err := ch.PostRequestSync(
		NewCallParams(accounts.FuncDeposit.Message()).
			WithAssets(assets).
			WithGasBudget(math.MaxUint64),
		user,
	)
	return err
}

// TransferAllowanceTo sends an on-ledger request to transfer funds to target account (sends extra base tokens to the sender account to cover gas)
func (ch *Chain) TransferAllowanceTo(
	allowance *isc.Assets,
	targetAccount isc.AgentID,
	wallet *cryptolib.KeyPair,
) error {
	callParams := NewCallParams(accounts.FuncTransferAllowanceTo.Message(targetAccount)).
		WithAllowance(allowance).
		WithAssets(allowance.Clone().AddBaseTokens(TransferAllowanceToGasBudgetBaseTokens)).
		WithGasBudget(math.MaxUint64)
	_, err := ch.PostRequestSync(callParams, wallet)
	return err
}

// DepositBaseTokensToL2 deposits ftokens on user's on-chain account
func (ch *Chain) DepositBaseTokensToL2(amount coin.Value, user *cryptolib.KeyPair) error {
	return ch.DepositAssetsToL2(isc.NewAssets(amount), user)
}

func (ch *Chain) MustDepositBaseTokensToL2(amount coin.Value, user *cryptolib.KeyPair) {
	err := ch.DepositBaseTokensToL2(amount, user)
	require.NoError(ch.Env.T, err)
}

// Withdraw sends assets from the L2 account to L1
func (ch *Chain) Withdraw(assets *isc.Assets, user *cryptolib.KeyPair) error {
	req := NewCallParams(accounts.FuncWithdraw.Message()).
		AddAllowance(assets).
		WithGasBudget(math.MaxUint64)

	if assets.BaseTokens() == 0 {
		req.AddAllowance(isc.NewEmptyAssets().AddBaseTokens(1 * isc.Million)) // for storage deposit
	}
	_, err := ch.PostRequestOffLedger(req, user)
	return err
}

// SendFromL1ToL2Account sends ftokens from L1 address to the target account on L2
// Sender pays the gas fee
func (ch *Chain) SendFromL1ToL2Account(totalBaseTokens coin.Value, toSend isc.CoinBalances, target isc.AgentID, user *cryptolib.KeyPair) error {
	require.False(ch.Env.T, toSend.IsEmpty())
	sumAssets := toSend.Clone().AddBaseTokens(totalBaseTokens)
	_, err := ch.PostRequestSync(
		NewCallParams(accounts.FuncTransferAllowanceTo.Message(target)).
			AddFungibleTokens(sumAssets).
			AddAllowance(toSend.ToAssets()).
			WithGasBudget(math.MaxUint64),
		user,
	)
	return err
}

func (ch *Chain) SendFromL1ToL2AccountBaseTokens(totalBaseTokens, baseTokensSend coin.Value, target isc.AgentID, user *cryptolib.KeyPair) error {
	return ch.SendFromL1ToL2Account(totalBaseTokens, isc.NewCoinBalances().AddBaseTokens(baseTokensSend), target, user)
}

// SendFromL2ToL2Account moves ftokens on L2 from user's account to the target
func (ch *Chain) SendFromL2ToL2Account(transfer *isc.Assets, target isc.AgentID, user *cryptolib.KeyPair) error {
	req := NewCallParams(accounts.FuncTransferAllowanceTo.Message(target)).
		AddBaseTokens(SendToL2AccountGasBudgetBaseTokens).
		AddAllowance(transfer).
		WithGasBudget(SendToL2AccountGasBudgetBaseTokens)
	_, err := ch.PostRequestSync(req, user)
	return err
}

func (ch *Chain) SendFromL2ToL2AccountBaseTokens(baseTokens coin.Value, target isc.AgentID, user *cryptolib.KeyPair) error {
	return ch.SendFromL2ToL2Account(isc.NewEmptyAssets().AddBaseTokens(baseTokens), target, user)
}

func (ch *Chain) SendFromL2ToL2AccountNativeTokens(coinType coin.Type, target isc.AgentID, amount coin.Value, user *cryptolib.KeyPair) error {
	transfer := isc.NewEmptyAssets()
	transfer.AddCoin(coinType, amount)
	return ch.SendFromL2ToL2Account(transfer, target, user)
}
