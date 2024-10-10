package accounts_test

import (
	"testing"

	"github.com/iotaledger/wasp/clients/iota-go/sui/suitest"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

func TestObjectRecordCodec(t *testing.T) {
	bcs.TestCodec(t, &accounts.ObjectRecord{
		ID:  *suitest.RandomAddress(),
		BCS: []byte{1, 2, 3},
	})
}
