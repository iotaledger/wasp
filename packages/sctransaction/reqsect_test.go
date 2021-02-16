package sctransaction

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWriteRead(t *testing.T) {
	cid := coretypes.NewContractID(coretypes.ChainID{}, root.Interface.Hname())
	rsec := NewRequestSectionByWallet(cid, coretypes.EntryPointInit).WithTransfer(nil)
	var buf, buf1 bytes.Buffer
	err := rsec.Write(&buf)
	require.NoError(t, err)
	err = rsec.Read(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	err = rsec.Write(&buf1)
	require.NoError(t, err)
	require.EqualValues(t, buf1.Bytes(), buf.Bytes())
}
