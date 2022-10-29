// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog_test

import (
	"fmt"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/commands"
	"github.com/leanovate/gopter/gen"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/chain/aaa2/cmtLog"
	"github.com/iotaledger/wasp/packages/isc"
)

// L1 can produce some AOs.
// We have:
//   - Confirmed chain.
//   - Unconfirmed chain.
//   - A set of rejected TXes.
//
// L1:
//   - Can confirm AOs.
//   - Can drop them and propose other AOs.
//
// Properties:
//   - If BaseAO is returned, it either last pending, or last confirmed (if no pending).
//   - // TODO: ...
func TestProps(t *testing.T) {
	t.Skip("Test is messed up")
	randAO := func() *isc.AliasOutputWithID {
		return isc.NewAliasOutputWithID(nil, tpkg.RandUTXOInput())
	}
	//
	// L1 Actions.
	genCmdL1NewAO := gen.Const(&commands.ProtoCommand{
		Name: "L1-NewAO",
		NextStateFunc: func(s commands.State) commands.State {
			fmt.Printf("L1: NewAO\n")
			// E.g. external rotation of a TX by other chain.
			// If some external entity produced and AO and it was confirmed,
			// all the TX'es proposed by us and not yet confirmed will be rejected.
			st := s.(*ptState)
			st.confirmed = append(st.confirmed, randAO())
			st.rejected = append(st.rejected, st.pending...)
			st.pending = []*isc.AliasOutputWithID{}
			return st
		},
	})
	genCmdL1Approve := gen.Const(&commands.ProtoCommand{
		Name: "L1-Approve",
		PreConditionFunc: func(s commands.State) bool {
			return len(s.(*ptState).pending) != 0
		},
		NextStateFunc: func(s commands.State) commands.State {
			fmt.Printf("L1: Approve\n")
			// E.g. Consensus TX was approved.
			// Take single TX from the pending log and approve it.
			st := s.(*ptState)
			st.confirmed = append(st.confirmed, st.pending[0])
			st.pending = st.pending[1:]
			return st
		},
	})
	genCmdL1Reject := gen.Const(&commands.ProtoCommand{
		Name: "L1-Reject",
		PreConditionFunc: func(s commands.State) bool {
			return len(s.(*ptState).pending) != 0
		},
		NextStateFunc: func(s commands.State) commands.State {
			fmt.Printf("L1: Reject\n")
			// E.g. Consensus TX was rejected.
			// All the pending TXes are marked as rejected.
			st := s.(*ptState)
			st.rejected = append(st.rejected, st.pending...)
			st.pending = []*isc.AliasOutputWithID{}
			return st
		},
	})
	//
	// CmtLog Actions.

	// cmdAliasOutputConfirmed := &commands.ProtoCommand{
	// 	Name: "CmtLog-AOConfirmed",
	// 	PreConditionFunc: func(s commands.State) bool {
	// 		return false
	// 		// return len(s.(state).pending) > 0
	// 	},
	// 	RunFunc: func(cl commands.SystemUnderTest) commands.Result {
	// 		// cl.(cmtLog.VarLocalView).AliasOutputConfirmed()
	// 		fmt.Printf("XXX: Run\n")
	// 		return nil
	// 	},
	// 	NextStateFunc: func(s commands.State) commands.State {
	// 		fmt.Printf("XXX: Next\n")
	// 		return s
	// 	},
	// }

	// cmdAliasOutputRejected := &commands.ProtoCommand{
	// 	Name: "CmtLog-AORejected",
	// }
	// cmdConsensusOutputDone := &commands.ProtoCommand{
	// 	Name: "CmtLog-AOPosted",
	// }
	genCmdGetBaseAliasOutput := gen.Const(&commands.ProtoCommand{
		Name: "GET",
		RunFunc: func(sut commands.SystemUnderTest) commands.Result {
			return sut.(cmtLog.VarLocalView).GetBaseAliasOutput()
		},
		PostConditionFunc: func(state commands.State, result commands.Result) *gopter.PropResult {
			// if result == (*isc.AliasOutputWithID)(nil) { // TODO: ...
			return &gopter.PropResult{Status: gopter.PropTrue}
			// }
			// return &gopter.PropResult{Status: gopter.PropFalse}
		},
	})

	//
	// The test itself.
	props := gopter.NewProperties(gopter.DefaultTestParameters())
	props.Property("first", commands.Prop(&commands.ProtoCommands{
		NewSystemUnderTestFunc: func(initialState commands.State) commands.SystemUnderTest {
			fmt.Printf("XXX: NEW ----------------------------------------\n")
			initialState.(*ptState).Reset()
			return cmtLog.NewVarLocalView()
		},
		InitialStateGen: gen.Const((&ptState{}).Reset()),
		DestroySystemUnderTestFunc: func(sut commands.SystemUnderTest) {
			// TODO: ...
		},
		GenCommandFunc: func(s commands.State) gopter.Gen {
			st := s.(*ptState)
			//
			// Models the L1.
			cs := []gopter.Gen{
				genCmdL1NewAO,
				genCmdL1Approve,
				genCmdL1Reject,
			}
			//
			// Reacts to L1.
			cs = append(cs, genCmdGetBaseAliasOutput)
			if len(st.confirmed) > 0 {
				cs = append(cs, gen.Const(&ptCmdAOConfirm{value: st.confirmed[len(st.confirmed)-1]}))
			}
			// cs = append(cs, Map(st.confirmed, func(ao *isc.AliasOutputWithID) gopter.Gen {
			// 	return gen.Const(&ptCmdAOConfirm{value: ao})
			// })...)
			cs = append(cs, Map(st.rejected, func(ao *isc.AliasOutputWithID) gopter.Gen {
				return gen.Const(&ptCmdAOReject{value: ao})
			})...)
			return gen.OneGenOf(cs...)
			// return gen.OneConstOf(
			// 	cmdConsensusOutputDone,  // Reacts to L1.
			// )
		},
	}))
	require.True(t, props.Run(gopter.ConsoleReporter(true)))
}

func Map[T, V any](ts []T, fn func(T) V) []V {
	result := make([]V, len(ts))
	for i := range ts {
		result[i] = fn(ts[i])
	}
	return result
}

func asPropResult(res bool) *gopter.PropResult {
	if res {
		return &gopter.PropResult{Status: gopter.PropTrue}
	}
	return &gopter.PropResult{Status: gopter.PropFalse}
}

type ptState struct {
	confirmed []*isc.AliasOutputWithID
	rejected  []*isc.AliasOutputWithID
	pending   []*isc.AliasOutputWithID
	suspended bool // TODO: Use.
}

func (st *ptState) Reset() *ptState {
	st.confirmed = []*isc.AliasOutputWithID{}
	st.rejected = []*isc.AliasOutputWithID{}
	st.pending = []*isc.AliasOutputWithID{}
	return st
}

// If BaseAO is returned, it either last pending, or last confirmed (if no pending).
func (st *ptState) PropReturnedIsLatest(val *isc.AliasOutputWithID) bool {
	fmt.Printf("XXX: PropReturnedIsLatest=%v\n", val)
	if st.suspended {
		return val == nil
	}
	if val == nil {
		return true
	}
	return st.isLatest(val)
}

func (st *ptState) isLatest(val *isc.AliasOutputWithID) bool {
	if len(st.pending) > 0 {
		return val.Equals(st.pending[len(st.pending)-1])
	}
	fmt.Printf("XXX: val=%v\n", val)
	for i := range st.confirmed {
		fmt.Printf("XXX: confirmed[%v]=%v\n", i, st.confirmed[i])
	}
	return val.Equals(st.confirmed[len(st.confirmed)-1])
}

// /////////////////////////////////////////////////////////////////////////////
// ptCmdAOConfirm

type ptCmdAOConfirm struct {
	value *isc.AliasOutputWithID
}

var _ commands.Command = &ptCmdAOConfirm{}

func (cmd *ptCmdAOConfirm) PreCondition(s commands.State) bool {
	fmt.Printf("XXX: AOConfirm[pre], val=%v\n", cmd.value)
	st := s.(*ptState)
	for _, c := range st.confirmed {
		if c.Equals(cmd.value) {
			return true
		}
	}
	return false
}

func (cmd *ptCmdAOConfirm) Run(sut commands.SystemUnderTest) commands.Result {
	fmt.Printf("XXX: AOConfirm[run], val=%v\n", cmd.value)
	cl := sut.(cmtLog.VarLocalView)
	cl.AliasOutputConfirmed(cmd.value)
	return cl.GetBaseAliasOutput()
}

func (cmd *ptCmdAOConfirm) NextState(s commands.State) commands.State {
	fmt.Printf("XXX: AOConfirm[nxt], val=%v\n", cmd.value)
	st := s.(*ptState)
	for i, c := range st.confirmed {
		if c.Equals(cmd.value) {
			st.confirmed = append(st.confirmed[0:i], st.confirmed[i:]...)
			return st
		}
	}
	panic("have to have a value here")
}

func (cmd *ptCmdAOConfirm) PostCondition(s commands.State, r commands.Result) *gopter.PropResult {
	fmt.Printf("XXX: AOConfirm[post], val=%v\n", cmd.value)
	return asPropResult(s.(*ptState).PropReturnedIsLatest(r.(*isc.AliasOutputWithID)))
}

func (cmd *ptCmdAOConfirm) String() string {
	return "AOConfirm"
}

// /////////////////////////////////////////////////////////////////////////////
// ptCmdAOReject

type ptCmdAOReject struct {
	value *isc.AliasOutputWithID
}

func (cmd *ptCmdAOReject) PreCondition(s commands.State) bool {
	st := s.(*ptState)
	for _, c := range st.rejected {
		if c.Equals(cmd.value) {
			return true
		}
	}
	return false
}

func (cmd *ptCmdAOReject) Run(sut commands.SystemUnderTest) commands.Result {
	fmt.Printf("XXX: AOReject, val=%v\n", cmd.value)
	return sut.(cmtLog.VarLocalView).AliasOutputRejected(cmd.value)
}

func (cmd *ptCmdAOReject) NextState(s commands.State) commands.State {
	st := s.(*ptState)
	for i, c := range st.rejected {
		if c.Equals(cmd.value) {
			st.rejected = append(st.rejected[0:i], st.rejected[i:]...)
			return st
		}
	}
	panic("have to have a value here")
}

func (cmd *ptCmdAOReject) PostCondition(s commands.State, result commands.Result) *gopter.PropResult {
	return &gopter.PropResult{Status: gopter.PropTrue}
}

func (cmd *ptCmdAOReject) String() string {
	return "AOReject"
}
