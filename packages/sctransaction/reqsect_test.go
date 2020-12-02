package sctransaction

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/coret"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWriteRead(t *testing.T) {
	cid := coret.NewContractID(coret.ChainID{}, root.Interface.Hname())
	rsec := NewRequestSectionByWallet(cid, coret.EntryPointInit).WithTransfer(nil)
	var buf, buf1 bytes.Buffer
	err := rsec.Write(&buf)
	require.NoError(t, err)
	err = rsec.Read(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	err = rsec.Write(&buf1)
	require.NoError(t, err)
	require.EqualValues(t, buf1.Bytes(), buf.Bytes())
}
