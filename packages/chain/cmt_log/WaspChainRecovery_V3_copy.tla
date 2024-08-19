---- MODULE WaspChainRecovery_V4 -----------------------------------------------
(*
Here we only consider the recovery of the consensus.
Make it more abstract than the code (as it was attempted in V3).
*)
EXTENDS
    Integers,
    Naturals,
    WaspByzEnv  \* Defines CN, FN, AN, N, F, Q?F.

CONSTANT MaxLI
CONSTANT Values \* Consensus the consensus decides on.

VARIABLE persistentLI
VARIABLE agreedLI
VARIABLE minLI
VARIABLE msgs
VARIABLE qcL1ReAO
VARIABLE qcConsOut
VARIABLE qcRecover
VARIABLE qcStarted
vars == <<persistentLI, agreedLI, minLI, msgs, qcL1ReAO, qcConsOut, qcRecover, qcRecover>>

LI == 1..MaxLI
NoLI == 0

NoVal == CHOOSE nv : nv \notin Values

NLI_Kind == {"NLI_ConsOut", "NLI_L1ReAO", "NLI_Recover", "NLI_Started"}
NLI_Msg(k, li, n) == [t |-> k, li |-> li, n |-> n]

Msgs == UNION {
    [t : {"C_PROP", "C_DEC"}, li : LI, n : AN, val : Values],   \* Consensus proposals and decisions.
    [t : NLI_Kind, li : LI, n : AN]                             \* Recovery messages.
}


--------------------------------------------------------------------------------
(* Helper: A quorum counter (QC).
   Corresponds to `packages/chain/cmt_log/quorum_counter.go`.
*)
QCTypeValSet ==
    [
        kind : NLI_Kind,      \* What kind of messages this QC is working with?
        recv : [AN -> LI],    \* Votes received from the peer.
        sent : LI             \* What is the last vote we have sent?
    ]
QCInit(k) ==
    [
        kind |-> k,
        recv |-> [ n \in AN |-> NoLI],
        sent |-> NoLI
    ]
QCWithRecv(qc, n, li) ==
    [qc EXCEPT !.recv[n] = IF li > @ THEN li ELSE @]
QCTrySend(qc, n, li) ==
    IF li > qc.sent
    THEN [qc |-> [qc EXCEPT !.sent = li], msgs |-> {NLI_Msg(qc.kind, li,n)}]
    ELSE [qc |-> qc,                      msgs |-> {}                      ]
QCReached(qc, qs, li) ==
    \E q \in qs : \A qn \in q : qs.recv[qn] >= li
QCMax(qc, qs, li) ==
    /\ QCReached(qc, qs, li)
    /\ \A li2 \in LI: li2 > li => ~QCReached(qc, qs, li2)

--------------------------------------------------------------------------------
TypeOK ==
    /\ persistentLI \in [CN -> LI]
    /\ agreedLI \in [CN -> LI \cup {NoLI}]
    /\ minLI \in [CN -> LI \cup {NoLI}]
    /\ qcL1ReAO  \in [CN -> QCTypeValSet]
    /\ qcConsOut \in [CN -> QCTypeValSet]
    /\ qcRecover \in [CN -> QCTypeValSet]
    /\ qcStarted \in [CN -> QCTypeValSet]
    /\ msgs \in SUBSET Msgs


--------------------------------------------------------------------------------
(* Actions: Consensus, as a black-box.                                        *)

sendConsensusInput(li, n) ==
    \E v \in Values : \* Any value van be proposed.
        msgs' = msgs \cup {[t |-> "C_PROP", li |-> li, n |-> n, val |-> v]}


ConsensusDecided(li) == \* TODO: Check if the nodes are alive and participate?
    \E li \in LI, v \in Values, q \in QNF :
        /\ ~\E m \in msgs : \* Not decided for this LI yet.
                /\ m.t = "C_DEC"
                /\ m.li = li
        /\ \A n \in q : \* Have a quorum of inputs.
                \E m \in msgs :
                    /\ m.t = "C_PROP"
                    /\ m.li = li
                    /\ m.n = n
        /\ \E n \in q \cap CN :
                \E m \in msgs : \* A correct node proposed v.
                    /\ m.t = "C_PROP"
                    /\ m.li = li
                    /\ m.n = n
                    /\ m.v = v
        /\ msgs' = msgs \cup [t : {"C_PROP"}, li : {li}, n : CN, val : {v}]
        /\ UNCHANGED vars \* TODO: ...

hasConsensusDecidedValue(li, n, v) ==
    \E m \in msgs : m.t = "C_DEC" /\ m.li = li /\ m.n = n /\ m.v = v


--------------------------------------------------------------------------------
(* Actions: Handling local events.                                            *)

(* Boot or reboot a node.
  - Load the last known persistent LI from the store, set the minLI based on that.
  - Reset the agreed LI.
*)
Reboot(n) ==
    /\ minLI' = [minLI EXCEPT ![n] = persistentLI[n]+1]
    /\ agreedLI' = [agreedLI EXCEPT ![n] = NoLI]
    /\ msgs' = msgs \cup {[t |-> "NLI_RepL1AO", li |-> minLI'[n], n |-> n]}
    /\ UNCHANGED vars \* TODO: ...


--------------------------------------------------------------------------------
(* Actions: Handling network messages.                                        *)

Recv_NextLI_ConsOut(n) ==
    \E m \in msgs :
        /\ m.t = "NLI_ConsOut"
        /\ UNCHANGED vars \* TODO: ...

\* Corresponds to msgNextLogIndexOnL1RepAO in packages/chain/cmt_log/var_log_index.go
Recv_NextLI_L1ReAO(n) ==
    \E m \in msgs :
        /\ m.t = "NLI_L1RepAO"
        /\ LET qc2 == QCWithRecv(qcL1ReAO, n, m.li)
               res == QCTrySend(qc2, Q1F, n, m.li)
           IN /\ qcL1ReAO' = res.qc
              /\ msgs'     = msgs \cup res.msgs
        /\ UNCHANGED <<persistentLI, agreedLI, minLI, qcConsOut, qcRecover, qcRecover>>

Recv_NextLI_Recover(n) ==
    \E m \in msgs :
        /\ m.t = "NLI_Recover"
        /\ UNCHANGED vars \* TODO: ...

Recv_NextLI_Started(n) ==
    \E m \in msgs :
        /\ m.t = "NLI_Started"
        /\ UNCHANGED vars \* TODO: ...

--------------------------------------------------------------------------------
(* The specification.                                                         *)

Init ==
    /\ persistentLI = [n \in AN |-> NoLI]
    /\ agreedLI = [n \in AN |-> NoLI]
    /\ minLI = [n \in AN |-> NoLI]
    /\ qcL1ReAO  = QCInit("NLI_L1ReAO")
    /\ qcConsOut = QCInit("NLI_ConsOut")
    /\ qcRecover = QCInit("NLI_Recover")
    /\ qcStarted = QCInit("NLI_Started")
    /\ msgs = {}

Next ==
    \/ \E n \in AN:
        \/ Reboot(n)
        \/ Recv_NextLI_ConsOut(n)
        \/ Recv_NextLI_L1ReAO(n)
        \/ Recv_NextLI_Recover(n)
        \/ Recv_NextLI_Started(n)
    \/ \E li \in LI :
        \/ ConsensusDecided(li)

Fair ==
    /\ WF_vars(\E li \in LI : ConsensusDecided(li))

Spec == Init /\ [][Next]_vars

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

THEOREM Spec =>
    /\ []TypeOK
    /\ ConsensusProposalAndDecisionsUnique
    PROOF OMITTED \* Model-checked.

================================================================================
