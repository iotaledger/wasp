package solo

import (
	"math"
	"math/big"
	"sort"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

// L2Accounts returns all accounts on the chain with non-zero balances
func (ch *Chain) L2Accounts() []isc.AgentID {
	d := accounts.NewStateAccess(lo.Must(ch.Store().LatestState())).AllAccounts()
	keys := d.KeysSorted()
	ret := make([]isc.AgentID, 0, len(keys)-1)
	for _, key := range keys {
		aid, err := accounts.AgentIDFromKey(key, ch.ChainID)
		require.NoError(ch.Env.T, err)
		ret = append(ret, aid)
	}
	return ret
}

func (ch *Chain) parseAccountBalance(d dict.Dict, err error) *isc.Assets {
	require.NoError(ch.Env.T, err)
	if d.IsEmpty() {
		return isc.NewEmptyAssets()
	}
	ret, err := isc.AssetsFromDict(d)
	require.NoError(ch.Env.T, err)
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
	assets := lo.Must(accounts.ViewBalance.Output.Decode(
		lo.Must(ch.CallViewAtState(chainState, accounts.ViewBalance.Message(&agentID))),
	))
	assets.NFTs = ch.L2NFTs(agentID)
	return assets
}

func (ch *Chain) L2BaseTokens(agentID isc.AgentID) uint64 {
	return ch.L2Assets(agentID).BaseTokens
}

func (ch *Chain) L2BaseTokensAtStateIndex(agentID isc.AgentID, stateIndex uint32) uint64 {
	return ch.L2AssetsAtStateIndex(agentID, stateIndex).BaseTokens
}

func (ch *Chain) L2NFTs(agentID isc.AgentID) []iotago.NFTID {
	res, err := ch.CallView(accounts.ViewAccountNFTs.Message(&agentID))
	require.NoError(ch.Env.T, err)
	return lo.Must(accounts.ViewAccountNFTs.Output.Decode(res))
}

func (ch *Chain) L2NativeTokens(agentID isc.AgentID, nativeTokenID iotago.NativeTokenID) *big.Int {
	return ch.L2Assets(agentID).AmountNativeToken(nativeTokenID)
}

func (ch *Chain) L2CommonAccountAssets() *isc.Assets {
	return ch.L2Assets(accounts.CommonAccount())
}

func (ch *Chain) L2CommonAccountBaseTokens() uint64 {
	return ch.L2Assets(accounts.CommonAccount()).BaseTokens
}

func (ch *Chain) L2CommonAccountNativeTokens(nativeTokenID iotago.NativeTokenID) *big.Int {
	return ch.L2Assets(accounts.CommonAccount()).AmountNativeToken(nativeTokenID)
}

// L2TotalAssets return total sum of ftokens contained in the on-chain accounts
func (ch *Chain) L2TotalAssets() *isc.Assets {
	r, err := ch.CallView(accounts.ViewTotalAssets.Message())
	require.NoError(ch.Env.T, err)
	return lo.Must(accounts.ViewTotalAssets.Output1.Decode(r))
}

// L2TotalBaseTokens return total sum of base tokens in L2 (all accounts)
func (ch *Chain) L2TotalBaseTokens() uint64 {
	return ch.L2TotalAssets().BaseTokens
}

func (ch *Chain) GetOnChainTokenIDs() []iotago.NativeTokenID {
	res, err := ch.CallView(accounts.ViewGetNativeTokenIDRegistry.Message())
	require.NoError(ch.Env.T, err)
	return lo.Must(accounts.ViewGetNativeTokenIDRegistry.Output1.Decode(res))
}

func (ch *Chain) GetFoundryOutput(sn uint32) (*iotago.FoundryOutput, error) {
	res, err := ch.CallView(accounts.ViewNativeToken.Message(sn))
	if err != nil {
		return nil, err
	}
	out, err := accounts.ViewNativeToken.Output.Decode(res)
	if err != nil {
		return nil, err
	}
	return out.(*iotago.FoundryOutput), nil
}

func (ch *Chain) GetNativeTokenIDByFoundrySN(sn uint32) (iotago.NativeTokenID, error) {
	o, err := ch.GetFoundryOutput(sn)
	if err != nil {
		return iotago.NativeTokenID{}, err
	}
	return o.MustNativeTokenID(), nil
}

type NewNativeTokenParams struct {
	ch            *Chain
	user          *cryptolib.KeyPair
	sch           iotago.TokenScheme
	tokenName     string
	tokenSymbol   string
	tokenDecimals uint8
}

// CreateFoundryGasBudgetBaseTokens always takes 100000 base tokens as gas budget and ftokens for the call
const (
	DestroyTokensGasBudgetBaseTokens       = 1 * isc.Million
	SendToL2AccountGasBudgetBaseTokens     = 1 * isc.Million
	DestroyFoundryGasBudgetBaseTokens      = 1 * isc.Million
	TransferAllowanceToGasBudgetBaseTokens = 1 * isc.Million
)

func (ch *Chain) NewNativeTokenParams(maxSupply *big.Int) *newNativeTokenParams { //nolint:revive
	ret := &newNativeTokenParams{
		ch: ch,
		sch: &iotago.SimpleTokenScheme{
			MaximumSupply: maxSupply,
			MeltedTokens:  big.NewInt(0),
			MintedTokens:  big.NewInt(0),
		},
		tokenSymbol:   "TST",
		tokenName:     "TEST",
		tokenDecimals: uint8(8),
	}
	return ret
}

func (fp *NewNativeTokenParams) WithUser(user *cryptolib.KeyPair) *NewNativeTokenParams {
	fp.user = user
	return fp
}

func (fp *NewNativeTokenParams) WithTokenScheme(sch iotago.TokenScheme) *NewNativeTokenParams {
	fp.sch = sch
	return fp
}

func (fp *NewNativeTokenParams) WithTokenName(tokenName string) *NewNativeTokenParams {
	fp.tokenName = tokenName
	return fp
}

func (fp *NewNativeTokenParams) WithTokenSymbol(tokenSymbol string) *NewNativeTokenParams {
	fp.tokenSymbol = tokenSymbol
	return fp
}

func (fp *NewNativeTokenParams) WithTokenDecimals(tokenDecimals uint8) *NewNativeTokenParams {
	fp.tokenDecimals = tokenDecimals
	return fp
}

const (
	allowanceForFoundryStorageDeposit = 1 * isc.Million
	allowanceForModifySupply          = 1 * isc.Million
)

func (fp *newNativeTokenParams) CreateFoundry() (uint32, iotago.NativeTokenID, error) {
	var sch *iotago.TokenScheme
	if fp.sch != nil {
		sch = &fp.sch
	}

	metadata := isc.NewIRC30NativeTokenMetadata(fp.tokenName, fp.tokenSymbol, fp.tokenDecimals)

	req := NewCallParams(accounts.FuncNativeTokenCreate.Message(metadata, sch)).
		WithAllowance(isc.NewAssetsBaseTokensU64(allowanceForFoundryStorageDeposit)).
		AddBaseTokens(allowanceForFoundryStorageDeposit).
		WithMaxAffordableGasBudget()

	_, estimate, err := fp.ch.EstimateGasOnLedger(req, fp.user)
	if err != nil {
		return 0, iotago.NativeTokenID{}, err
	}
	req.WithGasBudget(estimate.GasBurned)
	res, err := fp.ch.PostRequestSync(req, fp.user)
	if err != nil {
		return 0, iotago.NativeTokenID{}, err
	}
	resDeco := kvdecoder.New(res)
	retSN := resDeco.MustGetUint32(accounts.ParamFoundrySN)
	nativeTokenID, err := fp.ch.GetNativeTokenIDByFoundrySN(retSN)
	return retSN, nativeTokenID, err
}

func (ch *Chain) DestroyFoundry(sn uint32, user cryptolib.Signer) error {
	req := NewCallParams(accounts.FuncNativeTokenDestroy.Message(sn)).
		WithGasBudget(DestroyFoundryGasBudgetBaseTokens)
	_, err := ch.PostRequestSync(req, user)
	return err
}

func (ch *Chain) MintTokens(sn uint32, amount *big.Int, user *cryptolib.KeyPair) error {
	req := NewCallParams(accounts.FuncNativeTokenModifySupply.MintTokens(sn, amount)).
		AddBaseTokens(allowanceForModifySupply).
		WithAllowance(isc.NewAssetsBaseTokensU64(allowanceForModifySupply)). // enough allowance is needed for the storage deposit when token is minted first on the chain
		WithMaxAffordableGasBudget()

	_, rec, err := ch.EstimateGasOnLedger(req, user)
	if err != nil {
		return err
	}

	req.WithGasBudget(rec.GasBurned)
	_, err = ch.PostRequestSync(req, user)
	return err
}

// DestroyTokensOnL2 destroys tokens (identified by foundry SN) on user's on-chain account
func (ch *Chain) DestroyTokensOnL2(nativeTokenID iotago.NativeTokenID, amount *big.Int, user *cryptolib.KeyPair) error {
	req := NewCallParams(accounts.FuncNativeTokenModifySupply.DestroyTokens(nativeTokenID.FoundrySerialNumber(), amount)).
		WithAllowance(
			isc.NewAssets(0, iotago.NativeTokens{&iotago.NativeToken{
				ID:     nativeTokenID,
				Amount: amount,
			}}),
		).
		WithGasBudget(DestroyTokensGasBudgetBaseTokens)
	_, err := ch.PostRequestSync(req, user)
	return err
}

// DestroyTokensOnL1 sends tokens as ftokens and destroys in the same transaction
func (ch *Chain) DestroyTokensOnL1(nativeTokenID iotago.NativeTokenID, amount *big.Int, user cryptolib.Signer) error {
	req := NewCallParams(accounts.FuncNativeTokenModifySupply.DestroyTokens(nativeTokenID.FoundrySerialNumber(), amount)).
		WithMaxAffordableGasBudget().AddBaseTokens(1000)
	req.AddNativeTokens(nativeTokenID, amount)
	req.AddAllowanceNativeTokens(nativeTokenID, amount)
	_, err := ch.PostRequestSync(req, user)
	return err
}

// DepositAssetsToL2 deposits ftokens on user's on-chain account, if user is nil, then chain owner is assigned
func (ch *Chain) DepositAssetsToL2(assets *isc.Assets, user cryptolib.Signer) error {
	_, err := ch.PostRequestSync(
		NewCallParams(accounts.FuncDeposit.Message()).
			WithFungibleTokens(assets).
			WithGasBudget(math.MaxUint64),
		user,
	)
	return err
}

// TransferAllowanceTo sends an on-ledger request to transfer funds to target account (sends extra base tokens to the sender account to cover gas)
func (ch *Chain) TransferAllowanceTo(
	allowance *isc.Assets,
	targetAccount isc.AgentID,
	wallet cryptolib.Signer,
	nft ...*isc.NFT,
) error {
	callParams := NewCallParams(accounts.FuncTransferAllowanceTo.Message(targetAccount)).
		WithAllowance(allowance).
		WithFungibleTokens(allowance.Clone().AddBaseTokens(TransferAllowanceToGasBudgetBaseTokens)).
		WithGasBudget(math.MaxUint64)

	if len(nft) > 0 {
		callParams.WithNFT(nft[0])
	}
	_, err := ch.PostRequestSync(callParams, wallet)
	return err
}

// DepositBaseTokensToL2 deposits ftokens on user's on-chain account
func (ch *Chain) DepositBaseTokensToL2(amount uint64, user cryptolib.Signer) error {
	return ch.DepositAssetsToL2(isc.NewAssets(amount, nil), user)
}

func (ch *Chain) MustDepositBaseTokensToL2(amount uint64, user cryptolib.Signer) {
	err := ch.DepositBaseTokensToL2(amount, user)
	require.NoError(ch.Env.T, err)
}

func (ch *Chain) DepositNFT(nft *isc.NFT, to isc.AgentID, owner cryptolib.Signer) error {
	return ch.TransferAllowanceTo(
		isc.NewEmptyAssets().AddNFTs(nft.ID),
		to,
		owner,
		nft,
	)
}

func (ch *Chain) MustDepositNFT(nft *isc.NFT, to isc.AgentID, owner cryptolib.Signer) {
	err := ch.DepositNFT(nft, to, owner)
	require.NoError(ch.Env.T, err)
}

// Withdraw sends assets from the L2 account to L1
func (ch *Chain) Withdraw(assets *isc.Assets, user cryptolib.Signer) error {
	req := NewCallParams(accounts.FuncWithdraw.Message()).
		AddAllowance(assets).
		WithGasBudget(math.MaxUint64)
	if assets.BaseTokens == 0 {
		req.AddAllowance(isc.NewAssetsBaseTokensU64(1 * isc.Million)) // for storage deposit
	}
	_, err := ch.PostRequestOffLedger(req, user)
	return err
}

// SendFromL1ToL2Account sends ftokens from L1 address to the target account on L2
// Sender pays the gas fee
func (ch *Chain) SendFromL1ToL2Account(totalBaseTokens uint64, toSend *isc.Assets, target isc.AgentID, user cryptolib.Signer) error {
	require.False(ch.Env.T, toSend.IsEmpty())
	sumAssets := toSend.Clone().AddBaseTokens(totalBaseTokens)
	_, err := ch.PostRequestSync(
		NewCallParams(accounts.FuncTransferAllowanceTo.Message(target)).
			AddFungibleTokens(sumAssets).
			AddAllowance(toSend).
			WithGasBudget(math.MaxUint64),
		user,
	)
	return err
}

func (ch *Chain) SendFromL1ToL2AccountBaseTokens(totalBaseTokens, baseTokensSend uint64, target isc.AgentID, user cryptolib.Signer) error {
	return ch.SendFromL1ToL2Account(totalBaseTokens, isc.NewAssetsBaseTokensU64(baseTokensSend), target, user)
}

// SendFromL2ToL2Account moves ftokens on L2 from user's account to the target
func (ch *Chain) SendFromL2ToL2Account(transfer *isc.Assets, target isc.AgentID, user cryptolib.Signer) error {
	req := NewCallParams(accounts.FuncTransferAllowanceTo.Message(target)).
		AddBaseTokens(SendToL2AccountGasBudgetBaseTokens).
		AddAllowance(transfer).
		WithGasBudget(SendToL2AccountGasBudgetBaseTokens)
	_, err := ch.PostRequestSync(req, user)
	return err
}

func (ch *Chain) SendFromL2ToL2AccountBaseTokens(baseTokens uint64, target isc.AgentID, user cryptolib.Signer) error {
	return ch.SendFromL2ToL2Account(isc.NewAssetsBaseTokensU64(baseTokens), target, user)
}

func (ch *Chain) SendFromL2ToL2AccountNativeTokens(id iotago.NativeTokenID, target isc.AgentID, amount *big.Int, user cryptolib.Signer) error {
	transfer := isc.NewEmptyAssets()
	transfer.AddNativeTokens(id, amount)
	return ch.SendFromL2ToL2Account(transfer, target, user)
}
