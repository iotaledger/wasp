package genesis

import (
	"encoding/json"
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

	genesis := new(core.Genesis)
	if err := json.NewDecoder(file).Decode(genesis); err != nil {
		log.Fatalf("invalid genesis file: %v", err)
	}

	return genesis, nil
}
