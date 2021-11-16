package request

import (
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/requestargs"
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
	bals := colored.ToL1Map(colored.NewBalancesForIotas(42))
	out := ledgerstate.NewExtendedLockedOutput(bals, addr)
	out.SetID([ledgerstate.OutputIDLength]byte{123})
	return out
}

func TestOnLedger(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		req := OnLedgerFromOutput(rndOutput(), rndAddress(), time.Now())
		reqBack, err := FromMarshalUtil(marshalutil.New(req.Bytes()))
		require.NoError(t, err)
		_, ok := reqBack.(*OnLedger)
		require.True(t, ok)
		require.Equal(t, req.ID(), reqBack.ID())
		require.EqualValues(t, req.Bytes(), reqBack.Bytes())
	})
}

func TestOffLedger(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		target := iscp.Hn("target")
		ep := iscp.Hn("entry point")
		args := requestargs.New()
		req := NewOffLedger(iscp.RandomChainID(), target, ep, args)
		reqBack, err := FromMarshalUtil(marshalutil.New(req.Bytes()))
		require.NoError(t, err)
		_, ok := reqBack.(*OffLedger)
		require.True(t, ok)

		require.EqualValues(t, req.Bytes(), reqBack.Bytes())
	})
}
