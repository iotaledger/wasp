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
EXTENDS Naturals, FiniteSets, TLAPS, FiniteSetTheorems, NaturalsInduction
CONSTANT Time        \* A set of timestamps, represented as natural numbers to have =<.
CONSTANT Nodes       \* A set of node identifiers.
CONSTANT Byzantine   \* A set of byzantine node identifiers.
ASSUME ConstantAssms ==
  /\ IsFiniteSet(Time) /\ Time # {} /\ Time \subseteq Nat
  /\ IsFiniteSet(Nodes) /\ Nodes # {}
  /\ Byzantine \subseteq Nodes
Requests == Time \* Assume requests are identified by timestamps of their TX only.

VARIABLE acsNodes \* Nodes decided to be part of the round by the ACS.
VARIABLE npRq     \* Node proposal: A set of requests. 
VARIABLE npTS     \* Node proposal: Timestamp.
vars == <<acsNodes, npRq, npTS>>

N == Cardinality(Nodes)
F == CHOOSE F \in 0..N : 
       /\ N >= 3*F+1                           \* Byzantine quorum assumption.
       /\ \A f \in 0..N : N >= 3*f+1 => F >= f \* Consider maximal possible F.
ASSUME ByzantineAssms == F \in Nat /\ N >= 3*F+1 /\ (N >= 4 => F >= 1)

FQuorums  == {q \in SUBSET Nodes : Cardinality(q) = F}
F1Quorums == {q \in SUBSET Nodes : Cardinality(q) = F+1}
NFQuorums == {q \in SUBSET Nodes : Cardinality(q) = N-F}
TSQuorums == {q \in SUBSET Nodes : q \subseteq acsNodes /\ Cardinality(q) = Cardinality(acsNodes) - F}

(*
BatchRqs is a set of requests selected to the batch.
Requests are selected to a batch, if they are mentioned at least in F+1 proposals.
*)
BatchRq(rq) == \E q \in F1Quorums :
                 /\ q \subseteq acsNodes
                 /\ \A n \in q: rq \in npRq[n]
BatchRqs    == {rq \in Requests : BatchRq(rq)}

(*
BatchTS(ts) is a predicate, that is true for the timestamp that should be considered
as a batch timestamp. It must be maximal of the batch proposals, excluding F greatest ones.
*)
SubsetTS(s) == {npTS[n] : n \in s}
BatchTSx(ts) == \A q \in TSQuorums :
                 /\ ts \in SubsetTS(q)
                 /\ \A x \in SubsetTS(q) : ts >= x
                 /\ \A x \in SubsetTS(acsNodes \ q) : ts =< x
BatchTS(ts) ==
  \A q \in FQuorums: (
    /\ q \subseteq acsNodes
    /\ \A x \in q, y \in acsNodes \ q : npTS[x] >= npTS[y]
  ) => (
    /\ ts \in SubsetTS(acsNodes \ q)
    /\ \A x \in SubsetTS(acsNodes \ q) : ts >= x
    /\ \A x \in SubsetTS(q) : ts =< x
  )
               
(*
A batch proposal is valid, if its timestamp is not less than
timestamps of all the request transactions included to the proposal.
*)
ProposalValid(n) == \A rq \in npRq[n] : rq =< npTS[n]
-----------------------------------------------------------------------------
Init ==
  /\ acsNodes \in SUBSET Nodes /\ Cardinality(acsNodes) >= N-F 
  /\ npRq \in [acsNodes -> (SUBSET Requests) \ {{}}]
  /\ npTS \in [acsNodes -> Time]
  /\ \A n \in (acsNodes \ Byzantine) : ProposalValid(n) \* Fair node proposals are valid.
Next == UNCHANGED vars \* Only for model checking in TLC.
Spec == Init /\ [][Next]_vars 

TypeOK ==
  /\ acsNodes \subseteq Nodes
  /\ npRq \in [acsNodes -> SUBSET Requests]
  /\ npTS \in [acsNodes -> Time \cup {0}]

Invariant ==
  \A ts \in Time, rq \in BatchRqs: BatchTS(ts) => rq =< ts

THEOREM Spec => []TypeOK /\ []Invariant
  PROOF OMITTED \* Checked with TLC.
-----------------------------------------------------------------------------
LEMMA SubsetsAllCardinalities ==
  ASSUME NEW S, IsFiniteSet(S)
  PROVE \A x \in 0..Cardinality(S) : \E q \in SUBSET S : Cardinality(q) = x
  <1>1. \A x \in Nat : x =< Cardinality(S) => \E q \in SUBSET S : Cardinality(q) = x
        <2> DEFINE P(x) == x =< Cardinality(S) => \E q \in SUBSET S : Cardinality(q) = x
        <2>1. P(0) BY FS_EmptySet
        <2>2. \A x \in Nat : P(x) => P(x+1)
              <3>1. TAKE x \in Nat
              <3>2. HAVE P(x)
              <3>3. HAVE x + 1 =< Cardinality(S)
              <3>4. PICK qx \in SUBSET S : Cardinality(qx) = x BY <3>2, <3>3, FS_CardinalityType
              <3>5. PICK x1 \in S : x1 \notin qx BY <3>3, <3>4
              <3>6. WITNESS qx \cup {x1} \in SUBSET S
              <3>7. Cardinality(qx \cup {x1}) = x + 1 BY <3>4, <3>5, FS_AddElement, FS_Subset
              <3> QED BY <3>7
        <2>3. QED BY <2>1, <2>2, NatInduction
  <1>2. QED BY <1>1

THEOREM SpecTypeOK == Spec => []TypeOK
  <1>1. Init => TypeOK BY DEF Init, TypeOK
  <1>2. TypeOK /\ [Next]_vars => TypeOK' BY DEF vars, TypeOK, Next
  <1>3. QED BY <1>1, <1>2, PTL DEF Spec

THEOREM SpecInvariant == Byzantine = {} /\ Spec => []Invariant
  <1> SUFFICES ASSUME Byzantine = {} PROVE Spec => []Invariant OBVIOUS
  <1>1. TypeOK /\ Init => Invariant
    <2> SUFFICES ASSUME TypeOK, Init PROVE Invariant OBVIOUS
    <2> USE DEF Invariant
    <2> TAKE ts \in Time, rq \in BatchRqs
    <2> HAVE BatchTS(ts) \* PROVE: rq =< ts
    <2>c. \A q1 \in F1Quorums, q2 \in NFQuorums : q1 \cap q2 # {}
          <3> TAKE q1 \in F1Quorums, q2 \in NFQuorums
          <3>1. N \in Nat /\ F \in Nat
                BY ONLY ConstantAssms, ByzantineAssms, FS_CardinalityType DEF N, F
          <3>2. Cardinality(q1) + Cardinality(q2) > Cardinality(Nodes)
                BY ONLY <3>1 DEF N, F1Quorums, NFQuorums 
          <3>3. q1 \subseteq Nodes /\ q2 \subseteq Nodes
                BY ONLY DEF F1Quorums, NFQuorums
          <3>4. QED BY ONLY <3>2, <3>3, FS_MajoritiesIntersect, ConstantAssms
    <2>b. \A rr \in BatchRqs : \E q \in F1Quorums : \A n \in q: rr \in npRq[n]
          BY DEF BatchRqs, BatchRq
    <2>d. \A nn \in acsNodes : ProposalValid(nn)
          BY DEF Init
    <2>g. acsNodes \subseteq Nodes
          BY DEF Init
    <2>f. Cardinality(acsNodes) - F > 0
          <3>1. Cardinality(acsNodes) \in Nat BY <2>g, FS_CardinalityType, FS_Subset, ConstantAssms
          <3>2. F \in Nat BY ByzantineAssms
          <3>3. N \in Nat BY ConstantAssms, FS_CardinalityType DEF N
          <3>4. Cardinality(acsNodes) >= N-F BY DEF Init
          <3>5. N-F >= 2*F+1 BY ByzantineAssms, <3>2, <3>3
          <3>6. Cardinality(acsNodes) > F BY <3>1, <3>2, <3>3, <3>4, <3>5, ByzantineAssms
          <3> QED BY <3>1, <3>2, <3>6
    <2>j. Cardinality(acsNodes) - F >= 0
          BY <2>f
(*
    <2>i. \E q \in SUBSET Nodes : q \subseteq acsNodes /\ Cardinality(q) = Cardinality(acsNodes) - F
          <3>1. Cardinality(acsNodes) \in Nat BY <2>g, FS_CardinalityType, FS_Subset, ConstantAssms
          <3>6. Cardinality(Nodes) \in Nat BY FS_CardinalityType, ConstantAssms
          <3>2. F \in Nat BY ByzantineAssms
          <3>3. N \in Nat BY ConstantAssms, FS_CardinalityType DEF N
          <3>4. F >= 0 BY <3>2, FS_CardinalityType, FS_EmptySet DEF F
          <3>5. F < N-F BY <3>2, <3>3, FS_CardinalityType, ConstantAssms, ByzantineAssms DEF F
          <3>x. Cardinality(Nodes) >= Cardinality(acsNodes)
                BY FS_CardinalityType, FS_Subset, ConstantAssms DEF Init
          <3>z. Cardinality(acsNodes) >= (Cardinality(acsNodes) - F)
                BY FS_CardinalityType, FS_EmptySet, <2>f, <3>1, <3>4, <3>5, ByzantineAssms
          <3>y. Cardinality(Nodes) >= (Cardinality(acsNodes) - F)
                BY ONLY <3>x, <3>z, <3>1, <3>2, <3>6 
          <3>w1. \E q \in SUBSET Nodes : Cardinality(q) >= (Cardinality(acsNodes) - F)
                 BY <3>y, FS_CardinalityType, FS_Subset, FS_SUBSET
          <3>w2. \E q \in SUBSET Nodes : Cardinality(q) <= (Cardinality(acsNodes) - F)
                 BY <2>j, <3>1, <3>2, FS_CardinalityType, FS_Subset, FS_SUBSET, FS_EmptySet, ConstantAssms
          <3> DEFINE TSQCard == Cardinality(acsNodes) - F
          <3>u1. \E q \in SUBSET Nodes : Cardinality(q) >= 0 /\ Cardinality(q) <= Cardinality(Nodes)
                BY FS_CardinalityType, FS_Subset, FS_SUBSET, ConstantAssms
          <3>u2. TSQCard >= 0 BY <2>j
          <3>u3. TSQCard <= Cardinality(Nodes) BY <3>y
          <3>u4. TSQCard \in Nat BY <3>1, <3>2, <2>f
          <3> HIDE DEF TSQCard
          <3>u5. \A q \in SUBSET Nodes : Cardinality(q) \in Nat BY FS_Subset, FS_CardinalityType, ConstantAssms 
\*          <3>u6. \A x \in 0..Cardinality(Nodes) : \E q \in SUBSET Nodes : Cardinality(q) = x
\*                 BY ConstantAssms, FS_CardinalityType, FS_Subset, FS_SUBSET
          <3>q. \E q \in SUBSET Nodes : Cardinality(q) = TSQCard
                BY <3>u1, <3>u2, <3>u3, <3>u4, <3>u5, <3>6, FS_Subset, FS_CardinalityType, ConstantAssms, SubsetsAllCardinalities 
          <3> PICK q \in SUBSET Nodes : Cardinality(q) = Cardinality(acsNodes) - F
              BY <3>q, FS_CardinalityType, FS_Subset, FS_SUBSET DEF TSQCard
          <3> QED OBVIOUS 
    <2>h. TSQuorums # {} \* TSQuorums == {q \in SUBSET Nodes : q \subseteq acsNodes /\ Cardinality(q) = Cardinality(acsNodes) - F}
          BY <2>f DEF TSQuorums    
*)
    <2> QED OBVIOUS \*PROOF OMITTED \* TODO
  <1>2. Invariant /\ [Next]_vars => Invariant'
    <2>1. SUFFICES ASSUME Invariant PROVE [Next]_vars => Invariant'
          OBVIOUS
    <2>2. UNCHANGED vars => (Invariant')
          BY <2>1 DEF vars, Invariant, BatchRq, BatchRqs, BatchTS,
                      ProposalValid, SubsetTS, TSQuorums
    <2>3. SUFFICES ASSUME Next PROVE Invariant'
          BY <2>2
    <2>4. QED BY <2>1, <2>3 DEF vars, Next, Invariant, BatchRq,
              BatchRqs, BatchTS, ProposalValid, SubsetTS, TSQuorums
  <1>q. QED BY <1>1, <1>2, PTL, SpecTypeOK DEF Spec, vars


(*
THEOREM SpecInvariant == Byzantine = {} /\ Spec => []Invariant
  <1> SUFFICES ASSUME Byzantine = {} PROVE Spec => []Invariant OBVIOUS
  <1>1. Init => Invariant BY DEF Init, Invariant
  <1>2. TypeOK /\ TypeOK' /\ Invariant /\ [Propose]_vars => Invariant'
    <2>1. SUFFICES ASSUME TypeOK, TypeOK', Invariant PROVE [Propose]_vars => Invariant'
          OBVIOUS
    <2>2. UNCHANGED vars => (Invariant')
          BY <2>1 DEF vars, Invariant, BatchRq, BatchRqs, BatchTS,
                      ProposalValid, SubsetTS, TSQuorums
    <2>3. SUFFICES ASSUME Propose PROVE Invariant'
          BY <2>2
    <2>4. SUFFICES ASSUME proposed'
                   PROVE (\A ts \in Time, rq \in BatchRqs: BatchTS(ts) => rq =< ts)'
          BY DEF Invariant
    <2> TAKE ts \in Time, rq \in BatchRqs'
    <2> HAVE BatchTS(ts)' \* PROVE: rq =< ts
\*    <2>a1. \A n \in Nodes \ Byzantine : (\A rqx \in npRq[n] : rqx =< npTS[n])'
\*        BY DEF Propose, ProposalValid
\*    <2>a2. \A n \in Nodes : (\A rqx \in npRq[n] : rqx =< npTS[n])'
\*        BY <2>a1
    <2>b. \A x \in BatchRqs : \E q \in F1Quorums : \A n \in q: x \in npRq[n]
          BY DEF BatchRqs, BatchRq
    <2>c. \A q1 \in F1Quorums, q2 \in NFQuorums : q1 \cap q2 # {}
          <3> TAKE q1 \in F1Quorums, q2 \in NFQuorums
          <3>1. N \in Nat /\ F \in Nat
                BY ONLY ConstantAssms, ByzantineAssms, FS_CardinalityType DEF N, F
          <3>2. Cardinality(q1) + Cardinality(q2) > Cardinality(Nodes)
                BY ONLY <3>1 DEF N, F1Quorums, NFQuorums 
          <3>3. q1 \subseteq Nodes /\ q2 \subseteq Nodes
                BY ONLY DEF F1Quorums, NFQuorums
          <3>4. QED BY ONLY <3>2, <3>3, FS_MajoritiesIntersect, ConstantAssms
    <2>q. QED PROOF OMITTED \* TODO
          \* BY DEF BatchTS, BatchRq, BatchRqs, SubsetTS
          \* ,Requests, NFQuorums, F1Quorums, N, F
  <1>q. QED BY <1>1, <1>2, PTL, SpecTypeOK DEF Spec, vars
*)
=============================================================================
Counter-example with Nodes=101..104, Byzantine={104}, Time=1..3:
  PropposedRq: (101 :> {1} @@ 102 :> {1} @@ 103 :> {2} @@ 104 :> {2}),
  PropposedTS: (101 :> 1   @@ 102 :> 1   @@ 103 :> 2   @@ 104 :> 1  ),
  BatchRq: {1, 2},
  BatchTS: 1
