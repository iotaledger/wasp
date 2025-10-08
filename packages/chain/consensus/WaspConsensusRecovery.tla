---- MODULE WaspConsensusRecovery ----------------------------------------------
(*
The goal in this specification to check if quorums are enough in the consensus
to ensure certificate-based recovery after the quorum assumption violation.
*)
EXTENDS WaspByzEnv  \* Defines CN, FN, AN, N, F and Q**.
VARIABLE rng        \* From which nodes the RNG messages were received.
VARIABLE done       \* Is the block created and signed?
VARIABLE msgs       \* In-flight messages.
vars == <<rng, done, msgs>>

MsgT == {"nonce", "acs", "rng", "sig"}
Msgs ==
    [ t : MsgT, p : AN ]

TypeOK ==
    /\ rng  \in [AN -> SUBSET AN]
    /\ done \in [AN -> BOOLEAN]
    /\ msgs \in SUBSET Msgs

--------------------------------------------------------------------------------
(* Actions *)

send(m) ==
    /\ m \notin msgs
    /\ msgs' = msgs \cup {m}

ExecNonce(p) ==
    \E q \in QNF :
        /\ \A qp \in q : \E m \in msgs : m.t = "nonce" /\ m.p = qp
        /\ send([t |-> "acs", p |-> p])
        /\ UNCHANGED <<rng, done>>

ExecACS(p) ==
    \E q \in QNF :
        /\ \A qp \in q : \E m \in msgs : m.t = "acs" /\ m.p = qp
        /\ send([t |-> "rng", p |-> p])
        /\ UNCHANGED <<rng, done>>

ExecRNG(p) ==
    \E q \in Q1F :
        /\ \A qp \in q : \E m \in msgs : m.t = "rng" /\ m.p = qp
        /\ rng' = [rng EXCEPT ![p] = q]
        /\ send([t |-> "sig", p |-> p])
        /\ UNCHANGED done

ExecSIG(p) ==
    \E q \in QNF :
        /\ \A qp \in q : \E m \in msgs : m.t = "sig" /\ m.p = qp
        /\ done' = [done EXCEPT ![p] = TRUE]
        /\ UNCHANGED <<rng, msgs>>

--------------------------------------------------------------------------------
(* The specification *)

Init ==
    /\ rng  = [ p \in AN |-> {} ]
    /\ done = [ p \in AN |-> FALSE ]
    /\ msgs = UNION {
            [t : {"nonce"}, p : CN],
            [t : MsgT,      p : FN]
        }

next(p) ==
    \/ ExecNonce(p)
    \/ ExecACS(p)
    \/ ExecRNG(p)
    \/ ExecSIG(p)

Next ==
    \E p \in CN : next(p)

Fair ==
    \A p \in CN : WF_vars(next(p))

Spec == Init /\ [][Next]_vars /\ Fair

--------------------------------------------------------------------------------
(* Properties *)

(* The main property to check:

    If there is any node that might have a TX produced (done), then
    any N-F quorum (at the next log index) will include a fair node
    which will propose the recovery based on certificates and those
    certificates will be from F+1 nodes (from at least 1 correct one).
*)
HaveCertIfDone ==
    (\E p \in AN : done[p]) =>
        \A q \in QNF :                                  \* Amy N-F quorum.
            \E pq \in q :                               \* Will contain
                /\ pq \in CN                            \* A correct node
                /\ \E qq \in Q1F : qq \subseteq rng[pq] \* That proposes F+1 certs.
                /\ \E cn \in CN  : cn \in rng[pq]       \* At least 1 cert from CN.

THEOREM Spec => /\ []TypeOK
                /\ []HaveCertIfDone
PROOF OMITTED \* Checked with the TLC.

================================================================================
