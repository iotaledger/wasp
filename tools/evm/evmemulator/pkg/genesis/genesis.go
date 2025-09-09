package genesis

import (
	"encoding/json"
	"io"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/core"
	"github.com/iotaledger/wasp/v2/tools/evm/evmemulator/pkg/log"
)

func InitGenesis(genesisPath string) (*core.Genesis, error) {
	file, err := os.Open(genesisPath)
	if err != nil {
		log.Fatalf("Failed to read genesis file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("failed to read genesis file: %v", err)
	}

	genesis := new(core.Genesis)
	if err := json.Unmarshal(data, genesis); err != nil {
		log.Fatalf("invalid genesis file: %v", err)
	}

	return genesis, nil
}

const MaxPreFundAmount = 10_000

func RegulateGenesisAccountBalance(genesis *core.Genesis) *core.Genesis {
    max := big.NewInt(MaxPreFundAmount)
    for addr, acc := range genesis.Alloc {
        switch {
        case acc.Balance == nil:
            acc.Balance = big.NewInt(0)
        case acc.Balance.Cmp(max) > 0:
            acc.Balance = new(big.Int).Set(max)
        default:
            acc.Balance = new(big.Int).Set(acc.Balance)
        }
        genesis.Alloc[addr] = acc
    }
    return genesis
}
