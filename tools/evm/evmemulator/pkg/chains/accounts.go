package chains

import (
	"crypto/ecdsa"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iotaledger/wasp/v2/packages/solo"
	"github.com/iotaledger/wasp/v2/tools/evm/evmemulator/pkg/cli"
	"github.com/iotaledger/wasp/v2/tools/evm/evmemulator/pkg/log"
)

func CreateAccounts(chain *solo.Chain) (accounts []*ecdsa.PrivateKey) {
	log.Printf("creating accounts with funds...\n")
	header := []string{"private key", "address"}
	var rows [][]string

	if cli.IsHive {
		// Hive logic: create a single account using a random depositor
		pk, addr := chain.EthereumAccountByIndexWithL2FundsRandDepositor(0)
		accounts = append(accounts, pk)
		rows = append(rows, []string{hex.EncodeToString(crypto.FromECDSA(pk)), addr.String()})
	} else {
		// Normal logic: create the default set of ethereum accounts with L2 funds
		for i := 0; i < len(solo.EthereumAccounts); i++ {
			pk, addr := chain.EthereumAccountByIndexWithL2Funds(i)
			accounts = append(accounts, pk)
			rows = append(rows, []string{hex.EncodeToString(crypto.FromECDSA(pk)), addr.String()})
		}
	}

	log.PrintTable(header, rows)
	return accounts
}
