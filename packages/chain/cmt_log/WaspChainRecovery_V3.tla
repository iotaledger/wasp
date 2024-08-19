---- MODULE WaspChainRecovery_V3 -----------------------------------------------
(*
Here we only consider the recovery of the consensus.
*)
EXTENDS
    Integers,
    Naturals,
    WaspByzEnv  \* Defines CN, FN, AN, N, F, Q?F.

CONSTANT MaxLI
CONSTANT Values \* Consensus the consensus decides on.
CONSTANT AOs    \* A set of possible AOs.

ASSUME Assms == MaxLI \in Nat


(* The L1 state. *)
\* VARIABLE remainingAOs   \* That's for having a finite model.
\* VARIABLE latestAO       \* TODO -- Can I avoid having those AOs?
(*
    If we don't check the convergence of the AOs, just a recovery, then AOs should be irrelevant.
    I.e. the nodes could propose any aos, and the decision is on any AO as well.
*)

(* The chain/committee state. *)
VARIABLE persistentLI
VARIABLE agreedLI
VARIABLE minLI
VARIABLE msgs
VARIABLE qcL1RepAO
VARIABLE qcConsOut
VARIABLE qcRecover
VARIABLE qcStarted
vars == <<persistentLI, agreedLI, minLI, msgs, qcL1RepAO, qcConsOut, qcRecover, qcStarted>>

LI == 1..MaxLI
NoLI == 0
AnyLI == 0..MaxLI

NoVal == CHOOSE nv : nv \notin Values

NLI_Kind == {"NLI_ConsOut", "NLI_L1RepAO", "NLI_Recover", "NLI_Started"}
NLI_Msg(k, li, n) == [t |-> k, li |-> li, n |-> n]

Cons_MsgT == {"C_PROP", "C_DEC"}
Cons_Msg(t, li, n, val) == [t |-> t, li |-> li, n |-> n, val |-> val]

Msgs == UNION {
    [t : Cons_MsgT, li : LI, n : AN, val : Values], \* Consensus proposals and decisions.
    [t : NLI_Kind, li : LI, n : AN]                 \* Recovery messages.
}


--------------------------------------------------------------------------------
(* Helper: A quorum counter (QC).
   Corresponds to `packages/chain/cmt_log/quorum_counter.go`.
*)
QCTypeValSet ==
    [
        kind : NLI_Kind,        \* What kind of messages this QC is working with?
        recv : [AN -> AnyLI],   \* Votes received from the peer.
        sent : AnyLI            \* What is the last vote we have sent?
    ]
QCInit(k) ==
    [
        kind |-> k,
        recv |-> [ n \in AN |-> NoLI],
        sent |-> NoLI
    ]
QCReset(qc) ==
    [qc EXCEPT
        !.recv = [ n \in AN |-> NoLI],
        !.sent = NoLI
    ]
QCWithRecv(qc, n, li) ==
    [qc EXCEPT !.recv[n] = IF li > @ THEN li ELSE @]
QCNotSent(qc, li) ==
    li > qc.sent
QCTrySend(qc, n, li) ==
    IF li > qc.sent
    THEN [qc |-> [qc EXCEPT !.sent = li], msgs |-> {NLI_Msg(qc.kind, li,n)}]
    ELSE [qc |-> qc,                      msgs |-> {}                      ]
QCSetAndTrySend(qc, n, li) ==
    QCTrySend(QCWithRecv(qc, n, li), n, li)

QCReached(qc, qs, li) ==
    \E q \in qs : \A qn \in q : qc.recv[qn] >= li
QCMax(qc, qs, li) ==
    /\ QCReached(qc, qs, li)
    /\ \A li2 \in LI: li2 > li => ~QCReached(qc, qs, li2)

--------------------------------------------------------------------------------
TypeOK ==
    /\ persistentLI \in [CN -> LI \cup {NoLI}]
    /\ agreedLI     \in [CN -> LI \cup {NoLI}]
    /\ minLI        \in [CN -> LI]
    /\ qcL1RepAO    \in [CN -> QCTypeValSet]
    /\ qcConsOut    \in [CN -> QCTypeValSet]
    /\ qcRecover    \in [CN -> QCTypeValSet]
    /\ qcStarted    \in [CN -> QCTypeValSet]
    /\ msgs         \in SUBSET Msgs


--------------------------------------------------------------------------------
(* Actions: Consensus, as a black-box.                                        *)

sendConsensusInput(li, n) == \* TODO: Remove???
    \E v \in Values : \* Any value van be proposed.
        msgs' = msgs \cup {[t |-> "C_PROP", li |-> li, n |-> n, val |-> v]}

consensusInput(li, n, val) ==
    Cons_Msg("C_PROP", li, n, val)

ConsensusDecided(li) ==
    \* TODO: Do we need to check if the nodes are alive and participate?
    \E v \in Values, q \in QNF :
        /\ ~\E m \in msgs : \* Not decided for this LI yet.
                /\ m.t = "C_DEC"
                /\ m.li = li
        /\ \A n \in q : \* Have a quorum of inputs.
                \E m \in msgs :
                    /\ m.t  = "C_PROP"
                    /\ m.li = li
                    /\ m.n  = n
        /\ \E n \in q \cap CN :
                \E m \in msgs : \* A correct node proposed v.
                    /\ m.t   = "C_PROP"
                    /\ m.li  = li
                    /\ m.n   = n
                    /\ m.val = v
        /\ msgs' = msgs \cup [t : {"C_DEC"}, li : {li}, n : CN, val : {v}]
        /\ UNCHANGED <<persistentLI, agreedLI, minLI, qcL1RepAO, qcConsOut, qcRecover, qcStarted>>

hasConsensusDecidedValue(li, n, v) ==
    \E m \in msgs : m.t = "C_DEC" /\ m.li = li /\ m.n = n /\ m.val = v


--------------------------------------------------------------------------------
(* Actions: Handling local events.                                            *)

(* Boot or reboot a node.
  - Load the last known persistent LI from the store, set the minLI based on that.
  - Reset the agreed LI and all the counters.
*)
Reboot(n) == \* Only the state is lost here. The recover messages are on reception of L1 AO.
    /\ minLI'     = [minLI     EXCEPT ![n] = persistentLI[n]+1]
    /\ agreedLI'  = [agreedLI  EXCEPT ![n] = NoLI]
    /\ qcConsOut' = [qcConsOut EXCEPT ![n] = QCReset(@)]
    /\ qcL1RepAO' = [qcL1RepAO EXCEPT ![n] = QCReset(@)]
    /\ qcRecover' = [qcRecover EXCEPT ![n] = QCReset(@)]
    /\ qcStarted' = [qcStarted EXCEPT ![n] = QCReset(@)]
    /\ UNCHANGED <<persistentLI, msgs>>

ConsensusOutputReceived(n) ==
    \E consensusLI \in LI :
        /\  \E v \in Values : hasConsensusDecidedValue(consensusLI, n, v)
        /\  LET resConsOut == QCSetAndTrySend(qcConsOut[n], n, consensusLI + 1)
            IN  /\ qcConsOut' = [qcConsOut EXCEPT ![n] = resConsOut.qc]
                /\ msgs'      = msgs \cup resConsOut.msgs
                /\ UNCHANGED <<persistentLI, agreedLI, minLI, qcL1RepAO, qcRecover, qcStarted>>

ConsensusRecoverReceived(n) ==
    \E consensusLI \in LI :
        /\  \E v \in Values : consensusInput(consensusLI, n, v) \in msgs \* Was started.
        /\  LET tmpRecover == QCSetAndTrySend(qcRecover[n], n, consensusLI + 1)
            IN  /\ qcRecover' = [qcL1RepAO EXCEPT ![n] = tmpRecover.qc]
                /\ msgs'      = msgs \cup tmpRecover.msgs
                /\ UNCHANGED <<persistentLI, agreedLI, minLI, qcConsOut, qcL1RepAO, qcStarted>>

L1ReplacedBaseAliasOutput(n) ==
    LET voteForLI == IF agreedLI[n] >= minLI[n] THEN agreedLI[n] + 1 ELSE minLI[n]
        tmpRecover == QCSetAndTrySend(qcRecover[n], n, minLI[n])
        tmpL1RepAO == QCSetAndTrySend(qcL1RepAO[n], n, voteForLI)
    IN  /\  qcRecover' = [qcRecover EXCEPT ![n] = tmpRecover.qc]
        /\  qcL1RepAO' = [qcL1RepAO EXCEPT ![n] = tmpL1RepAO.qc]
        /\  msgs'      = msgs \cup tmpRecover.msgs \cup tmpL1RepAO.msgs
        /\  UNCHANGED <<persistentLI, agreedLI, minLI, qcConsOut, qcStarted>>

L1ConfirmedAliasOutput(n) ==
    \E li \in LI :
        /\  \E v \in Values : hasConsensusDecidedValue(li, n, v) \* Should be at least decided by the consensus.
        /\  LET tmpL1RepAO == QCSetAndTrySend(qcL1RepAO[n], n, li)
            IN  /\  qcL1RepAO' = [qcL1RepAO EXCEPT ![n] = tmpL1RepAO.qc]
                /\  msgs'      = msgs \cup tmpL1RepAO.msgs
                /\  UNCHANGED <<persistentLI, agreedLI, minLI, qcConsOut, qcRecover, qcStarted>>

--------------------------------------------------------------------------------
(* Actions: Handling network messages.                                        *)

Recv_NextLI_ConsOut(n) ==
    \E m \in msgs :
        /\ m.t = "NLI_ConsOut"
        /\ qcConsOut' = [qcConsOut EXCEPT ![n] = QCWithRecv(@, m.n, m.li)]
        /\ UNCHANGED <<persistentLI, agreedLI, minLI, msgs, qcL1RepAO, qcRecover, qcStarted>>

Recv_NextLI_L1RepAO(n) ==
    \E m \in msgs :
        /\ m.t = "NLI_L1RepAO"
        /\ qcL1RepAO' = [qcL1RepAO EXCEPT ![n] = QCWithRecv(@, m.n, m.li)]
        /\ UNCHANGED <<persistentLI, agreedLI, minLI, msgs, qcConsOut, qcRecover, qcStarted>>

Recv_NextLI_Recover(n) ==
    \E m \in msgs :
        /\ m.t = "NLI_Recover"
        /\ qcRecover' = [qcRecover EXCEPT ![n] = QCWithRecv(@, m.n, m.li)]
        /\ UNCHANGED <<persistentLI, agreedLI, minLI, msgs, qcConsOut, qcL1RepAO, qcStarted>>

Recv_NextLI_Started(n) ==
    \E m \in msgs :
        /\ m.t = "NLI_Started"
        /\ qcStarted' = [qcStarted EXCEPT ![n] = QCWithRecv(@, m.n, m.li)]
        /\ UNCHANGED <<persistentLI, agreedLI, minLI, msgs, qcConsOut, qcL1RepAO, qcRecover>>

--------------------------------------------------------------------------------
(* Actions: React to quorums.                                                 *)

\* Helper, not a complete action.
tryDecide(n, li) ==
    /\  li > agreedLI[n]
    /\  li >= minLI[n]
    /\  agreedLI' = [agreedLI EXCEPT ![n] = li]
    /\  LET res == QCTrySend(qcStarted[n], n, li) IN
            /\ qcStarted' = [qcStarted EXCEPT ![n] = res.qc]
            /\ \E val \in Values : msgs' = msgs \cup res.msgs \cup {consensusInput(li, n, val)}
    /\ UNCHANGED <<persistentLI, minLI, qcL1RepAO, qcConsOut, qcRecover>>

\* When N-F votes are collected, try decide that LI.
NLI_ConsOut_QNF(n) ==
    \E li \in LI :
        /\ QCMax(qcConsOut[n], QNF, li)
        /\ tryDecide(n, li)

\* When N-F votes are collected, try decide that LI.
NLI_L1RepAO_QNF(n) ==
    \E li \in LI :
        /\ QCMax(qcL1RepAO[n], QNF, li)
        /\ tryDecide(n, li)


\* Upon reaching F+1 votes, node supports the corresponding LI.
\* And when N-F votes are collected, try decide that LI.
NLI_Recover_Q1F(n) ==
    \E li \in LI :
        /\ QCMax(qcRecover[n], Q1F, li)
        /\ QCNotSent(qcRecover[n], li)
        /\ LET res == QCTrySend(qcRecover[n], n, li)
           IN /\ qcRecover' = [qcRecover EXCEPT ![n] = res.qc]
              /\ msgs' = msgs \cup res.msgs
        /\ UNCHANGED <<persistentLI, agreedLI, minLI, qcConsOut, qcL1RepAO, qcStarted>>
NLI_Recover_QNF(n) ==
    \E li \in LI :
        /\ QCMax(qcRecover[n], QNF, li)
        /\ tryDecide(n, li)


\* Upon the F+1 quorum reached for NLI(Started, li), ...
NLI_Started_Q1F(n) ==
    \E li \in LI :
        /\ QCMax(qcStarted[n], Q1F, li)
        /\ tryDecide(n, li)


--------------------------------------------------------------------------------
(* The specification.                                                         *)

Init ==
    /\ persistentLI = [n \in CN |-> NoLI]
    /\ agreedLI     = [n \in CN |-> NoLI]
    /\ minLI        = [n \in CN |-> 1]
    /\ qcL1RepAO = [ n \in CN |-> QCInit("NLI_L1RepAO") ]
    /\ qcConsOut = [ n \in CN |-> QCInit("NLI_ConsOut") ]
    /\ qcRecover = [ n \in CN |-> QCInit("NLI_Recover") ]
    /\ qcStarted = [ n \in CN |-> QCInit("NLI_Started") ]
    /\ msgs = {}

Next ==
    \/ \E n \in CN:
        \/ Reboot(n)
        \/ ConsensusOutputReceived(n)
        \/ ConsensusRecoverReceived(n)
        \/ L1ReplacedBaseAliasOutput(n)
        \/ L1ConfirmedAliasOutput(n)
        \/ Recv_NextLI_ConsOut(n) \/ NLI_ConsOut_QNF(n)
        \/ Recv_NextLI_L1RepAO(n) \/ NLI_L1RepAO_QNF(n)
        \/ Recv_NextLI_Recover(n) \/ NLI_Recover_QNF(n) \/ NLI_Recover_Q1F(n)
        \/ Recv_NextLI_Started(n) \/ NLI_Started_Q1F(n)
    \/ \E li \in LI :
        \/ ConsensusDecided(li)

Fair ==
    /\  WF_vars(
            \E li \in LI :
                \/ ConsensusDecided(li)
        )
    /\  SF_vars(
            \E n \in CN:
                \* Reboot(n) -- No reboot here.
                \/ ConsensusOutputReceived(n)
                \/ ConsensusRecoverReceived(n)
                \/ L1ReplacedBaseAliasOutput(n)
                \/ L1ConfirmedAliasOutput(n)
                \/ Recv_NextLI_ConsOut(n) \/ NLI_ConsOut_QNF(n)
                \/ Recv_NextLI_L1RepAO(n) \/ NLI_L1RepAO_QNF(n)
                \/ Recv_NextLI_Recover(n) \/ NLI_Recover_QNF(n) \/ NLI_Recover_Q1F(n)
                \/ Recv_NextLI_Started(n) \/ NLI_Started_Q1F(n)
        )

Spec == Init /\ [][Next]_vars /\ Fair

--------------------------------------------------------------------------------
(* Properties.                                                                *)

(*
There will be no duplicate proposals, nor decisions from the consensus.
The property we are checking is uniqueness of the proposals.
The decisions are unique by the definition of the mock consensus.
TODO: Assumption on enough nodes not loosing their state is required.
*)
ConsensusProposalAndDecisionsUnique ==
    \A m1, m2 \in msgs:
        /\ m1.t  = m2.t /\ m1.t \in {"C_PROP", "C_DEC"}
        /\ m1.li = m2.li
        /\ m1.n  = m2.n /\ m1.n \in CN
        => m1.val = m2.val

\* TODO: Progress property: all LIs are used.
EventuallyAllLIsAreUsed ==
    <>\A li \in LI :
        \E n \in CN, v \in Values : Cons_Msg("C_PROP", li, n, v) \in msgs

THEOREM Spec =>
    /\ []TypeOK
    /\ []ConsensusProposalAndDecisionsUnique
    /\ EventuallyAllLIsAreUsed
    PROOF OMITTED \* Model-checked.


--------------------------------------------------------------------------------

LEMMA LIProps == NoLI \in AnyLI
    BY Assms DEF NoLI, AnyLI

LEMMA QCTypeOK ==
    /\ \A k \in NLI_Kind : QCInit(k) \in QCTypeValSet
    /\ TRUE \* TODO: ...
    BY LIProps DEF QCInit, QCTypeValSet


================================================================================
