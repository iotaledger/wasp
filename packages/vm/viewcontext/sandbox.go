package viewcontext

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
)

type sandboxview struct {
	vctx     *viewcontext
	params   codec.ImmutableCodec
	state    codec.ImmutableMustCodec
	contract coretypes.Hname
	chainID  coretypes.ChainID
}

func NewSandboxView(vctx *viewcontext, chainID coretypes.ChainID, contractHname coretypes.Hname, params codec.ImmutableCodec) *sandboxview {
	return &sandboxview{
		vctx:     vctx,
		params:   params,
		state:    codec.NewMustCodec(subrealm.New(vctx.state, kv.Key(contractHname.Bytes()))),
		contract: contractHname,
		chainID:  chainID,
	}
}

func (s *sandboxview) Params() codec.ImmutableCodec {
	return s.params
}

func (s *sandboxview) State() codec.ImmutableMustCodec {
	return s.state
}

func (s *sandboxview) MyBalances() coretypes.ColoredBalances {
	panic("not implemented") // TODO: Implement
}

func (s *sandboxview) Call(contractHname coretypes.Hname, entryPoint coretypes.Hname, params codec.ImmutableCodec) (codec.ImmutableCodec, error) {
	return s.vctx.CallView(contractHname, entryPoint, params)
}

func (s *sandboxview) MyContractID() coretypes.ContractID {
	return coretypes.NewContractID(s.chainID, s.contract)
}

func (s *sandboxview) Event(msg string) {
	fmt.Printf("====================   VIEWMSG %s: %s\n", s.MyContractID(), msg)
	//s.vctx.Log().Infof("VMMSG contract %s '%s'", s.MyContractID().String(), msg)
	//s.vctx.Publish(msg)
}

func (s *sandboxview) Eventf(format string, args ...interface{}) {
	msgf := fmt.Sprintf(format, args...)
	fmt.Printf("====================   VIEWMSG %s: %s\n", s.MyContractID(), msgf)
	//s.vctx.Log().Infof("VMMSG: "+format, args...)
	//s.vctx.Publishf(format, args...)
}
