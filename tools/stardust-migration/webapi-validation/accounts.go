package webapi_validation

import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/iotaledger/wasp/packages/isc"
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

func (a *AccountValidation) ValidateBaseTokenBalances(stateIndex uint32) {
	addresses := make(chan ParsedAddress)
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

	go func() {
		for _, address := range a.addresses {
			addresses <- address
		}
		close(addresses)
	}()

	wg := sync.WaitGroup{}
	wg.Add(len(a.addresses))

	for address := range addresses {
		semaphore <- struct{}{}
		go func(addr ParsedAddress) {
			defer wg.Done()
			defer func() {
				<-semaphore
				progressChan <- struct{}{}
			}()

			oldBalance, newBalance := a.client.AccountsGetAccountBalance(stateIndex, addr.Address)
			oldBalance.BaseTokens = stardustBalanceToRebased(oldBalance.BaseTokens)
			require.EqualValues(base.T, oldBalance.BaseTokens, newBalance.BaseTokens)
		}(address)
	}

	wg.Wait()
	close(progressChan)
}

func (a *AccountValidation) ValidateNativeTokenBalances(ctx base.ValidationContext) {

}

func (a *AccountValidation) ValidateNFTs(ctx base.ValidationContext) {

}

type ParsedAddress struct {
	Address string
	Type    isc.AgentIDKind
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
