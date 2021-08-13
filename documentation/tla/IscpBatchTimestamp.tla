------------------------- MODULE IscpBatchTimestamp -------------------------
(*`^


Let's assume the transaction currently being built is $t_i$ and the
previous one is $t_{i-1}$. The following requirements apply to the
timestamp $t_i.ts$ of the transaction $t_i$:

\begin{enumerate}
  \item
  Transaction timestamps are non-decreasing function in a chain,
  i.e. $$t_i.ts \geq t_{i-1}.ts.$$
  \item
  A transaction timestamp is not smaller than the timestamps
  of request transactions taken as inputs in $t_i$, i.e.
  $$\forall r \in t_i.req: t_i.ts \geq t_i.req[r].tx.ts,$$
  where $t_i.req$ is a list of requests processed as inputs in
  the transaction $t_i$, $t_i.req[r]$ is a particular request
  and $t_i.req[r].tx$ is a transaction the request belongs to.
\end{enumerate}

The initial attempt was to use the timestamp $t_i.ts$ as a median of
timestamps proposed by the committee nodes accepted to participate
in the transaction $t_i$ by the ACS procedure. This approach conflicts
with the rules of selecting requests for the batch (take requests that
are mentioned in at least $F+1$ proposals). In this way it is possible
that the median is smaller than some request transaction timestamp  .

\textbf{In this document we model the case}, when we take maximal of the proposed
timestamps excluding the $F$ highest values. This value is close to the 66th
percentile (while median is the 50th percentile). In this case all the
requests selected to the batch will have timestamp lower than the
batch timestamp IF THE BATCH PROPOSALS MEET THE CONDITION
$$\forall p \in batchProposals : \forall r \in p.req : p.req[r].tx.ts \leq p.ts.$$

It is possible that it can be not the case, because of the byzantine
nodes. The specification bellow shows, that property (2) can be violated,
in the case of byzantine node sending timestamp lower than the requests
in the proposal.

The receiving node thus needs to check, if the proposals are correct.
For this check it must have all the transactions received before deciding
the final batch. The detected invalid batch proposals must be excluded
from the following procedure. But that can decrease number of requests
included into the final batch (because requests are included if mentioned
in $F+1$ proposals). It is safe on the receiver side to "fix" such proposals
by setting their timestamp to the maximal transaction timestamp of the
requests in the proposal. 
^'*)
EXTENDS Naturals, FiniteSets, TLAPS
CONSTANT Nodes       \* A set of node identifiers.
CONSTANT Byzantine   \* A set of byzantine node identifiers.
CONSTANT Time        \* A set of timestamps, represented as natural numbers to have =<.
ASSUME ConstantAssms == Byzantine \subseteq Nodes /\ Time \subseteq Nat /\ Time # {} /\ Nodes # {}
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
Init == \* Dummy values, on init (to make TLC faster), see Propose instead. 
  /\ proposed = FALSE
  /\ npRq = [n \in Nodes |-> {}]
  /\ npTS = [n \in Nodes |-> 0]
Spec == Init /\ [][Propose]_vars \* For model checking in TLC.

TypeOK ==
  /\ proposed \in BOOLEAN
  /\ npRq \in [Nodes -> SUBSET Requests]
  /\ npTS \in [Nodes -> Time \cup {0}]

Invariant ==
  proposed => \A ts \in Time, rq \in BatchRqs: BatchTS(ts) => rq =< ts

THEOREM Spec => []TypeOK /\ []Invariant
  PROOF OMITTED \* Checked with TLC.
-----------------------------------------------------------------------------
THEOREM SpecTypeOK == Spec => []TypeOK
  <1> Init => TypeOK BY DEF Init, TypeOK
  <1> TypeOK /\ [Propose]_vars => TypeOK'
      <2> SUFFICES ASSUME TypeOK, Propose PROVE TypeOK' BY DEF vars, TypeOK
      <2> QED BY DEF TypeOK, Propose
  <1> QED BY PTL DEF Spec

THEOREM Byzantine = {} /\ Spec => []Invariant
  <1> SUFFICES ASSUME Byzantine = {} PROVE Spec => []Invariant OBVIOUS
  <1>1. Init => Invariant BY DEF Init, Invariant
  <1>2. TypeOK /\ TypeOK' /\ Invariant /\ [Propose]_vars => Invariant'
    <2> SUFFICES ASSUME TypeOK, TypeOK', Invariant, Propose PROVE Invariant'
        BY DEF vars, Invariant, BatchRq, BatchRqs, BatchTS, ProposalValid, SubsetTS
    <2> SUFFICES ASSUME proposed'
                 PROVE (\A ts \in Time, rq \in BatchRqs: BatchTS(ts) => rq =< ts)'
        BY DEF Invariant
    <2> TAKE ts \in Time, rq \in BatchRqs'
    <2> HAVE BatchTS(ts)'
    <2> \A n \in Nodes \ Byzantine : (\A rqx \in npRq[n] : rqx =< npTS[n])'
        BY DEF Propose, ProposalValid
    <2> \A n \in Nodes : (\A rqx \in npRq[n] : rqx =< npTS[n])'
        OBVIOUS
    <2>q. QED PROOF OMITTED \* TODO
          \* BY DEF BatchTS, BatchRq, BatchRqs, SubsetTS
          \* ,Requests, NFQuorums, F1Quorums, N, F
  <1>q. QED BY <1>1, <1>2, PTL, SpecTypeOK DEF Spec, vars
=============================================================================
Counter-example with Nodes=101..104, Byzantine={104}, Time=1..3:
  PropposedRq: (101 :> {1} @@ 102 :> {1} @@ 103 :> {2} @@ 104 :> {2}),
  PropposedTS: (101 :> 1   @@ 102 :> 1   @@ 103 :> 2   @@ 104 :> 1  ),
  BatchRq: {1, 2},
  BatchTS: 1
