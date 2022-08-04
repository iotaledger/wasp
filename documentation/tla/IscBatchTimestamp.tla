------------------------- MODULE IscBatchTimestamp -------------------------
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
  This property is modelled bellow as the formula $Invariant$.
\end{enumerate}

The initial attempt was to use the timestamp $t_i.ts$ as a median of
timestamps proposed by the committee nodes (accepted to participate
in the transaction $t_i$ by the ACS procedure). This approach conflicts
with the rules of selecting requests for the batch (take requests that
are mentioned in at least $F+1$ proposals). In this way it is possible
that the median is smaller than some request transaction timestamp  .

\textbf{In this document we model the case}, when we take maximum of the proposed
timestamps excluding $F$ highest values. This value is close to the 66th
percentile (while median is the 50th percentile). In this case all the
requests selected to the batch will have timestamp lower than the
batch timestamp IF THE BATCH PROPOSALS MEET THE CONDITION (modelled bellow
by the formula $ProposalValid$)
$$\forall p \in batchProposals : \forall r \in p.req : p.ts \geq p.req[r].tx.ts.$$

It is possible that this rule can be violated, because of the byzantine
nodes. The specification bellow shows, that property (2) can be violated,
in the case of byzantine node sending timestamp lower than the requests
in the proposal.

The receiving node thus needs to check, if the proposals are correct.
For this check it must have all the request transactions received before deciding
the final batch. The invalid batch proposals cannot be used as is.
Removing them would decrease number of requests included into the final batch
(because requests are included if mentioned in $F+1$ proposals). It is safe
however on the receiver side to "fix" such proposals by setting their timestamps
to the highest transaction timestamp of the requests in the proposal or to adjust
the final batch timestamp to the highest timestamp of the requests selected to it.
In this way the timestamps give no additional means to censor requests
and the batch timestamp cannot be influenced by the adversaries, because
only requests from F+1 nodes are used for such "timestamp fix".

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
ASSUME ByzantineAssms ==
  /\ F \in Nat          \* Implies CHOOSE found a suitable value.
  /\ N >= 3*F+1         \* Standard byzantine Quorum assumption.
  /\ (N >= 4 => F >= 1) \* Just to double-check in TLC.

FQuorums  == {q \in SUBSET Nodes : Cardinality(q) = F}
F1Quorums == {q \in SUBSET Nodes : Cardinality(q) = F+1}
NFQuorums == {q \in SUBSET Nodes : Cardinality(q) = N-F}

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
  /\ npTS \in [acsNodes -> Time]

Invariant ==
  \A ts \in Time, rq \in BatchRqs: BatchTS(ts) => rq =< ts

THEOREM Spec => []TypeOK /\ []Invariant
  PROOF OMITTED \* Checked with TLC, and check the proofs bellow.
-----------------------------------------------------------------------------
LEMMA SubsetsAllCardinalities ==
  ASSUME NEW S, IsFiniteSet(S)
  PROVE \A x \in 0..Cardinality(S) : \E q \in SUBSET S : Cardinality(q) = x
PROOF
<1> DEFINE P(x) == x =< Cardinality(S) => \E q \in SUBSET S : Cardinality(q) = x
<1>1. \A x \in Nat : P(x)
  <2>1. P(0) BY FS_EmptySet
  <2>2. \A x \in Nat : P(x) => P(x+1)
    <3>1. TAKE x \in Nat
    <3>2. HAVE P(x)
    <3>3. HAVE x + 1 =< Cardinality(S)
    <3>4. PICK qx \in SUBSET S : Cardinality(qx) = x
          BY <3>2, <3>3, FS_CardinalityType
    <3>5. PICK x1 \in S : x1 \notin qx
          BY <3>3, <3>4
    <3>6. WITNESS qx \cup {x1} \in SUBSET S
    <3>7. Cardinality(qx \cup {x1}) = x + 1
          BY <3>4, <3>5, FS_AddElement, FS_Subset
    <3> QED BY <3>7
  <2>3. QED BY <2>1, <2>2, NatInduction
<1>2. QED BY <1>1

LEMMA NatSubsetHasMax ==
  ASSUME NEW S, IsFiniteSet(S), S # {}, S \in SUBSET Nat
  PROVE \E n \in S : \A s \in S : s =< n
<1> DEFINE P(x) == x # {} /\ x \subseteq S => \E n \in x : \A s \in x : s =< n
<1> SUFFICES ASSUME TRUE PROVE P(S) OBVIOUS
<1>0. IsFiniteSet(S) OBVIOUS
<1>1. P({}) OBVIOUS
<1>2. ASSUME NEW T, NEW x, IsFiniteSet(T), P(T), x \notin T PROVE P(T \cup {x})
  <2>1. CASE \A t \in T : x >= t
    <3>0. HAVE T \cup {x} # {} /\ T \cup {x} \subseteq S
    <3>1. WITNESS x \in T \cup {x}
    <3> QED BY <2>1, <3>0
  <2>2. CASE ~\A t \in T : x >= t
    <3>4. CASE T = {} \/ ~ T \subseteq S BY <3>4
    <3>5. CASE T # {} /\ T \subseteq S
      <4>1. P(T) BY <1>2
      <4>2. \E n \in T : \A s \in T : s =< n BY <4>1, <3>5
      <4> QED BY <4>2, <3>5, <2>2
    <3> QED BY <3>4, <3>5
  <2>3. QED BY <2>1, <2>2
<1> HIDE DEF P
<1> QED BY ONLY <1>0, <1>1, <1>2, FS_Induction

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
    <2>1. \A q1 \in F1Quorums, q2 \in NFQuorums : q1 \cap q2 # {}
      <3> TAKE q1 \in F1Quorums, q2 \in NFQuorums
      <3>1. N \in Nat /\ F \in Nat BY ONLY ConstantAssms, ByzantineAssms, FS_CardinalityType DEF N, F
      <3>2. Cardinality(q1) + Cardinality(q2) > Cardinality(Nodes) BY ONLY <3>1 DEF N, F1Quorums, NFQuorums 
      <3>3. q1 \subseteq Nodes /\ q2 \subseteq Nodes BY ONLY DEF F1Quorums, NFQuorums
      <3>4. QED BY ONLY <3>2, <3>3, FS_MajoritiesIntersect, ConstantAssms
    <2>2. \A rr \in BatchRqs : \E q \in F1Quorums : \A n \in q: rr \in npRq[n] BY DEF BatchRqs, BatchRq
    <2>3. \A nn \in acsNodes : ProposalValid(nn) BY DEF Init
    <2>4. acsNodes \subseteq Nodes BY DEF Init
    <2>5. Cardinality(acsNodes) - F > 0
      <3>1. Cardinality(acsNodes) \in Nat BY <2>4, FS_CardinalityType, FS_Subset, ConstantAssms
      <3>2. F \in Nat BY ByzantineAssms
      <3>3. N \in Nat BY ConstantAssms, FS_CardinalityType DEF N
      <3>4. Cardinality(acsNodes) >= N-F BY DEF Init
      <3>5. N-F >= 2*F+1 BY ByzantineAssms, <3>2, <3>3
      <3>6. Cardinality(acsNodes) > F BY <3>1, <3>2, <3>3, <3>4, <3>5, ByzantineAssms
      <3> QED BY <3>1, <3>2, <3>6
    <2>6. Cardinality(acsNodes) - F >= 0 BY <2>5
    <2>7. \A fq \in FQuorums, f1q \in F1Quorums : ~ f1q \subseteq fq
      <3>1. TAKE fq \in FQuorums, f1q \in F1Quorums
      <3>2. SUFFICES ASSUME f1q \subseteq fq PROVE FALSE OBVIOUS
      <3>3. IsFiniteSet(f1q) /\ IsFiniteSet(fq) BY ConstantAssms, FS_Subset DEF FQuorums, F1Quorums
      <3>4. Cardinality(f1q) =< Cardinality(fq) BY <3>2, <3>3, FS_Subset
      <3>5. Cardinality(f1q) > Cardinality(fq) BY ByzantineAssms DEF F1Quorums, FQuorums  
      <3>q. QED BY <3>3, <3>4, <3>5, FS_CardinalityType
    <2>8. F \in Nat /\ F >= 0 /\ F <= N /\ F+1 =< N
      <3>1. F \in Nat BY ByzantineAssms
      <3>2. F >= 0 BY <3>1, ConstantAssms DEF F 
      <3>3. N \in Nat BY  ConstantAssms, FS_CardinalityType DEF N
      <3>4. F =< N BY ONLY <3>1, <3>3, ConstantAssms, ByzantineAssms DEF F
      <3>5. F+1 =< N BY ONLY <3>1, <3>3, ConstantAssms, ByzantineAssms DEF F
      <3>q. QED BY ONLY <3>1, <3>2, <3>4, <3>5
    <2>9. FQuorums # {} /\ F1Quorums # {} /\ NFQuorums # {}
           BY <2>8, FS_CardinalityType, ConstantAssms, SubsetsAllCardinalities
           DEF FQuorums, F1Quorums, NFQuorums, N
    <2>10. PICK fq \in FQuorums : fq \subseteq acsNodes /\ \A x \in fq, y \in acsNodes \ fq : npTS[x] >= npTS[y]
      <3>1. SUFFICES \E fq \in FQuorums : fq \subseteq acsNodes /\ \A x \in fq, y \in acsNodes \ fq : npTS[x] >= npTS[y] OBVIOUS
      <3>2. Cardinality(acsNodes) >= N-F BY DEF Init
      <3>3. N-F >= F BY <2>8, ByzantineAssms, ConstantAssms, FS_CardinalityType DEF N
      <3>4. N-F > 0 BY <2>8, ByzantineAssms, ConstantAssms, FS_CardinalityType DEF N
      <3>5. N \in Nat BY FS_CardinalityType, ConstantAssms DEF N
      <3>6. acsNodes \subseteq Nodes BY DEF Init
      <3>7. acsNodes # {} BY ONLY <3>2, <3>4, <3>5, <3>6, <2>8, FS_EmptySet DEF Init
      <3>8. IsFiniteSet(acsNodes) BY FS_Subset, ConstantAssms DEF Init
      <3>9. PICK card \in Nat : card = Cardinality(acsNodes) BY <3>8, FS_CardinalityType
      <3>10. card >= 0 /\ card >= N-F /\ card >= F BY <3>2, <3>3, <2>8, <3>5, <3>9
      <3>11. PICK q \in SUBSET acsNodes : Cardinality(q) = F /\ \A x \in q, y \in acsNodes \ q : npTS[x] >= npTS[y]
        <4> \A q \in SUBSET acsNodes : acsNodes \ q \subseteq Nodes BY DEF Init
        <4> \A q \in SUBSET acsNodes : acsNodes \ q \subseteq acsNodes BY DEF Init
        <4> \A n \in acsNodes : npTS[n] \in Nat BY ConstantAssms DEF TypeOK
        <4> \A c \in 0..card : \E q \in SUBSET acsNodes : Cardinality(q) = c /\ \A x \in q, y \in acsNodes \ q : npTS[x] >= npTS[y]
          <5> DEFINE P(c) == c <= card => \E q \in SUBSET acsNodes : Cardinality(q) = c /\ \A x \in q, y \in acsNodes \ q : npTS[x] >= npTS[y]
          <5>1. SUFFICES ASSUME TRUE PROVE \A c \in Nat : P(c) OBVIOUS
          <5>2. P(0) BY <3>9, FS_EmptySet
          <5>3. \A c \in Nat : P(c) => P(c+1)
            <6>1. TAKE c \in Nat
            <6>2. HAVE P(c)
            <6>3. HAVE c + 1 =< card
            <6>4. PICK q \in SUBSET acsNodes : Cardinality(q) = c /\ (\A x \in q, y \in acsNodes \ q : npTS[x] >= npTS[y]) BY <6>2, <6>3
            <6>5. PICK x \in (acsNodes \ q) : \A xx \in acsNodes \ q : npTS[x] >= npTS[xx]
              <7>1. Cardinality(acsNodes) >= c+1 BY <6>3, <3>9
              <7>2. Cardinality(q) = c BY <6>4
              <7> DEFINE Q == acsNodes \ q
              <7>3. Q # {} BY <7>1, <7>2, FS_Subset
              <7>4. IsFiniteSet(Q) BY <3>8, FS_Subset
              <7>5. Q \in SUBSET acsNodes BY DEF TypeOK
              <7>6. PICK tt \in {npTS[xx] : xx \in Q} : \A ttt \in {npTS[xx] : xx \in Q} : ttt =< tt
                <8> DEFINE QTS == {npTS[xx] : xx \in Q}
                <8> HIDE DEF Q
                <8>1. npTS \in [acsNodes -> Time] BY DEF TypeOK
                <8>2. QTS # {} BY ONLY <7>3, <7>5, <8>1
                <8>3. QTS \in SUBSET Nat BY DEF TypeOK, Q
                <8>4. IsFiniteSet(QTS) BY ONLY <7>4, FS_Image
                <8>5. \E tt \in QTS : \A x \in QTS : tt >= x BY ONLY <8>2, <8>3, <8>4, NatSubsetHasMax
                <8>6. PICK tt \in QTS : \A x \in QTS : tt >= x BY <8>5
                <8>7. WITNESS tt \in QTS
                <8>8. QED BY <8>6
              <7>7. \E nn \in Q : npTS[nn] = tt BY ONLY <7>6, <7>3, TypeOK DEF TypeOK
              <7>8. PICK nn \in Q :  npTS[nn] = tt BY <7>7
              <7>9. WITNESS nn \in Q
              <7> QED BY <7>6, <7>8
            <6>6. q \cup {x} \in SUBSET acsNodes BY <6>4, <6>5
            <6>7. WITNESS q \cup {x} \in SUBSET acsNodes
            <6>8. IsFiniteSet(q) BY <3>8, <6>4, FS_Subset
            <6>9. Cardinality(q \cup {x}) = c + 1 BY FS_AddElement, <6>5, <6>4, <6>8
            <6>10. \A xx \in q \cup {x}, y \in acsNodes \ (q \cup {x}) : npTS[xx] >= npTS[y]
              <7>1. TAKE xx \in q \cup {x}, y \in acsNodes \ (q \cup {x})
              <7>2. CASE xx = x BY <7>2, <6>5
              <7>3. CASE xx \in q BY <7>3, <6>4
              <7>4. QED BY <7>2, <7>3
            <6>11. QED BY <6>9, <6>10
          <5>4. HIDE DEF P
          <5>5. QED BY <5>2, <5>3, NatInduction
        <4> QED BY <3>8, <3>9, <3>10, <2>8, FS_Subset, FS_CardinalityType, SubsetsAllCardinalities
      <3>12. q \in FQuorums /\ \A x \in q, y \in acsNodes \ q : npTS[x] >= npTS[y] BY <3>11, <3>6 DEF FQuorums
      <3>13. q \in FQuorums BY <3>11, <3>6 DEF FQuorums
      <3>14. WITNESS q \in FQuorums
      <3>15. QED BY <3>12, <3>14 
    <2>11. \A x \in BatchRqs : x =< ts
      <3>1. TAKE x \in BatchRqs
      <3>2. x \in Requests /\ BatchRq(x) BY <3>1 DEF BatchRqs
      <3>3. PICK xf1q \in F1Quorums : xf1q \subseteq acsNodes /\ \A n \in xf1q: x \in npRq[n] BY <3>2 DEF BatchRq
      <3>4. xf1q \ fq # {}
        <4>1. Cardinality(xf1q) = F+1 BY <3>3 DEF F1Quorums
        <4>2. Cardinality(fq) = F BY <2>10 DEF FQuorums
        <4>3. F \in Nat BY ByzantineAssms
        <4>4. xf1q \subseteq Nodes /\ fq \subseteq Nodes BY <3>3, <2>10 DEF F1Quorums, FQuorums
        <4>5. IsFiniteSet(xf1q) /\ IsFiniteSet(fq) BY <4>4, ConstantAssms, FS_Subset
        <4>6. QED BY <4>1, <4>2, <4>3, <4>5, FS_Subset
      <3>5. \A n \in (xf1q \ fq) : \A r \in npRq[n] : r =< ts
        <4>1. xf1q \ fq \subseteq acsNodes BY <2>10, <3>3
        <4>2. TAKE xn \in (xf1q \ fq)
        <4>3. TAKE xr \in npRq[xn]
        <4>4. xr \in Nat BY <4>3, <4>1, ConstantAssms DEF TypeOK, Requests
        <4>5. ts \in Nat BY ConstantAssms
        <4>6. npTS[xn] \in Nat BY <4>2, <4>1, ConstantAssms DEF TypeOK
        <4>7. npTS[xn] =< ts
          <5>1. xn \in acsNodes BY <4>2, <4>1
          <5>2. xn \notin fq BY <4>2
          <5>3. /\ ts \in SubsetTS(acsNodes \ fq)
                /\ \A xx \in SubsetTS(acsNodes \ fq) : ts >= xx
                /\ \A xx \in SubsetTS(fq) : ts =< xx
                BY <2>10 DEF BatchTS
          <5>4. QED BY <5>1, <5>2, <5>3 DEF SubsetTS
        <4>8. xr =< npTS[xn]
          <5> ProposalValid(xn) BY <4>1 DEF Init
          <5> QED BY DEF ProposalValid 
        <4>9. QED BY ONLY <4>7, <4>8, <4>4, <4>5, <4>6
      <3>6. \E n \in (xf1q \ fq) : x \in npRq[n] BY <3>4, <3>3
      <3>7. QED BY <3>5, <3>6
    <2>12. QED BY <2>11
  <1>2. Invariant /\ [Next]_vars => Invariant'
    <2>1. SUFFICES ASSUME Invariant PROVE [Next]_vars => Invariant'
          OBVIOUS
    <2>2. UNCHANGED vars => (Invariant')
          BY <2>1 DEF vars, Invariant, BatchRq, BatchRqs, BatchTS,
                      ProposalValid, SubsetTS
    <2>3. SUFFICES ASSUME Next PROVE Invariant'
          BY <2>2
    <2>4. QED BY <2>1, <2>3 DEF vars, Next, Invariant, BatchRq,
              BatchRqs, BatchTS, ProposalValid, SubsetTS
  <1>q. QED BY <1>1, <1>2, PTL, SpecTypeOK DEF Spec, vars

=============================================================================
Counter-example with Nodes=101..104, Byzantine={104}, Time=1..3:
  PropposedRq: (101 :> {1} @@ 102 :> {1} @@ 103 :> {2} @@ 104 :> {2}),
  PropposedTS: (101 :> 1   @@ 102 :> 1   @@ 103 :> 2   @@ 104 :> 1  ),
  BatchRq: {1, 2},
  BatchTS: 1
