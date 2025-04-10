package webapi_validation

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	stardust_apiclient "github.com/nnikolash/wasp-types-exported/clients/apiclient"

	iotago "github.com/iotaledger/iota.go/v3"
	rebased_apiclient "github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/tools/stardust-migration/webapi-validation/base"
)

type CoreBlockLogValidation struct {
	base.ValidationContext
	client base.BlockLogClientWrapper
}

func NewCoreBlockLogValidation(validationContext base.ValidationContext) CoreBlockLogValidation {
	return CoreBlockLogValidation{
		ValidationContext: validationContext,
		client:            base.BlockLogClientWrapper{ValidationContext: validationContext},
	}
}

func (c *CoreBlockLogValidation) Validate(stateIndex uint32) {
	c.validateBlockInfo(stateIndex)
	c.validateRequestIDsInBlock(stateIndex)
	c.validateRequestsInBlock(stateIndex)
}

func trimStardustRequestIDForRebased(requestID string) string {
	decodedRequestID := hexutil.MustDecode(requestID)
	require.Len(base.T, decodedRequestID, 34) // 34 == requestID(32)+Digest(2)
	trimmedRequestID := hexutil.Encode(decodedRequestID[:32])
	return trimmedRequestID
}

func trimStardustRequestIDsForRebased(requestIDs []string) []string {
	stardustRequestIDs := make([]string, len(requestIDs))
	for i, requestID := range requestIDs {
		stardustRequestIDs[i] = trimStardustRequestIDForRebased(requestID)
	}

	return stardustRequestIDs
}

func stardustBurnRecordToRebased(b []stardust_apiclient.BurnRecord) []rebased_apiclient.BurnRecord {
	rebasedBurnRecords := make([]rebased_apiclient.BurnRecord, len(b))
	for i, record := range b {
		rebasedBurnRecords[i] = rebased_apiclient.BurnRecord{
			Code:      record.Code,
			GasBurned: record.GasBurned,
		}
	}
	return rebasedBurnRecords
}

func stardustParamsToRebased(params []stardust_apiclient.Item) []string {
	rebasedParams := make([]string, len(params))

	for i, param := range params {
		rebasedParams[i] = param.Value
	}

	return rebasedParams
}

func stardustWebHNameToRebased(name string) string {
	return fmt.Sprintf("0x%s", name)
}

func bech32ToHex(address string) string {

	if strings.Contains(address, "@") {
		// This is either an EVM address or a contract call
		// EVM@ChainID
		// Contract@ChainID

		splitString := strings.Split(address, "@")
		return splitString[0]
	}

	_, bytes, err := iotago.ParseBech32(address)
	require.NoError(base.T, err, fmt.Sprintf("Passed address: %s", address))

	switch bytes.(type) {
	case *iotago.Ed25519Address:
		return hexutil.Encode(bytes.(*iotago.Ed25519Address)[:])
	case *iotago.NFTAddress:
		return hexutil.Encode(bytes.(*iotago.NFTAddress)[:])
	case *iotago.AliasAddress:
		return hexutil.Encode(bytes.(*iotago.AliasAddress)[:])
	default:
		panic("failed to convert Bech32 address!")
	}
}

func stardustBalanceToRebased(balance string) string {
	number, err := strconv.ParseUint(balance, 10, 64)
	require.NoError(base.T, err)
	return strconv.FormatUint(number*1000, 10)
}

func (c *CoreBlockLogValidation) validateRequestIDsInBlock(stateIndex uint32) {
	sRes, rRes := c.client.BlocklogGetRequestIDsForBlock(stateIndex)
	// Stardust supported multiple requests per output. This required us to add a "digest" at the end of each request which was formed round about like so: <OutputID;RequestIndex;>
	// In Rebased, RequestIDs are unique per request, letting us drop the digest.
	// To make the comparison simpler, drop the last two bytes and write them into a new array which then can be compared.

	stardustRequestIDs := trimStardustRequestIDsForRebased(sRes.RequestIds)

	require.EqualValues(base.T, stardustRequestIDs, rRes.RequestIds)
}

func (c *CoreBlockLogValidation) validateRequestsInBlock(stateIndex uint32) {
	sRes, rRes := c.client.BlocklogGetRequestReceiptsOfBlock(stateIndex)

	require.Len(base.T, sRes, len(rRes))
	// With being sure, that the arrays have an equal length, validate the receipt.

	stardustReceiptMap := map[string]stardust_apiclient.ReceiptResponse{}
	lo.ForEach(sRes, func(item stardust_apiclient.ReceiptResponse, index int) {
		stardustReceiptMap[item.Request.RequestId] = item
	})

	rebasedReceiptMap := map[string]rebased_apiclient.ReceiptResponse{}
	lo.ForEach(rRes, func(item rebased_apiclient.ReceiptResponse, index int) {
		rebasedReceiptMap[item.Request.RequestId] = item
	})

	for requestID, stardustReceipt := range stardustReceiptMap {
		rebasedReceipt := rebasedReceiptMap[trimStardustRequestIDForRebased(requestID)]

		require.Equal(base.T, rebasedReceipt.BlockIndex, stardustReceipt.BlockIndex)
		require.Equal(base.T, rebasedReceipt.ErrorMessage, stardustReceipt.ErrorMessage)
		require.Equal(base.T, rebasedReceipt.GasBurned, stardustReceipt.GasBurned)
		require.Equal(base.T, rebasedReceipt.GasBudget, stardustReceipt.GasBudget)
		require.Equal(base.T, rebasedReceipt.GasFeeCharged, stardustBalanceToRebased(stardustReceipt.GasFeeCharged))

		// If one is nil, the other must be too. Enforce the check by checking either, then both.
		if rebasedReceipt.RawError != nil || stardustReceipt.RawError != nil {
			require.NotNil(base.T, rebasedReceipt.RawError)
			require.NotNil(base.T, stardustReceipt.RawError)

			require.EqualValues(base.T, rebasedReceipt.RawError.Code, stardustReceipt.RawError.Code)
			require.EqualValues(base.T, rebasedReceipt.RawError.Params, stardustReceipt.RawError.Params)
		}

		require.EqualValues(base.T, rebasedReceipt.RawError, stardustReceipt.RawError)
		require.Equal(base.T, rebasedReceipt.RequestIndex, stardustReceipt.RequestIndex)
		require.EqualValues(base.T, rebasedReceipt.GasBurnLog, stardustBurnRecordToRebased(stardustReceipt.GasBurnLog))
		require.Equal(base.T, rebasedReceipt.StorageDepositCharged, stardustReceipt.StorageDepositCharged)

		// Excluding this for now, as the migrator converts known function args into the new Rebased format and a direct comparison won't work
		// Try using `migrateContractCall` based on the stringified params maybe..
		// require.EqualValues(base.T, rebasedReceipt.Request.Params, stardustParamsToRebased(stardustReceipt.Request.Params.Items))

		require.EqualValues(base.T, rebasedReceipt.Request.GasBudget, stardustReceipt.Request.GasBudget)
		require.EqualValues(base.T, rebasedReceipt.Request.RequestId, trimStardustRequestIDForRebased(stardustReceipt.Request.RequestId))
		require.EqualValues(base.T, rebasedReceipt.Request.CallTarget.ContractHName, stardustWebHNameToRebased(stardustReceipt.Request.CallTarget.ContractHName))
		require.EqualValues(base.T, rebasedReceipt.Request.CallTarget.FunctionHName, stardustWebHNameToRebased(stardustReceipt.Request.CallTarget.FunctionHName))
		require.EqualValues(base.T, rebasedReceipt.Request.IsEVM, stardustReceipt.Request.IsEVM)
		require.EqualValues(base.T, rebasedReceipt.Request.SenderAccount, bech32ToHex(stardustReceipt.Request.SenderAccount))
		require.EqualValues(base.T, rebasedReceipt.Request.IsOffLedger, stardustReceipt.Request.IsOffLedger)
		// Omiting stardustReceipt.Request.TargetAddress as it's always the ChainID and therefore removed on Rebased
	}
}

func (c *CoreBlockLogValidation) validateBlockInfo(stateIndex uint32) {
	sRes, rRes := c.client.BlocklogGetBlockInfo(stateIndex)

	require.Equal(base.T, sRes.BlockIndex, rRes.BlockIndex)
	require.Equal(base.T, sRes.GasBurned, rRes.GasBurned)
	require.Equal(base.T, stardustBalanceToRebased(sRes.GasFeeCharged), rRes.GasFeeCharged, fmt.Sprintf("Rebased Balance: '%s', Stardust Balance: '%s'", rRes.GasFeeCharged, sRes.GasFeeCharged))
	require.Equal(base.T, sRes.NumOffLedgerRequests, rRes.NumOffLedgerRequests)
	require.Equal(base.T, sRes.NumSuccessfulRequests, rRes.NumSuccessfulRequests)
	require.Equal(base.T, sRes.TotalRequests, rRes.TotalRequests)
	require.Equal(base.T, sRes.Timestamp, rRes.Timestamp)
	// PreviousAliasOutput / PreviousAnchor omitted, as the Blocks have been rebuilt, and contain different IDs.
}
