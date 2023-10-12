package isc

import (
	"encoding/json"
	"strconv"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
)

type RequestJSON struct {
	Allowance     *AssetsJSON    `json:"allowance" swagger:"required"`
	CallTarget    CallTargetJSON `json:"callTarget" swagger:"required"`
	Assets        *AssetsJSON    `json:"fungibleTokens" swagger:"required"`
	GasBudget     string         `json:"gasBudget,string" swagger:"required,desc(The gas budget (uint64 as string))"`
	IsEVM         bool           `json:"isEVM" swagger:"required"`
	IsOffLedger   bool           `json:"isOffLedger" swagger:"required"`
	NFT           *NFTJSON       `json:"nft" swagger:"required"`
	Params        dict.JSONDict  `json:"params" swagger:"required"`
	RequestID     string         `json:"requestId" swagger:"required"`
	SenderAccount string         `json:"senderAccount" swagger:"required"`
	TargetAddress string         `json:"targetAddress" swagger:"required"`
}

func RequestToJSONObject(request Request) RequestJSON {
	gasBudget, isEVM := request.GasBudget()

	return RequestJSON{
		Allowance:     assetsToJSONObject(request.Allowance()),
		CallTarget:    callTargetToJSONObject(request.CallTarget()),
		Assets:        assetsToJSONObject(request.Assets()),
		GasBudget:     strconv.FormatUint(gasBudget, 10),
		IsEVM:         isEVM,
		IsOffLedger:   request.IsOffLedger(),
		NFT:           NFTToJSONObject(request.NFT()),
		Params:        request.Params().JSONDict(),
		RequestID:     request.ID().String(),
		SenderAccount: request.SenderAccount().String(),
		TargetAddress: request.TargetAddress().Bech32(parameters.L1().Protocol.Bech32HRP),
	}
}

func RequestToJSON(req Request) ([]byte, error) {
	return json.Marshal(RequestToJSONObject(req))
}

// ----------------------------------------------------------------------------

type AssetsJSON struct {
	BaseTokens   string             `json:"baseTokens" swagger:"required,desc(The base tokens (uint64 as string))"`
	NativeTokens []*NativeTokenJSON `json:"nativeTokens" swagger:"required"`
	NFTs         []string           `json:"nfts" swagger:"required"`
}

func assetsToJSONObject(assets *Assets) *AssetsJSON {
	if assets == nil {
		return nil
	}

	ret := &AssetsJSON{
		BaseTokens:   strconv.FormatUint(assets.BaseTokens, 10),
		NativeTokens: NativeTokensToJSONObject(assets.NativeTokens),
		NFTs:         make([]string, len(assets.NFTs)),
	}

	for k, v := range assets.NFTs {
		ret.NFTs[k] = v.ToHex()
	}
	return ret
}

// ----------------------------------------------------------------------------

type NFTJSON struct {
	ID       string `json:"id" swagger:"required"`
	Issuer   string `json:"issuer" swagger:"required"`
	Metadata string `json:"metadata" swagger:"required"`
	Owner    string `json:"owner" swagger:"required"`
}

func NFTToJSONObject(nft *NFT) *NFTJSON {
	if nft == nil {
		return nil
	}

	ownerString := ""
	if nft.Owner != nil {
		ownerString = nft.Owner.String()
	}

	return &NFTJSON{
		ID:       nft.ID.ToHex(),
		Issuer:   nft.Issuer.String(),
		Metadata: iotago.EncodeHex(nft.Metadata),
		Owner:    ownerString,
	}
}

// ----------------------------------------------------------------------------

type NativeTokenJSON struct {
	ID     string `json:"id" swagger:"required"`
	Amount string `json:"amount" swagger:"required"`
}

func NativeTokenToJSONObject(token *iotago.NativeToken) *NativeTokenJSON {
	return &NativeTokenJSON{
		ID:     token.ID.ToHex(),
		Amount: token.Amount.String(),
	}
}

func NativeTokensToJSONObject(tokens iotago.NativeTokens) []*NativeTokenJSON {
	nativeTokens := make([]*NativeTokenJSON, len(tokens))

	for k, v := range tokens {
		nativeTokens[k] = NativeTokenToJSONObject(v)
	}

	return nativeTokens
}

// ----------------------------------------------------------------------------

type CallTargetJSON struct {
	ContractHName string `json:"contractHName" swagger:"desc(The contract name as HName (Hex)),required"`
	FunctionHName string `json:"functionHName" swagger:"desc(The function name as HName (Hex)),required"`
}

func callTargetToJSONObject(target CallTarget) CallTargetJSON {
	return CallTargetJSON{
		ContractHName: target.Contract.String(),
		FunctionHName: target.EntryPoint.String(),
	}
}
