package solo

import (
	"fmt"
	"math"
	"math/big"
	"sort"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

// L2Accounts returns all accounts on the chain with non-zero balances
func (ch *Chain) L2Accounts() []isc.AgentID {
	d, err := ch.CallView(accounts.Contract.Name, accounts.ViewAccounts.Name)
	require.NoError(ch.Env.T, err)
	keys := d.KeysSorted()
	ret := make([]isc.AgentID, 0, len(keys)-1)
	for _, key := range keys {
		aid, err := codec.DecodeAgentID([]byte(key))
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
	return ch.parseAccountBalance(
		ch.CallView(accounts.Contract.Name, accounts.ViewBalance.Name, accounts.ParamAgentID, agentID),
	)
}

func (ch *Chain) L2BaseTokens(agentID isc.AgentID) uint64 {
	return ch.L2Assets(agentID).BaseTokens
}

func (ch *Chain) L2NFTs(agentID isc.AgentID) []iotago.NFTID {
	ret := make([]iotago.NFTID, 0)
	res, err := ch.CallView(accounts.Contract.Name, accounts.ViewAccountNFTs.Name, accounts.ParamAgentID, agentID)
	require.NoError(ch.Env.T, err)
	nftIDs := collections.NewArray16ReadOnly(res, accounts.ParamNFTIDs)
	nftLen := nftIDs.MustLen()
	for i := uint16(0); i < nftLen; i++ {
		nftID := iotago.NFTID{}
		copy(nftID[:], nftIDs.MustGetAt(i))
		ret = append(ret, nftID)
	}
	return ret
}

func (ch *Chain) L2NativeTokens(agentID isc.AgentID, nativeTokenID iotago.NativeTokenID) *big.Int {
	return ch.L2Assets(agentID).AmountNativeToken(nativeTokenID)
}

func (ch *Chain) L2CommonAccountAssets() *isc.Assets {
	return ch.L2Assets(ch.CommonAccount())
}

func (ch *Chain) L2CommonAccountBaseTokens() uint64 {
	return ch.L2Assets(ch.CommonAccount()).BaseTokens
}

func (ch *Chain) L2CommonAccountNativeTokens(nativeTokenID iotago.NativeTokenID) *big.Int {
	return ch.L2Assets(ch.CommonAccount()).AmountNativeToken(nativeTokenID)
}

// L2TotalAssets return total sum of ftokens contained in the on-chain accounts
func (ch *Chain) L2TotalAssets() *isc.Assets {
	return ch.parseAccountBalance(
		ch.CallView(accounts.Contract.Name, accounts.ViewTotalAssets.Name),
	)
}

// L2TotalBaseTokens return total sum of base tokens in L2 (all accounts)
func (ch *Chain) L2TotalBaseTokens() uint64 {
	return ch.L2TotalAssets().BaseTokens
}

func mustNativeTokenIDFromBytes(data []byte) iotago.NativeTokenID {
	if len(data) != iotago.NativeTokenIDLength {
		panic("len(data) != iotago.NativeTokenIDLength")
	}
	ret := iotago.NativeTokenID{}
	copy(ret[:], data)
	return ret
}

func (ch *Chain) GetOnChainTokenIDs() []iotago.NativeTokenID {
	res, err := ch.CallView(accounts.Contract.Name, accounts.ViewGetNativeTokenIDRegistry.Name)
	require.NoError(ch.Env.T, err)
	ret := make([]iotago.NativeTokenID, 0, len(res))
	for k := range res {
		ret = append(ret, mustNativeTokenIDFromBytes([]byte(k)))
	}
	return ret
}

func (ch *Chain) GetFoundryOutput(sn uint32) (*iotago.FoundryOutput, error) {
	res, err := ch.CallView(accounts.Contract.Name, accounts.ViewFoundryOutput.Name,
		accounts.ParamFoundrySN, sn,
	)
	if err != nil {
		return nil, err
	}
	outBin := res.MustGet(accounts.ParamFoundryOutputBin)
	out := &iotago.FoundryOutput{}
	_, err = out.Deserialize(outBin, serializer.DeSeriModeNoValidation, nil)
	require.NoError(ch.Env.T, err)
	return out, nil
}

func (ch *Chain) GetNativeTokenIDByFoundrySN(sn uint32) (iotago.NativeTokenID, error) {
	o, err := ch.GetFoundryOutput(sn)
	if err != nil {
		return iotago.NativeTokenID{}, err
	}
	return o.MustNativeTokenID(), nil
}

type foundryParams struct {
	ch   *Chain
	user *cryptolib.KeyPair
	sch  iotago.TokenScheme
}

// CreateFoundryGasBudgetBaseTokens always takes 100000 base tokens as gas budget and ftokens for the call
const (
	DestroyTokensGasBudgetBaseTokens       = 1 * isc.Million
	SendToL2AccountGasBudgetBaseTokens     = 1 * isc.Million
	DestroyFoundryGasBudgetBaseTokens      = 1 * isc.Million
	TransferAllowanceToGasBudgetBaseTokens = 1 * isc.Million
)

func (ch *Chain) NewFoundryParams(maxSupply interface{}) *foundryParams { //nolint:revive
	ret := &foundryParams{
		ch: ch,
		sch: &iotago.SimpleTokenScheme{
			MaximumSupply: util.ToBigInt(maxSupply),
			MeltedTokens:  big.NewInt(0),
			MintedTokens:  big.NewInt(0),
		},
	}
	return ret
}

func (fp *foundryParams) WithUser(user *cryptolib.KeyPair) *foundryParams {
	fp.user = user
	return fp
}

func (fp *foundryParams) WithTokenScheme(sch iotago.TokenScheme) *foundryParams {
	fp.sch = sch
	return fp
}

const (
	allowanceForFoundryStorageDeposit = 1 * isc.Million
	allowanceForModifySupply          = 1 * isc.Million
)

func (fp *foundryParams) CreateFoundry() (uint32, iotago.NativeTokenID, error) {
	par := dict.New()
	if fp.sch != nil {
		par.Set(accounts.ParamTokenScheme, codec.EncodeTokenScheme(fp.sch))
	}
	user := fp.ch.OriginatorPrivateKey
	if fp.user != nil {
		user = fp.user
	}
	req := NewCallParamsFromDict(accounts.Contract.Name, accounts.FuncFoundryCreateNew.Name, par).
		WithAllowance(isc.NewAssetsBaseTokens(allowanceForFoundryStorageDeposit))

	gas, _, err := fp.ch.EstimateGasOnLedger(req, user, true)
	if err != nil {
		return 0, iotago.NativeTokenID{}, err
	}
	req.WithGasBudget(gas)
	res, err := fp.ch.PostRequestSync(req, user)
	if err != nil {
		return 0, iotago.NativeTokenID{}, err
	}
	resDeco := kvdecoder.New(res)
	retSN := resDeco.MustGetUint32(accounts.ParamFoundrySN)
	nativeTokenID, err := fp.ch.GetNativeTokenIDByFoundrySN(retSN)
	return retSN, nativeTokenID, err
}

func toFoundrySN(foundry interface{}) uint32 {
	switch f := foundry.(type) {
	case uint32:
		return f
	case *iotago.NativeTokenID:
		return f.FoundrySerialNumber()
	case iotago.NativeTokenID:
		return f.FoundrySerialNumber()
	}
	panic(fmt.Sprintf("toFoundrySN: type %T not supported", foundry))
}

func (ch *Chain) DestroyFoundry(sn uint32, user *cryptolib.KeyPair) error {
	req := NewCallParams(accounts.Contract.Name, accounts.FuncFoundryDestroy.Name,
		accounts.ParamFoundrySN, sn).
		WithGasBudget(DestroyFoundryGasBudgetBaseTokens)
	_, err := ch.PostRequestSync(req, user)
	return err
}

func (ch *Chain) MintTokens(foundry, amount interface{}, user *cryptolib.KeyPair) error {
	req := NewCallParams(accounts.Contract.Name, accounts.FuncFoundryModifySupply.Name,
		accounts.ParamFoundrySN, toFoundrySN(foundry),
		accounts.ParamSupplyDeltaAbs, util.ToBigInt(amount),
	).
		WithAllowance(isc.NewAssetsBaseTokens(allowanceForModifySupply)) // enough allowance is needed for the storage deposit when token is minted first on the chain
	g, _, err := ch.EstimateGasOnLedger(req, user, true)
	if err != nil {
		return err
	}

	req.WithGasBudget(g)
	if user == nil {
		user = ch.OriginatorPrivateKey
	}
	_, err = ch.PostRequestSync(req, user)
	return err
}

// DestroyTokensOnL2 destroys tokens (identified by foundry SN) on user's on-chain account
func (ch *Chain) DestroyTokensOnL2(nativeTokenID iotago.NativeTokenID, amount interface{}, user *cryptolib.KeyPair) error {
	req := NewCallParams(accounts.Contract.Name, accounts.FuncFoundryModifySupply.Name,
		accounts.ParamFoundrySN, toFoundrySN(nativeTokenID),
		accounts.ParamSupplyDeltaAbs, util.ToBigInt(amount),
		accounts.ParamDestroyTokens, true,
	).WithAllowance(
		isc.NewAssets(0, iotago.NativeTokens{
			&iotago.NativeToken{
				ID:     nativeTokenID,
				Amount: util.ToBigInt(amount),
			},
		}),
	).WithGasBudget(DestroyTokensGasBudgetBaseTokens)

	if user == nil {
		user = ch.OriginatorPrivateKey
	}
	_, err := ch.PostRequestSync(req, user)
	return err
}

// DestroyTokensOnL1 sends tokens as ftokens and destroys in the same transaction
func (ch *Chain) DestroyTokensOnL1(nativeTokenID iotago.NativeTokenID, amount interface{}, user *cryptolib.KeyPair) error {
	req := NewCallParams(accounts.Contract.Name, accounts.FuncFoundryModifySupply.Name,
		accounts.ParamFoundrySN, toFoundrySN(nativeTokenID),
		accounts.ParamSupplyDeltaAbs, util.ToBigInt(amount),
		accounts.ParamDestroyTokens, true,
	).WithGasBudget(DestroyTokensGasBudgetBaseTokens).AddBaseTokens(1000)
	req.AddNativeTokens(nativeTokenID, amount)
	req.AddAllowanceNativeTokens(nativeTokenID, amount)
	if user == nil {
		user = ch.OriginatorPrivateKey
	}
	_, err := ch.PostRequestSync(req, user)
	return err
}

// DepositAssetsToL2 deposits ftokens on user's on-chain account
func (ch *Chain) DepositAssetsToL2(assets *isc.Assets, user *cryptolib.KeyPair) error {
	_, err := ch.PostRequestSync(
		NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name).
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
	wallet *cryptolib.KeyPair,
	nft ...*isc.NFT,
) error {
	callParams := NewCallParams(
		accounts.Contract.Name, accounts.FuncTransferAllowanceTo.Name,
		dict.Dict{
			accounts.ParamAgentID: codec.EncodeAgentID(targetAccount),
		}).
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
func (ch *Chain) DepositBaseTokensToL2(amount uint64, user *cryptolib.KeyPair) error {
	return ch.DepositAssetsToL2(isc.NewAssets(amount, nil), user)
}

func (ch *Chain) MustDepositBaseTokensToL2(amount uint64, user *cryptolib.KeyPair) {
	err := ch.DepositBaseTokensToL2(amount, user)
	require.NoError(ch.Env.T, err)
}

func (ch *Chain) DepositNFT(nft *isc.NFT, to isc.AgentID, owner *cryptolib.KeyPair) error {
	return ch.TransferAllowanceTo(
		isc.NewEmptyAssets().AddNFTs(nft.ID),
		to,
		owner,
		nft,
	)
}

func (ch *Chain) MustDepositNFT(nft *isc.NFT, to isc.AgentID, owner *cryptolib.KeyPair) {
	err := ch.DepositNFT(nft, to, owner)
	require.NoError(ch.Env.T, err)
}

// Withdraw sends assets from the L2 account to L1
func (ch *Chain) Withdraw(assets *isc.Assets, user *cryptolib.KeyPair) error {
	_, err := ch.PostRequestSync(
		NewCallParams(accounts.Contract.Name, accounts.FuncWithdraw.Name).
			AddAllowance(assets).
			AddAllowance(isc.NewAssetsBaseTokens(1*isc.Million)). // for storage deposit
			WithGasBudget(math.MaxUint64),
		user,
	)
	return err
}

// SendFromL1ToL2Account sends ftokens from L1 address to the target account on L2
// Sender pays the gas fee
func (ch *Chain) SendFromL1ToL2Account(totalBaseTokens uint64, toSend *isc.Assets, target isc.AgentID, user *cryptolib.KeyPair) error {
	require.False(ch.Env.T, toSend.IsEmpty())
	sumAssets := toSend.Clone().AddBaseTokens(totalBaseTokens)
	_, err := ch.PostRequestSync(
		NewCallParams(accounts.Contract.Name, accounts.FuncTransferAllowanceTo.Name,
			accounts.ParamAgentID, target,
		).
			AddFungibleTokens(sumAssets).
			AddAllowance(toSend).
			WithGasBudget(math.MaxUint64),
		user,
	)
	return err
}

func (ch *Chain) SendFromL1ToL2AccountBaseTokens(totalBaseTokens, baseTokensSend uint64, target isc.AgentID, user *cryptolib.KeyPair) error {
	return ch.SendFromL1ToL2Account(totalBaseTokens, isc.NewAssetsBaseTokens(baseTokensSend), target, user)
}

// SendFromL2ToL2Account moves ftokens on L2 from user's account to the target
func (ch *Chain) SendFromL2ToL2Account(transfer *isc.Assets, target isc.AgentID, user *cryptolib.KeyPair) error {
	req := NewCallParams(accounts.Contract.Name, accounts.FuncTransferAllowanceTo.Name,
		accounts.ParamAgentID, target,
	)

	req.AddBaseTokens(SendToL2AccountGasBudgetBaseTokens).
		AddAllowance(transfer).
		WithGasBudget(SendToL2AccountGasBudgetBaseTokens)
	_, err := ch.PostRequestSync(req, user)
	return err
}

func (ch *Chain) SendFromL2ToL2AccountBaseTokens(baseTokens uint64, target isc.AgentID, user *cryptolib.KeyPair) error {
	return ch.SendFromL2ToL2Account(isc.NewAssetsBaseTokens(baseTokens), target, user)
}

func (ch *Chain) SendFromL2ToL2AccountNativeTokens(id iotago.NativeTokenID, target isc.AgentID, amount interface{}, user *cryptolib.KeyPair) error {
	transfer := isc.NewEmptyAssets()
	transfer.AddNativeTokens(id, amount)
	return ch.SendFromL2ToL2Account(transfer, target, user)
}
