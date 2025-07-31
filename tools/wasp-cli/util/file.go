package util

import (
	"os"

	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
)

func ReadFile(fname string) []byte {
	b, err := os.ReadFile(fname)
	log.Check(err)
	return b
}
