// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
)

// Here we implement the derivation of the `vals` (auxVals) variable in the following:
//
// >         ∗ wait until at least (N − f) AUX_r messages have been
// >           received, such that the set of values carried by these
// >           messages, vals are a subset of bin_values_r (note that
// >           bin_values_r may continue to change as BVAL_r messages
// >           are received, thus this condition may be triggered upon
// >           arrival of either an AUX_r or a BVAL_r message)
//
// For this we have to get updates to the binValues variable and exchange the AUX messages.
type varAuxVals struct {
	n         int
	f         int
	nodeIDs   []gpa.NodeID
	recv      map[gpa.NodeID]bool
	readyCB   func(auxVals []bool) gpa.OutMessages
	ready     bool
	round     int
	sent      bool
	binValues []bool
}

func newAuxVals(nodeIDs []gpa.NodeID, f int, readyCB func(auxVals []bool) gpa.OutMessages) *varAuxVals {
	v := &varAuxVals{
		n:         len(nodeIDs),
		f:         f,
		nodeIDs:   nodeIDs,
		recv:      map[gpa.NodeID]bool{},
		readyCB:   readyCB,
		ready:     false,
		round:     -1,
		sent:      false,
		binValues: nil,
	}
	return v
}

func (v *varAuxVals) startRound(round int) {
	v.recv = map[gpa.NodeID]bool{}
	v.ready = false
	v.round = round
	v.sent = false
	v.binValues = nil
}

// >     – wait until bin_values_r != {}, then
// >         ∗ multicast AUX_r(w) where w ∈ bin_values_r
// ...
// >                                                       (note that
// >           bin_values_r may continue to change as BVAL_r messages
// >           are received, thus this condition may be triggered upon
// >           arrival of either an AUX_r or a BVAL_r message)
func (v *varAuxVals) binValuesUpdated(binValues []bool) gpa.OutMessages {
	msgs := gpa.NoMessages()
	if len(binValues) == 1 {
		msgs.AddAll(v.multicast(binValues[0]))
	}
	v.binValues = binValues
	return msgs.AddAll(v.tryOutput())
}

// >         ∗ wait until at least (N − f) AUX_r messages have been
// >           received, such that the set of values carried by these
// >           messages, vals are a subset of bin_values_r ...
func (v *varAuxVals) msgVoteAUXReceived(msg *msgVote) gpa.OutMessages {
	if _, ok := v.recv[msg.sender]; ok {
		return nil // Duplicate.
	}
	v.recv[msg.sender] = msg.value
	return v.tryOutput()
}

// >         ∗ wait until at least (N − f) AUX_r messages have been
// >           received, such that the set of values carried by these
// >           messages, vals are a subset of bin_values_r (note that
// >           bin_values_r may continue to change as BVAL_r messages
// >           are received, thus this condition may be triggered upon
// >           arrival of either an AUX_r or a BVAL_r message)
func (v *varAuxVals) tryOutput() gpa.OutMessages {
	if v.ready || len(v.recv) < v.n-v.f || v.binValues == nil {
		return nil
	}
	hasBinValsT := false
	hasBinValsF := false
	for _, b := range v.binValues {
		if b {
			hasBinValsT = true
		} else {
			hasBinValsF = true
		}
	}
	count := 0
	hasAuxValsT := false
	hasAuxValsF := false
	for _, vote := range v.recv {
		if vote && hasBinValsT {
			count++
			hasAuxValsT = true
			continue
		}
		if !vote && hasBinValsF {
			hasAuxValsF = true
			count++
		}
	}
	if count >= v.n-v.f {
		auxVals := make([]bool, 0, 2)
		if hasAuxValsT {
			auxVals = append(auxVals, true)
		}
		if hasAuxValsF {
			auxVals = append(auxVals, false)
		}
		v.ready = true
		return v.readyCB(auxVals)
	}
	return nil
}

func (v *varAuxVals) multicast(value bool) gpa.OutMessages {
	if v.sent {
		return nil
	}
	v.sent = true
	return multicastMsgVote(v.nodeIDs, v.round, AUX, value)
}

func (v *varAuxVals) statusString() string {
	return fmt.Sprintf("AUX(N=%v,recv=%v)", len(v.nodeIDs), len(v.recv))
}
