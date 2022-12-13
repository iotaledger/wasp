package corecontracts

import (
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common/hexutil"

	iotago "github.com/iotaledger/iota.go/v3"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"
	"github.com/labstack/echo/v4"
)

type NativeToken struct {
	ID     string
	Amount string
}

func parseNativeToken(token *iotago.NativeToken) *NativeToken {
	return &NativeToken{
		ID:     token.ID.ToHex(),
		Amount: token.Amount.String(),
	}
}

func parseNativeTokens(tokens iotago.NativeTokens) []*NativeToken {
	nativeTokens := make([]*NativeToken, len(tokens))

	for k, v := range tokens {
		nativeTokens[k] = parseNativeToken(v)
	}

	return nativeTokens
}

type AccountListResponse struct {
	Accounts []string
}

func (c *Controller) getAccounts(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	accounts, err := c.accounts.GetAccounts(chainID)

	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	accountsResponse := &AccountListResponse{
		Accounts: make([]string, len(accounts)),
	}

	for k, v := range accounts {
		accountsResponse.Accounts[k] = v.String()
	}

	return e.JSON(http.StatusOK, accountsResponse)
}

type AssetsResponse struct {
	BaseTokens uint64
	Tokens     []*NativeToken
}

func (c *Controller) getTotalAssets(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	assets, err := c.accounts.GetTotalAssets(chainID)

	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	assetsResponse := &AssetsResponse{
		BaseTokens: assets.BaseTokens,
		Tokens:     parseNativeTokens(assets.Tokens),
	}

	return e.JSON(http.StatusOK, assetsResponse)
}

func (c *Controller) getAccountBalance(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	agentID, err := isc.NewAgentIDFromString(e.Param("agentID"))
	if err != nil {
		return apierrors.InvalidPropertyError("agentID", err)
	}

	assets, err := c.accounts.GetAccountBalance(chainID, agentID)

	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	assetsResponse := &AssetsResponse{
		BaseTokens: assets.BaseTokens,
		Tokens:     parseNativeTokens(assets.Tokens),
	}

	return e.JSON(http.StatusOK, assetsResponse)
}

type AccountNFTsResponse struct {
	NFTIDs []string
}

func (c *Controller) getAccountNFTs(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	agentID, err := isc.NewAgentIDFromString(e.Param("agentID"))
	if err != nil {
		return apierrors.InvalidPropertyError("agentID", err)
	}

	nfts, err := c.accounts.GetAccountNFTs(chainID, agentID)

	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	nftsResponse := &AccountNFTsResponse{
		NFTIDs: make([]string, len(nfts)),
	}

	for k, v := range nfts {
		nftsResponse.NFTIDs[k] = v.ToHex()
	}

	return e.JSON(http.StatusOK, nftsResponse)
}

type AccountNonceResponse struct {
	Nonce uint64
}

func (c *Controller) getAccountNonce(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	agentID, err := isc.NewAgentIDFromString(e.Param("agentID"))
	if err != nil {
		return apierrors.InvalidPropertyError("agentID", err)
	}

	nonce, err := c.accounts.GetAccountNonce(chainID, agentID)

	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	nonceResponse := &AccountNonceResponse{
		Nonce: nonce,
	}

	return e.JSON(http.StatusOK, nonceResponse)
}

type NFTDataResponse struct {
	ID       string
	Issuer   string
	Metadata string
	Owner    string
}

func (c *Controller) getNFTData(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	nftIDBytes, err := hexutil.Decode(e.Param("nftID"))
	if err != nil {
		return apierrors.InvalidPropertyError("nftID", err)
	}

	if len(nftIDBytes) != iotago.NFTIDLength {
		return apierrors.InvalidPropertyError("nftID", err)
	}

	var nftID iotago.NFTID
	copy(nftID[:], nftIDBytes)

	if err != nil {
		return apierrors.InvalidPropertyError("agentID", err)
	}

	nftData, err := c.accounts.GetNFTData(chainID, nftID)

	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	nftDataResponse := &NFTDataResponse{
		ID:       nftData.ID.ToHex(),
		Issuer:   nftData.Issuer.String(),
		Metadata: hexutil.Encode(nftData.Metadata),
	}

	if nftData.Owner != nil {
		nftDataResponse.Owner = nftData.Owner.String()
	}

	return e.JSON(http.StatusOK, nftDataResponse)
}

type NativeTokenIDRegistryResponse struct {
	NativeTokenRegistryIDs []string
}

func (c *Controller) getNativeTokenIDRegistry(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	registries, err := c.accounts.GetNativeTokenIDRegistry(chainID)

	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	nativeTokenIDRegistryResponse := &NativeTokenIDRegistryResponse{
		NativeTokenRegistryIDs: make([]string, len(registries)),
	}

	for k, v := range registries {
		nativeTokenIDRegistryResponse.NativeTokenRegistryIDs[k] = v.String()
	}

	return e.JSON(http.StatusOK, nativeTokenIDRegistryResponse)
}

type FoundryOutputResponse struct {
	FoundryID string
	Token     AssetsResponse
}

func (c *Controller) getFoundryOutput(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	serialNumber, err := strconv.ParseUint(e.Param("serialNumber"), 10, 64)
	if err != nil {
		return apierrors.InvalidPropertyError("serialNumber", err)
	}

	foundryOutput, err := c.accounts.GetFoundryOutput(chainID, uint32(serialNumber))

	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	foundryOutputID, err := foundryOutput.ID()

	if err != nil {
		return apierrors.InvalidPropertyError("FoundryOutput.ID", err)
	}

	foundryOutputResponse := &FoundryOutputResponse{
		FoundryID: foundryOutputID.ToHex(),
		Token: AssetsResponse{
			BaseTokens: foundryOutput.Amount,
			Tokens:     parseNativeTokens(foundryOutput.NativeTokens),
		},
	}

	return e.JSON(http.StatusOK, foundryOutputResponse)
}
