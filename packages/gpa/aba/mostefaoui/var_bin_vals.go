// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
)

// Represents the `binValues` variable and sends/handles the BVAL messages.
// This type implements the following logic from the algorithm:
//
// >     – multicast BVAL_r(est_r)
// >     – bin_values_r := {}
// >     – upon receiving BVAL_r(b) messages from f + 1 nodes, if
// >       BVAL_r(b) has not been sent, multicast BVAL_r(b)
// >     – upon receiving BVAL_r(b) messages from 2f + 1 nodes,
// >       bin_values_r := bin_values_r ∪ {b}
// >     – wait until bin_values_r != {}, then
type varBinVals struct {
	n         int
	f         int
	nodeIDs   []gpa.NodeID
	updateCB  func(binVals []bool) gpa.OutMessages
	round     int
	est       bool
	recvT     map[gpa.NodeID]bool
	recvF     map[gpa.NodeID]bool
	sentT     bool
	sentF     bool
	binValues []bool
}

func newBinVals(nodeIDs []gpa.NodeID, f int, updateCB func(binVals []bool) gpa.OutMessages) *varBinVals {
	v := &varBinVals{
		n:        len(nodeIDs),
		f:        f,
		nodeIDs:  nodeIDs,
		updateCB: updateCB,
	}
	return v
}

// >     – multicast BVAL_r(est_r)
// >     – bin_values_r := {}
func (v *varBinVals) startRound(round int, est bool) gpa.OutMessages {
	v.round = round
	v.est = est
	v.recvT = map[gpa.NodeID]bool{}
	v.recvF = map[gpa.NodeID]bool{}
	v.sentT = false
	v.sentF = false
	v.binValues = []bool{}
	return v.multicast(v.est)
}

// >     – upon receiving BVAL_r(b) messages from f + 1 nodes, if
// >       BVAL_r(b) has not been sent, multicast BVAL_r(b)
// >     – upon receiving BVAL_r(b) messages from 2f + 1 nodes,
// >       bin_values_r := bin_values_r ∪ {b}
// >     – wait until bin_values_r != {}, then
func (v *varBinVals) msgVoteBVALReceived(msg *msgVote) gpa.OutMessages {
	recv := v.recv(msg.value) // NOTE: A reference to a field.

	if ok := recv[msg.sender]; ok {
		return nil // Duplicate.
	}
	recv[msg.sender] = true

	msgs := gpa.NoMessages()
	if len(recv) == v.f+1 {
		msgs.AddAll(v.multicast(msg.value)) // This checks, if already sent.
	}

	if len(recv) == 2*v.f+1 {
		v.binValues = append(v.binValues, msg.value)
		return msgs.AddAll(v.updateCB(v.binValues))
	}
	return msgs
}

// Misc helpers.

func (v *varBinVals) sent(value bool) *bool {
	if value {
		return &v.sentT
	}
	return &v.sentF
}

func (v *varBinVals) recv(value bool) map[gpa.NodeID]bool {
	if value {
		return v.recvT
	}
	return v.recvF
}

func (v *varBinVals) multicast(value bool) gpa.OutMessages {
	sent := v.sent(value)
	if *sent {
		return nil
	}
	*sent = true
	return multicastMsgVote(v.nodeIDs, v.round, BVAL, value)
}

func (v *varBinVals) statusString() string {
	return fmt.Sprintf("BIN(N=%v,T=%v,F=%v)", len(v.nodeIDs), len(v.recvT), len(v.recvF))
}
