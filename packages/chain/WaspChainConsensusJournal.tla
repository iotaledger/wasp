--------------------- MODULE WaspChainConsensusJournal -------------------------
(*
This spec is to check, how lagging nodes (by their log indexes) could recover.
The idea here is that:

  - If there are N-F nodes running (CN) with the same LogIndex, they can
    complete the ACS and increase the log index.
  - If a node is lagging, it increases its LogIndex, if sees at least F+1
    messages from other nodes with higher LogIndex.
  - Assume nodes can fail arbitrary, but we always have a N-F quorum of
    correct nodes (maybe with lagging log indexes.)

The TLC says OK.

*)

EXTENDS Naturals, FiniteSets

CONSTANT AN         \* All Nodes
CONSTANT F          \* Max number of faulty nodes.
CONSTANT LogIndex   \* N
ASSUME AllNodesAssms == IsFiniteSet(AN) /\ AN # {}
ASSUME LogIndexAssms == \A li \in LogIndex : 0..li \subseteq LogIndex

VARIABLE CN         \* Correct nodes.
VARIABLE nodes      \* Current Log Indexes for all nodes.
VARIABLE msgs       \* Messages that were sent.
vars == <<CN, nodes, msgs>>

--------------------------------------------------------------------------------
(*
Define quorums.
*)
N == Cardinality(AN)
Q1F == {q \in SUBSET AN : Cardinality(q) = F+1}
QNF == {q \in SUBSET AN : Cardinality(q) = N-F}
ASSUME FAssm == F \in Nat /\ N >= 3*F+1

--------------------------------------------------------------------------------
(*
Types.
*)
Msg == [t: {"LI"}, src: AN, li: LogIndex]

TypeOK ==
    /\ CN    \subseteq AN
    /\ nodes \in [AN -> LogIndex]
    /\ msgs  \subseteq Msg

--------------------------------------------------------------------------------
(*
Actions.
*)

PublishLogIndex ==
    \E n \in CN :
        /\ ~\E m \in msgs : m.t = "LI" /\ m.src = n /\ m.li = nodes[n]
        /\ msgs' = msgs \cup {[t |-> "LI", src |-> n, li |-> nodes[n]]}
        /\ UNCHANGED <<CN, nodes>>

CatchUpLogIndex ==
    \E n \in CN, li \in LogIndex, q \in Q1F :
        /\ li > nodes[n]
        /\ \A qn \in q : \E m \in msgs : m.t = "LI" /\ m.src = qn /\ m.li = li
        /\ nodes' = [nodes EXCEPT ![n] = li]
        /\ UNCHANGED <<CN, msgs>>

AdvanceLogIndex ==
    \E li \in LogIndex, q \in QNF:
        /\ q \subseteq CN
        /\ \A qn \in q : nodes[qn] = li
        /\ li + 1 \in LogIndex
        /\ nodes' = [n \in AN |-> IF n \in q THEN li+1 ELSE nodes[n]]
        /\ UNCHANGED <<CN, msgs>>

FailureChanged ==
    \E q \in SUBSET AN :
        /\ \E qnf \in QNF : qnf \subseteq q \* At least N-F correct nodes.
        /\ CN' = q                          \* Just change the set of correct nodes.
        /\ UNCHANGED <<nodes, msgs>>

--------------------------------------------------------------------------------
(*
The specification.
*)
Init ==
    /\ CN    = AN
    /\ nodes = [n \in AN |-> 0]
    /\ msgs  = {}
Next == PublishLogIndex \/ CatchUpLogIndex \/ AdvanceLogIndex \/ FailureChanged
Fairness == WF_vars(PublishLogIndex \/ CatchUpLogIndex \/ AdvanceLogIndex)
Spec == Init /\ [][Next]_vars /\ Fairness
--------------------------------------------------------------------------------

IsMaxLogIndex(li) == \A lj \in LogIndex : lj <= li

SomeNodesReachMaxLogIndex ==
    <> \E q \in QNF : \A qn \in q : IsMaxLogIndex(nodes[qn])

THEOREM Spec => /\ []TypeOK
                /\ SomeNodesReachMaxLogIndex
PROOF OMITTED \* Checked by the TLC.

================================================================================
