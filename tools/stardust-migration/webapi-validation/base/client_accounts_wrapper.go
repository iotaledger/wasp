package base

import (
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	stardust_client "github.com/nnikolash/wasp-types-exported/clients/apiclient"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	rebased_client "github.com/iotaledger/wasp/clients/apiclient"
)

type AccountsClientWrapper struct {
	ValidationContext
}

// AccountsGetAccountBalance wraps both API calls for getting account balance
func (c *AccountsClientWrapper) AccountsGetAccountBalance(stateIndex uint32, agentID string) (*stardust_client.AssetsResponse, *rebased_client.AssetsResponse) {
	oldAgentIDStr := addHexPrefix(c.oldAgentIDFromHex(agentID).String())
	sRes, rawResponse, err := c.SClient.CorecontractsApi.AccountsGetAccountBalance(c.Ctx, MainnetChainID, oldAgentIDStr).Block(Uint32ToString(stateIndex)).Execute()
	require.NoError(T, err, "response body: %s, agentId: %s", string(lo.Must1(io.ReadAll(rawResponse.Body))), oldAgentIDStr)
	// baseTokens := lo.Must(strconv.Atoi(sRes.BaseTokens))
	// if /*baseTokens > 0 || */ len(sRes.NativeTokens) > 0 {
	// 	fmt.Printf("stardust: agent %s, base tokens: %s, native tokens: %s\n", agentID, sRes.BaseTokens, sRes.NativeTokens)
	// }

	rRes, rawResponse, err := c.RClient.CorecontractsAPI.AccountsGetAccountBalance(c.Ctx, addHexPrefix(agentID)).Block(Uint32ToString(stateIndex)).Execute()
	require.NoError(T, err, "response body: %s, agentId: %s", string(lo.Must1(io.ReadAll(rawResponse.Body))), agentID)
	// baseTokens = lo.Must(strconv.Atoi(rRes.BaseTokens))
	// if /*baseTokens > 0 || */ len(rRes.NativeTokens) > 0 {
	// 	fmt.Printf("rebased: agent %s, base tokens: %s, native tokens: %s\n", agentID, rRes.BaseTokens, rRes.NativeTokens)
	// }

	return sRes, rRes
}

// AccountsGetAccountNFTIDs wraps both API calls for getting account NFT IDs
func (c *AccountsClientWrapper) AccountsGetAccountNFTIDs(stateIndex uint32, agentID string) (*stardust_client.AccountNFTsResponse, *rebased_client.AccountNFTsResponse) {
	oldAgentIDStr := addHexPrefix(c.oldAgentIDFromHex(agentID).String())
	sRes, rawResponse, err := c.SClient.CorecontractsApi.AccountsGetAccountNFTIDs(c.Ctx, MainnetChainID, oldAgentIDStr).Block(Uint32ToString(stateIndex)).Execute()
	require.NoError(T, err, "response body: %s, agentId: %s", string(lo.Must1(io.ReadAll(rawResponse.Body))), oldAgentIDStr)

	rRes, rawResponse, err := c.RClient.CorecontractsAPI.AccountsGetAccountNFTIDs(c.Ctx, addHexPrefix(agentID)).Block(Uint32ToString(stateIndex)).Execute()
	require.NoError(T, err, "response body: %s, agentId: %s", string(lo.Must1(io.ReadAll(rawResponse.Body))), agentID)

	return sRes, rRes
}

// AccountsGetAccountNonce wraps both API calls for getting account nonce
func (c *AccountsClientWrapper) AccountsGetAccountNonce(stateIndex uint32, agentID string) (*stardust_client.AccountNonceResponse, *rebased_client.AccountNonceResponse) {
	if isEVMAddress(agentID) {
		// TODO: handle evm address
		return nil, nil
	}

	oldAgentIDStr := addHexPrefix(c.oldAgentIDFromHex(agentID).String())
	sRes, rawResponse, err := c.SClient.CorecontractsApi.AccountsGetAccountNonce(c.Ctx, MainnetChainID, oldAgentIDStr).Block(Uint32ToString(stateIndex)).Execute()
	require.NoError(T, err, "response body: %s, agentId: %s", string(lo.Must1(io.ReadAll(rawResponse.Body))), oldAgentIDStr)

	rRes, rawResponse, err := c.RClient.CorecontractsAPI.AccountsGetAccountNonce(c.Ctx, addHexPrefix(agentID)).Block(Uint32ToString(stateIndex)).Execute()
	require.NoError(T, err, "response body: %s, agentId: %s", string(lo.Must1(io.ReadAll(rawResponse.Body))), agentID)

	return sRes, rRes
}

// AccountsGetTotalAssets wraps both API calls for getting total assets
func (c *AccountsClientWrapper) AccountsGetTotalAssets(stateIndex uint32) (*stardust_client.AssetsResponse, *rebased_client.AssetsResponse) {
	sRes, rawResponse, err := c.SClient.CorecontractsApi.AccountsGetTotalAssets(c.Ctx, MainnetChainID).Block(Uint32ToString(stateIndex)).Execute()
	require.NoError(T, err, "response body: %s", string(lo.Must1(io.ReadAll(rawResponse.Body))))

	rRes, rawResponse, err := c.RClient.CorecontractsAPI.AccountsGetTotalAssets(c.Ctx).Block(Uint32ToString(stateIndex)).Execute()
	require.NoError(T, err, "response body: %s", string(lo.Must1(io.ReadAll(rawResponse.Body))))

	return sRes, rRes
}

func isEVMAddress(s string) bool {
	return len(s) == 42 && s[:2] == "0x"
}

func addHexPrefix(s string) string {
	// skip if contains 0x or is not a contract hname
	if len(s) > 8 || s[:2] == "0x" {
		return s
	}

	return "0x" + s
}

func (c *AccountsClientWrapper) oldAgentIDFromHex(agentID string) old_isc.AgentID {
	switch {
	case len(agentID) == 42:
		// evm address
		return old_isc.NewEthereumAddressAgentID(ChainID, common.HexToAddress(agentID))
	case len(agentID) == 66:
		// iota address
		s := strings.Replace(agentID, "0x", "", 1)
		s = "0100" + s
		return lo.Must(old_isc.AgentIDFromBytes(lo.Must(hex.DecodeString(s))))
	case len(agentID) == 8:
		// contract hname
		hname := lo.Must(old_isc.HnameFromString(agentID))
		return old_isc.NewContractAgentID(ChainID, hname)
	default:
		panic(fmt.Sprintf("Unknown agent ID: %s", agentID))
	}
}
