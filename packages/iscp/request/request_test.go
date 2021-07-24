package request

import (
	"testing"

	"github.com/iotaledger/wasp/packages/iscp/requestargs"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/marshalutil"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/stretchr/testify/require"
)

func TestMetadata(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		sender := iscp.Hn("sender")
		target := iscp.Hn("target")
		ep := iscp.Hn("entryp")
		md := NewMetadata().WithSender(sender).WithTarget(target).WithEntryPoint(ep)

		data := md.Bytes()
		back := MetadataFromBytes(data)
		require.True(t, back.ParsedOk())
		require.NoError(t, back.ParsedError())
		require.EqualValues(t, md.Bytes(), back.Bytes())
	})
	t.Run("parse  error", func(t *testing.T) {
		var data []byte
		md := MetadataFromBytes(data)
		require.False(t, md.ParsedOk())
		require.Error(t, md.ParsedError())
		require.EqualValues(t, 0, md.SenderContract())
		require.EqualValues(t, 0, md.TargetContract())
		require.EqualValues(t, 0, md.EntryPoint())
		require.EqualValues(t, 0, len(md.Args()))
		t.Logf("ParsedError: %v", md.ParsedError())

		data = []byte("random data")
		md = MetadataFromBytes(data)
		require.False(t, md.ParsedOk())
		require.Error(t, md.ParsedError())
		require.EqualValues(t, 0, md.SenderContract())
		require.EqualValues(t, 0, md.TargetContract())
		require.EqualValues(t, 0, md.EntryPoint())
		require.EqualValues(t, 0, len(md.Args()))
		t.Logf("ParsedError: %v", md.ParsedError())

		data = []byte("random data dddddddddddddddddddddddddddddddddd")
		md = MetadataFromBytes(data)
		require.False(t, md.ParsedOk())
		require.Error(t, md.ParsedError())
		require.EqualValues(t, 0, md.SenderContract())
		require.EqualValues(t, 0, md.TargetContract())
		require.EqualValues(t, 0, md.EntryPoint())
		require.EqualValues(t, 0, len(md.Args()))
		t.Logf("ParsedError: %v", md.ParsedError())
	})
}

func rndAddress() ledgerstate.Address {
	kp := ed25519.GenerateKeyPair()
	ret := ledgerstate.NewED25519Address(kp.PublicKey)
	return ret
}

func rndOutput() *ledgerstate.ExtendedLockedOutput {
	addr := rndAddress()
	bals := map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: 42}
	return ledgerstate.NewExtendedLockedOutput(bals, addr)
}

func TestOnLedger(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		req := OnLedgerFromOutput(rndOutput(), rndAddress())
		reqBack, err := FromMarshalUtil(marshalutil.New(req.Bytes()))
		require.NoError(t, err)
		_, ok := reqBack.(*RequestOnLedger)
		require.True(t, ok)

		require.EqualValues(t, req.Bytes(), reqBack.Bytes())
	})
}

func TestOffLedger(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		target := iscp.Hn("target")
		ep := iscp.Hn("entry point")
		args := requestargs.New()
		req := NewOffLedger(target, ep, args)
		reqBack, err := FromMarshalUtil(marshalutil.New(req.Bytes()))
		require.NoError(t, err)
		_, ok := reqBack.(*RequestOffLedger)
		require.True(t, ok)

		require.EqualValues(t, req.Bytes(), reqBack.Bytes())
	})
}
