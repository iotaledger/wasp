package webapi_validation

import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"math"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/tools/stardust-migration/validation"
	"github.com/iotaledger/wasp/tools/stardust-migration/webapi-validation/base"
	"github.com/stretchr/testify/require"
)

//go:embed all_accounts_src.txt
var accountsFS embed.FS

type AccountValidation struct {
	client    base.AccountsClientWrapper
	addresses []ParsedAddress
}

func NewAccountValidation(validationContext base.ValidationContext) AccountValidation {

	file, err := accountsFS.Open("all_accounts_src.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	addresses, err := parseAddressesFromReader(file)
	if err != nil {
		panic(err)
	}

	return AccountValidation{
		addresses: addresses,
		client:    base.AccountsClientWrapper{ValidationContext: validationContext},
	}
}

func (a *AccountValidation) ValidateAccountBalances(stateIndex uint32) {
	a.parallel(func(address ParsedAddress) error {
		oldBalance, newBalance := a.client.AccountsGetAccountBalance(stateIndex, address.Address)
		oldBalance.BaseTokens = stardustBalanceToRebased(oldBalance.BaseTokens)

		// check if base tokens are equal
		require.EqualValues(base.T, oldBalance.BaseTokens, newBalance.BaseTokens)

		// Convert native tokens to map for comparison
		oldNativeTokens := make(map[string]string)
		newNativeTokens := make(map[string]string)

		for _, token := range oldBalance.NativeTokens {
			token.Amount = convertOverflowedBalance(token.Amount)
			oldNativeTokens[token.Id] = token.Amount
		}
		for _, token := range newBalance.NativeTokens {
			newNativeTokens[validation.CoinTypeToOldNTID(coin.MustTypeFromString(token.CoinType)).ToHex()] = token.Balance
		}

		// check if native tokens balances are equal
		require.EqualValues(base.T, oldNativeTokens, newNativeTokens)

		return nil
	})
}

func (a *AccountValidation) ValidateNFTs(stateIndex uint32) {
	a.parallel(func(address ParsedAddress) error {
		oldNFTs, newNFTs := a.client.AccountsGetAccountNFTIDs(stateIndex, address.Address)

		slices.Sort(oldNFTs.NftIds)
		slices.Sort(newNFTs.NftIds)

		// check if nft ids are equal
		require.EqualValues(base.T, oldNFTs.NftIds, newNFTs.NftIds)

		return nil
	})
}

func (a *AccountValidation) ValidateNonce(stateIndex uint32) {
	a.parallel(func(address ParsedAddress) error {
		oldNonce, newNonce := a.client.AccountsGetAccountNonce(stateIndex, address.Address)

		require.EqualValues(base.T, oldNonce, newNonce)

		return nil
	})
}

func (a *AccountValidation) ValidateTotalAssets(stateIndex uint32) {
	oldTotalAssets, newTotalAssets := a.client.AccountsGetTotalAssets(stateIndex)

	// Convert base tokens to rebased value
	oldTotalAssets.BaseTokens = stardustBalanceToRebased(oldTotalAssets.BaseTokens)

	// check if base tokens are equal
	require.EqualValues(base.T, oldTotalAssets.BaseTokens, newTotalAssets.BaseTokens)

	// Convert native tokens to map for comparison
	oldNativeTokens := make(map[string]string)
	newNativeTokens := make(map[string]string)

	for _, token := range oldTotalAssets.NativeTokens {
		token.Amount = convertOverflowedBalance(token.Amount)
		oldNativeTokens[token.Id] = token.Amount
	}

	for _, token := range newTotalAssets.NativeTokens {
		newNativeTokens[validation.CoinTypeToOldNTID(coin.MustTypeFromString(token.CoinType)).ToHex()] = token.Balance
	}

	// check if native tokens balances are equal
	require.EqualValues(base.T, oldNativeTokens, newNativeTokens)
}

func convertOverflowedBalance(balance string) string {
	_, err := strconv.ParseUint(balance, 10, 64)
	if err != nil {
		return strconv.FormatUint(math.MaxUint64, 10)
	}

	return balance
}

type ParsedAddress struct {
	Address string
	Type    isc.AgentIDKind
}

func (a *AccountValidation) parallel(f func(address ParsedAddress) error) {
	semaphore := make(chan struct{}, 100)
	progressChan := make(chan struct{}, 100)
	var processedCount atomic.Int32

	go func() {
		for range progressChan {
			fmt.Printf("processed %d addresses", processedCount.Add(1))
			fmt.Printf("\r")
		}
		fmt.Println()
	}()

	wg := sync.WaitGroup{}
	wg.Add(len(a.addresses))

	for _, address := range a.addresses {
		semaphore <- struct{}{}
		go func(addr ParsedAddress) {
			defer wg.Done()
			defer func() {
				<-semaphore
				progressChan <- struct{}{}
			}()

			err := f(address)
			if err != nil {
				panic(err)
			}

		}(address)
	}

	wg.Wait()
	close(progressChan)
}

func parseAddressesFromReader(r io.Reader) ([]ParsedAddress, error) {
	var result []ParsedAddress

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		addrStr := scanner.Text()
		addrStr = strings.TrimSpace(addrStr)
		if addrStr == "" {
			continue
		}

		switch {
		case strings.HasPrefix(addrStr, "EthereumAddressAgentID(") && strings.HasSuffix(addrStr, ")"):
			addr := strings.TrimPrefix(addrStr, "EthereumAddressAgentID(")
			addr = strings.TrimSuffix(addr, ")")
			result = append(result, ParsedAddress{
				Address: addr,
				Type:    isc.AgentIDKindEthereumAddress,
			})
			continue

		case strings.HasPrefix(addrStr, "AddressAgentID(") && strings.HasSuffix(addrStr, ")"):
			addr := strings.TrimPrefix(addrStr, "AddressAgentID(")
			addr = strings.TrimSuffix(addr, ")")
			result = append(result, ParsedAddress{
				Address: addr,
				Type:    isc.AgentIDKindAddress,
			})
			continue

		case strings.HasPrefix(addrStr, "ContractAgentID(") && strings.HasSuffix(addrStr, ")"):
			addr := strings.TrimPrefix(addrStr, "ContractAgentID(")
			addr = strings.TrimSuffix(addr, ")")
			result = append(result, ParsedAddress{
				Address: addr,
				Type:    isc.AgentIDKindContract,
			})
			continue

		case addrStr == "0x0000000000000000000000000000000000000000":
			result = append(result, ParsedAddress{
				Address: addrStr,
				Type:    isc.AgentIDKindEthereumAddress,
			})
			continue

		default:
			return nil, fmt.Errorf("invalid address format: %s", addrStr)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return result, nil
}
