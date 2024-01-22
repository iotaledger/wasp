package solo

import (
	"fmt"
	"math/big"

	"github.com/samber/lo"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

// GrantDeployPermission gives permission to the specified agentID to deploy SCs into the chain
func (ch *Chain) GrantDeployPermission(keyPair *cryptolib.KeyPair, deployerAgentID isc.AgentID) error {
	if keyPair == nil {
		keyPair = ch.OriginatorPrivateKey
	}

	req := NewCallParams(root.Contract.Name, root.FuncGrantDeployPermission.Name, root.ParamDeployer, deployerAgentID).AddBaseTokens(1)
	_, err := ch.PostRequestSync(req, keyPair)
	return err
}

// RevokeDeployPermission removes permission of the specified agentID to deploy SCs into the chain
func (ch *Chain) RevokeDeployPermission(keyPair *cryptolib.KeyPair, deployerAgentID isc.AgentID) error {
	if keyPair == nil {
		keyPair = ch.OriginatorPrivateKey
	}

	req := NewCallParams(root.Contract.Name, root.FuncRevokeDeployPermission.Name, root.ParamDeployer, deployerAgentID).AddBaseTokens(1)
	_, err := ch.PostRequestSync(req, keyPair)
	return err
}

func (ch *Chain) ContractAgentID(name string) isc.AgentID {
	return isc.NewContractAgentID(ch.ChainID, isc.Hn(name))
}

// Warning: if the same `req` is passed in different occasions, the resulting request will have different IDs (because the ledger state is different)
func ISCRequestFromCallParams(ch *Chain, req *CallParams, keyPair *cryptolib.KeyPair) (isc.Request, error) {
	tx, _, err := ch.RequestFromParamsToLedger(req, keyPair)
	if err != nil {
		return nil, err
	}
	requestsFromSignedTx, err := isc.RequestsInTransaction(tx)
	if err != nil {
		return nil, err
	}
	return requestsFromSignedTx[ch.ChainID][0], nil
}

// only used in internal tests and solo
func CheckLedger(v isc.SchemaVersion, state kv.KVStoreReader, checkpoint string) {
	t := accounts.GetTotalL2FungibleTokens(v, state)
	c := calcL2TotalFungibleTokens(v, state)
	if !t.Equals(c) {
		panic(fmt.Sprintf("inconsistent on-chain account ledger @ checkpoint '%s'\n total assets: %s\ncalc total: %s\n",
			checkpoint, t, c))
	}

	totalAccNFTs := accounts.GetTotalL2NFTs(state)
	if len(lo.FindDuplicates(totalAccNFTs)) != 0 {
		panic(fmt.Sprintf("inconsistent on-chain account ledger @ checkpoint '%s'\n duplicate NFTs\n", checkpoint))
	}
	calculatedNFTs := calcL2TotalNFTs(state)
	if len(lo.FindDuplicates(calculatedNFTs)) != 0 {
		panic(fmt.Sprintf("inconsistent on-chain account ledger @ checkpoint '%s'\n duplicate NFTs\n", checkpoint))
	}
	left, right := lo.Difference(calculatedNFTs, totalAccNFTs)
	if len(left)+len(right) != 0 {
		panic(fmt.Sprintf("inconsistent on-chain account ledger @ checkpoint '%s'\n NFTs don't match\n", checkpoint))
	}
}

func calcL2TotalFungibleTokens(v isc.SchemaVersion, state kv.KVStoreReader) *isc.Assets {
	ret := isc.NewEmptyAssets()
	totalBaseTokens := big.NewInt(0)

	accounts.AllAccountsMapR(state).IterateKeys(func(accountKey []byte) bool {
		// add all native tokens owned by each account
		accounts.NativeTokensMapR(state, kv.Key(accountKey)).Iterate(func(idBytes []byte, val []byte) bool {
			ret.AddNativeTokens(
				isc.MustNativeTokenIDFromBytes(idBytes),
				new(big.Int).SetBytes(val),
			)
			return true
		})
		// use the full decimals for each account, so no dust balance is lost in the calculation
		baseTokensFullDecimals := accounts.GetBaseTokensFullDecimals(v)(state, kv.Key(accountKey))
		totalBaseTokens = new(big.Int).Add(totalBaseTokens, baseTokensFullDecimals)
		return true
	})

	// convert from 18 decimals, remainder must be 0
	ret.BaseTokens = util.MustEthereumDecimalsToBaseTokenDecimalsExact(totalBaseTokens, parameters.L1().BaseToken.Decimals)
	return ret
}

func calcL2TotalNFTs(state kv.KVStoreReader) []iotago.NFTID {
	var ret []iotago.NFTID
	accounts.AllAccountsMapR(state).IterateKeys(func(key []byte) bool {
		agentID, err := isc.AgentIDFromBytes(key) // obs: this can only be done because the key saves the entire bytes of agentID, unlike the BaseTokens/NativeTokens accounting
		if err != nil {
			panic(fmt.Errorf("calcL2TotalNFTs: %w", err))
		}
		ret = append(ret, accounts.GetAccountNFTs(state, agentID)...)
		return true
	})
	return ret
}
