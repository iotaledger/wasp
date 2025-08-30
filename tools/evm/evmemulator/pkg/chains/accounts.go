package chains

import (
	"crypto/ecdsa"
	"encoding/hex"
	"math/rand"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iotaledger/wasp/v2/packages/solo"
	"github.com/iotaledger/wasp/v2/tools/evm/evmemulator/pkg/log"
)

func CreateAccounts(chain *solo.Chain) (accounts []*ecdsa.PrivateKey) {
	log.Printf("creating accounts with funds...\n")
	header := []string{"private key", "address"}
	var rows [][]string
	// FIXME we cant afford prefund that much address
	pk, addr := chain.EthereumAccountByIndexWithL2FundsRandDepositor(rand.Int())
	accounts = append(accounts, pk)
	rows = append(rows, []string{hex.EncodeToString(crypto.FromECDSA(pk)), addr.String()})
	log.PrintTable(header, rows)
	return accounts
}
