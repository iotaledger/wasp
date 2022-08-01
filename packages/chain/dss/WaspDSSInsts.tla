------------------------- MODULE WaspDSSInsts ----------------------------------
(******************************************************************************)
(*
This specification models the use of DSS nonces in the consensus rounds.
  - The main concern -- nonces must be used once.
  - Nodes have to agree which nonces to use.
  - Faulty nodes should be not able to force nodes to drop their nonces as used.
  - Each chain consensus can take several nonces for signing its TXes.
  - Each node proposes its unused instances only, and don't use its nonce twice,
    even if other nodes decided to use that instance (i.e. don't participate in
    signing, unless the same message is being signed).

Emphasis:

  - Not only two final signatures cannot be produced by the same nonce,
    but also a node cannot make two partial signatures with a nonce.

  - Quorums. If we select nonce instance proposals offered by N-F nodes,
    then we can have to use F+1 for signing. That's because F of those N-F
    can be malicious. We have similar situation with the BLS for randomness.

Conclusions:

  - Have to review the consensus as a replicated state machine. Decided batches
    have to remain decided, even if not published. That will allow to deduce
    the used and unused nonces. No one can even try to reuse the same Nonce
    instance, thus there has to be consensus on already used nonces. E.g.
    consider all nonces used, that were planned to be used in the previous
    rounds / chain blocks, even if the corresponding block/output wasn't
    published or confirmed.

*)
(******************************************************************************)
EXTENDS FiniteSets, Naturals, TLC
CONSTANT Fair    \* A set of Fair node identifiers.
CONSTANT Faulty  \* A set of Faulty node identifiers.
CONSTANT SIds    \* A set of Signature instance IDs to consider.
CONSTANT TIds    \* A set of TX IDs to consider.

VARIABLE dss     \* State of the DSS instances at each node.
VARIABLE txSIds  \* SIds decided to use for the specific TX.
VARIABLE txDone  \* Decision state for each TX.
VARIABLE msgs    \* Messages that were sent.
vars == <<dss, txSIds, txDone, msgs>>

(*
Define node sets, quorums and assumptions on constants.
*)
Nodes == Fair \cup Faulty
N == Cardinality(Nodes)
F == Cardinality(Faulty)
QNF == {q \in SUBSET Nodes : Cardinality(q) = N-F}
QF1 == {q \in SUBSET Nodes : Cardinality(q) = F+1}
ASSUME Assms ==
    /\ IsFiniteSet(Fair \cup Faulty)            \* Node sets are finite.
    /\ Fair \cap Faulty = {}                    \* Node should be either Fair or Faulty.
    /\ 3*F+1 <= N                               \* Byzantine quorum assumption.
    /\ \A sid \in SIds: 0..sid \subseteq SIds   \* SIds are subsequences of N starting at 0.
    /\ \A tid \in TIds: 0..tid \subseteq TIds   \* TIds are subsequences of N starting at 0.

(*
Message types.
*)
Msgs == UNION {
    [t: {"INST_PROPOSAL", "INST_START"}, sid: SIds, src: Nodes],
    [t: {"TX_PROPOSAL"}, tid: TIds, sids: SUBSET SIds, src: Nodes]
}

(*
Type correctness.
*)
TypeOK ==
    /\ dss     \in [Fair -> [SIds -> {"PENDING", "GENERATING", "READY", "USED"}]]
    /\ txSIds  \in [TIds -> SUBSET SIds] \* That's global knowledge, implemented by a consensus.
    /\ txDone  \in [TIds -> BOOLEAN]     \* That's global knowledge, implemented by a consensus.
    /\ msgs    \subseteq Msgs

(*
Utility definitions.
*)
IsNextFreeTX(tid) ==
    /\ ~txDone[tid]                            \* It is not finalized yet.
    /\ \A t \in TIds: ~txDone[t] => (tid <= t) \* It is the smallest such.

IsNextPendingSId(n, sid) ==
    /\ dss[n][sid] = "PENDING"                               \* Is still pending.
    /\ \A s \in SIds : dss[n][sid] = "PENDING" => (sid <= s) \* It is the smallest such.

--------------------------------------------------------------------------------
(******************************************************************************)
(*`^\center{\textbf{             ACTIONS: DSS Instances                   }}^'*)
(*
Advancing SIDs:
  - Initially it is PENDING.
  - A node can issue proposal to start an instance by publishing INST_PROPOSAL message.
    The node does that in sequence, only for the first pending SId.
  - A node starts the instance (its state changed to GENERATING),
    if it receives F+1 proposals to start.
  - A nonce becomes generated (READY) after N-F nodes start to generate it,
    including the node itself.
  - A ready nonce can be used (USED) to make a partial signature.
*)
(******************************************************************************)

(*
Invoked by a node when it wants to start new DSS instance (to generate a nonce).
*)
InstPropose(n) == \E sid \in SIds :
    /\ IsNextPendingSId(n, sid)
    /\ ~\E m \in msgs : m.t = "INST_PROPOSAL" /\ m.src = n /\ m.sid = sid
    /\ msgs' = msgs \cup {[t |-> "INST_PROPOSAL", sid |-> sid, src |-> n]}
    /\ UNCHANGED <<dss, txSIds, txDone>>

(*
Node actually starts an instance (or starts to participate) if it receives F+1
proposals to start that instance.
*)
InstStart(n) == \E q \in QF1, sid \in SIds:
    /\ \A qn \in q: \E m \in msgs: m.t = "INST_PROPOSAL" /\ m.src = qn /\ m.sid = sid \* Enough of proposals.
    /\ ~\E m \in msgs: m.t = "INST_START" /\ m.src = n /\ m.sid = sid                 \* We haven't started yet.
    /\ dss' = [dss EXCEPT ![n][sid] = "GENERATING"]
    /\ msgs' = msgs \cup {[t |-> "INST_START", sid |-> sid, src |-> n]}
    /\ UNCHANGED <<txSIds, txDone>>

(*
This action models the event, when a node has prepared the index proposal for
a particular DSS instance. We don't mode the DSS itself, so just make this
event happen when there is enough messages indicating DSS start.
*)
InstPrepared(n) == \E q \in QNF, sid \in SIds:
    /\ \A qn \in q: \E m \in msgs: m.t = "INST_START" /\ m.src = qn /\ m.sid = sid
    /\ dss[n][sid] = "GENERATING"
    /\ dss' = [dss EXCEPT ![n][sid] = "READY"]
    /\ UNCHANGED <<txSIds, txDone, msgs>>


--------------------------------------------------------------------------------
(******************************************************************************)
(*`^\center{\textbf{             ACTIONS: TX                              }}^'*)
(*
Advancing TXes:
  - The current TX is the one with the smallest TId and is not done yet.
  - A TX can be in several stages (changes sequentially):
      - started (by marking the previous TX as done),
      - being decided (TX proposals are published until some quorum is reached).
      - decided (by the ACS),
      - being signed (partial signatures are issued by using nonces).
      - done (i.e. signed everything, that we wanted, now go to the next TX).
*)
(******************************************************************************)
(*
This action models an event, when a node proposes its input to the vector consensus.
The proposal includes also Ids of DSS instances proposed to be used.
*)
TxProposal(n) == \E tid \in TIds : IsNextFreeTX(tid) /\
    LET txNotDecided == txSIds[tid] = {}
        msgNotSent   == ~\E m \in msgs : m.t = "TX_PROPOSAL" /\ m.src = n
        proposedSIds == { sid \in SIds : dss[n][sid] = "READY"}
    IN  /\ txNotDecided
        /\ msgNotSent
        /\ msgs' = msgs \cup {[t |-> "TX_PROPOSAL", tid |-> tid, sids |-> proposedSIds, src |-> n]}
        /\ UNCHANGED <<dss, txSIds, txDone>>

(*
This action models decision of a consensus.
We have to take SID proposals proposed by N-F nodes.
That's because F of them can be byzantine, thus we will have at least F+1 proposals from fair nodes.
The Signing round should then use threshold not greater than F+1.
*)
TxDecision == \E q \in QNF, tid \in TIds : IsNextFreeTX(tid) /\
    LET enoughParties == \A qn \in q : \E m \in msgs : m.t = "TX_PROPOSAL" /\ m.src = qn
        decidedSIds == {sid \in SIds :
            \A qn \in q :
                \E m \in msgs : m.t = "TX_PROPOSAL" /\ m.src = qn /\ sid \in m.sids
        }
    IN  /\ txSIds[tid] = {} \* Not decided yet.
        /\ enoughParties    \* We have enough participants in the consensus.
        /\ txSIds' = [txSIds EXCEPT ![tid] = decidedSIds]
        /\ UNCHANGED <<dss, txDone, msgs>>

(*
Nodes issue partial signatures on a TX, when TX is decided (ACS produced a decision),
but not done yet (TXes are signed, and we are about to work on the next TX).
This action is per node.
Assume any of the proposed SIds in a TX can be used to make the partial
signature (not necessary in their sequence order).
After using a nonce, it is marked as used to make sure it will not be used again.
*)
TxSign(n) == \E sid \in SIds, tid \in TIds :
    /\ IsNextFreeTX(tid)
    /\ sid \in txSIds[tid]
    /\ dss[n][sid] = "READY"
    /\ dss' = [dss EXCEPT ![n][sid] = "USED"]
    /\ UNCHANGED  <<txSIds, txDone, msgs>>

(*
Nodes can advance the TX at any time, actually.
*)
TxDone == \E tid \in TIds :
    /\ IsNextFreeTX(tid)
    /\ txDone' = [txDone EXCEPT ![tid] = TRUE]
    /\ UNCHANGED <<dss, txSIds, msgs>>

--------------------------------------------------------------------------------
Init ==
    /\ dss     = [n \in Fair |-> [sid \in SIds |-> "PENDING"]]
    /\ txSIds  = [t \in TIds |-> {}]
    /\ txDone  = [t \in TIds |-> FALSE]
    /\ msgs    = {m \in Msgs : m.src \in Faulty}
Next == \E n \in Fair :
    \/ InstPropose(n)
    \/ InstStart(n)
    \/ InstPrepared(n)
    \/ TxProposal(n)
    \/ TxDecision
    \/ TxSign(n)
    \/ TxDone
Fairness == WF_vars(Next)
Spec == Init /\ [][Next]_vars /\ Fairness
--------------------------------------------------------------------------------
\* // TODO: Properties...

SomeTXGetsDone == <>txDone[0]

THEOREM Spec => []TypeOK
PROOF OMITTED \* Checked by the TLC.

TLC_Fair_permutations == Permutations(Fair) \* For TLC, to be used as a symmetry set.
================================================================================
