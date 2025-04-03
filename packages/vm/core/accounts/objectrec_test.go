package accounts_test

import (
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

func TestObjectRecordCodec(t *testing.T) {
	bcs.TestCodec(t, &accounts.ObjectRecord{
		ID:  *iotatest.RandomAddress(),
		BCS: []byte{1, 2, 3},
	})
}
