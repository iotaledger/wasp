------------------------- MODULE IscpBatchTimestamp -------------------------
(*`^

asd

$$ \forall i : a_i = b$$


Can we determine if proposal is invalid with regards to the timestamp?
That is done after the ACS.


^'*)
EXTENDS Naturals, FiniteSets, TLAPS
CONSTANT Nodes       \* A set of node identifiers.
CONSTANT Byzantine   \* A set of byzantine node identifiers.
CONSTANT Time        \* A set of timestamps, represented as natural numbers to have =<.
ASSUME ConstantAssms == Byzantine \subseteq Nodes /\ Time \subseteq Nat
Requests == Time \* Assume requests are identified by timestamps of their TX only.

VARIABLE proposed \* Was the proposal made?
VARIABLE npRq     \* Node proposal: A set of requests. 
VARIABLE npTS     \* Node proposal: Timestamp.
vars == <<proposed, npRq, npTS>>

F == Cardinality(Byzantine)
N == Cardinality(Nodes)
ASSUME ByzantineAssm == N >= 3*F+1

F1Quorums == {q \in SUBSET Nodes : Cardinality(q) = F+1}
NFQuorums == {q \in SUBSET Nodes : Cardinality(q) = N-F}

BatchRq(rq) == \E q \in F1Quorums : \A n \in q: rq \in npRq[n]
BatchRqs    == {rq \in Requests : BatchRq(rq)}

SubsetTS(s) == {npTS[n] : n \in s}
BatchTS(ts) == \E q \in NFQuorums :
                 /\ ts \in SubsetTS(q)
                 /\ \A x \in SubsetTS(q) : ts >= x
                 /\ \A x \in SubsetTS(Nodes \ q) : ts =< x

ProposalValid(n) == \A rq \in npRq[n] : rq =< npTS[n]
Propose == ~proposed /\ proposed' = TRUE
  /\ npRq' \in [Nodes -> (SUBSET Requests) \ {{}}]    \* Some node non-empty proposals.
  /\ npTS' \in [Nodes -> Time]                        \* Some timestamps.
  /\ \A n \in (Nodes \ Byzantine) : ProposalValid(n)' \* Fair node proposals are valid.
-----------------------------------------------------------------------------
Init == \* Dummy values, on init.
  /\ proposed = FALSE
  /\ npRq = [n \in Nodes |-> {}]
  /\ npTS = [n \in Nodes |-> 0]
Spec == Init /\ [][Propose]_vars                   \* For model checking in TLC.
TypeOK ==
  /\ proposed \in BOOLEAN
  /\ npRq \in [Nodes -> SUBSET Requests]
  /\ npTS \in [Nodes -> Time \cup {0}]
Invariant ==
  proposed => \A ts \in Time, rq \in BatchRqs: BatchTS(ts) => rq =< ts

THEOREM Spec => []TypeOK /\ []Invariant
  PROOF OMITTED \* Checked with TLC.

THEOREM Byzantine = {} /\ Spec => []Invariant
  <1> SUFFICES ASSUME Byzantine = {} PROVE Spec => []Invariant OBVIOUS
  <1>1. Init => Invariant BY DEF Init, Invariant
  <1>2. Invariant /\ [Propose]_vars => Invariant'
    <2> SUFFICES ASSUME Invariant, Propose PROVE Invariant'
        BY DEF Propose, Invariant, vars, BatchRqs, BatchTS, ProposalValid, SubsetTS, NFQuorums, F1Quorums
    <2>q. QED
  <1>q. QED BY <1>1, <1>2, PTL DEF Spec, vars
=============================================================================
Counter-example with Nodes=101..104, Byzantine={104}, Time=1..3:
  PropposedRq: (101 :> {1} @@ 102 :> {1} @@ 103 :> {2} @@ 104 :> {2}),
  PropposedTS: (101 :> 1   @@ 102 :> 1   @@ 103 :> 2   @@ 104 :> 1  ),
  BatchRq: {1, 2},
  BatchTS: 1
