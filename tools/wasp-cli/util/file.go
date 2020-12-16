package util

import (
	"io/ioutil"

	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func ReadFile(fname string) []byte {
	b, err := ioutil.ReadFile(fname)
	log.Check(err)
	return b
}
