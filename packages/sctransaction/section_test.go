package sctransaction

import (
	"bytes"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWriteRead(t *testing.T) {
	chainID := coretypes.ChainID(ledgerstate.NewBLSAddress([]byte("dummy public key")).Array())
	cid := coretypes.NewContractID(chainID, root.Interface.Hname())
	sender := blob.Interface.Hname()
	args := codec.MakeDict(map[string]interface{}{
		"param1": 21,
		"param2": chainID,
	})
	rargs := requestargs.New()
	rargs.AddEncodeSimpleMany(args)
	rsec := NewRequestSection(sender, cid, coretypes.EntryPointInit).WithArgs(rargs)
	var buf, buf1 bytes.Buffer
	err := rsec.Write(&buf)
	require.NoError(t, err)
	err = rsec.Read(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	err = rsec.Write(&buf1)
	require.NoError(t, err)
	require.EqualValues(t, buf1.Bytes(), buf.Bytes())
}
