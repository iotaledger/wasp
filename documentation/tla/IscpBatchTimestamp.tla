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
EXTENDS Naturals, FiniteSets, TLAPS, FiniteSetTheorems, NaturalsInduction, FunctionTheorems
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
BatchTSx(ts) == \A q \in TSQuorums : \* TODO: Remove
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
  /\ npTS \in [acsNodes -> Time]

Invariant ==
  \A ts \in Time, rq \in BatchRqs: BatchTS(ts) => rq =< ts

THEOREM Spec => []TypeOK /\ []Invariant
  PROOF OMITTED \* Checked with TLC.
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
    <2>k. \A fq \in FQuorums, f1q \in F1Quorums : ~ f1q \subseteq fq
          <3>1. TAKE fq \in FQuorums, f1q \in F1Quorums
          <3>2. SUFFICES ASSUME f1q \subseteq fq PROVE FALSE OBVIOUS
          <3>3. IsFiniteSet(f1q) /\ IsFiniteSet(fq) BY ConstantAssms, FS_Subset DEF FQuorums, F1Quorums
          <3>4. Cardinality(f1q) =< Cardinality(fq) BY <3>2, <3>3, FS_Subset
          <3>5. Cardinality(f1q) > Cardinality(fq) BY ByzantineAssms DEF F1Quorums, FQuorums  
          <3>q. QED BY <3>3, <3>4, <3>5, FS_CardinalityType
\*    <2> DEFINE InBiggestF == \A q \in FQuorums : q \subseteq acsNodes /\ \A x \in q, y \in acsNodes \ q : npTs[x] >
\*    <2>1. rq > ts => \A q \in FQuorums : q \subseteq acsNodes /\ \A x \in q, y \in acsNodes \ q : npTS[x] >= npTS[x]
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
    <2>10. F \in Nat /\ F >= 0 /\ F <= N /\ F+1 =< N
           <3>1. F \in Nat BY ByzantineAssms
           <3>2. F >= 0 BY <3>1, ConstantAssms DEF F 
           <3>3. N \in Nat BY  ConstantAssms, FS_CardinalityType DEF N
           <3>4. F =< N BY ONLY <3>1, <3>3, ConstantAssms, ByzantineAssms DEF F
           <3>5. F+1 =< N BY ONLY <3>1, <3>3, ConstantAssms, ByzantineAssms DEF F
           <3>q. QED BY ONLY <3>1, <3>2, <3>4, <3>5
    <2>11. FQuorums # {} /\ F1Quorums # {} /\ NFQuorums # {}
           BY <2>10, FS_CardinalityType, ConstantAssms, SubsetsAllCardinalities
           DEF FQuorums, F1Quorums, NFQuorums, N
    <2>12. PICK fq \in FQuorums : fq \subseteq acsNodes /\ \A x \in fq, y \in acsNodes \ fq : npTS[x] >= npTS[y]
           <3>1. FQuorums # {} BY <2>11
           <3>3. \E fq \in FQuorums : fq \subseteq acsNodes /\ \A x \in fq, y \in acsNodes \ fq : npTS[x] >= npTS[y]
                 <4>1. Cardinality(acsNodes) >= N-F BY DEF Init
                 <4>2. \A fq \in FQuorums : Cardinality(fq) = F BY DEF FQuorums
                 <4>3. N-F >= F BY <2>10, ByzantineAssms, ConstantAssms, FS_CardinalityType DEF N
                 <4>4. N-F > 0 BY <2>10, ByzantineAssms, ConstantAssms, FS_CardinalityType DEF N
                 <4>5. N \in Nat BY FS_CardinalityType, ConstantAssms DEF N
                 <4>6. \A fq \in FQuorums: fq \subseteq Nodes BY DEF FQuorums
                 <4>7. acsNodes \subseteq Nodes BY DEF Init
                 <4>8. acsNodes # {} BY ONLY <4>1, <4>4, <4>5, <4>7, <2>10, FS_EmptySet DEF Init
                 <4>9. IsFiniteSet(acsNodes) BY FS_Subset, ConstantAssms DEF Init
                 <4>10. PICK card \in Nat : card = Cardinality(acsNodes) BY <4>9, FS_CardinalityType
                 <4>11. card >= 0 /\ card >= N-F /\ card >= F BY <4>1, <4>3, <2>10, <4>5, <4>10
                 <4>12. PICK q \in SUBSET acsNodes : Cardinality(q) = F /\ \A x \in q, y \in acsNodes \ q : npTS[x] >= npTS[y]
                     <5> \A q \in SUBSET acsNodes : acsNodes \ q \subseteq Nodes BY DEF Init
                     <5> \A q \in SUBSET acsNodes : acsNodes \ q \subseteq acsNodes BY DEF Init
                     <5> \A n \in acsNodes : npTS[n] \in Nat BY ConstantAssms DEF TypeOK
                     <5> \A c \in 0..card : \E q \in SUBSET acsNodes : Cardinality(q) = c /\ \A x \in q, y \in acsNodes \ q : npTS[x] >= npTS[y]
                         <6>1. DEFINE P(c) == c <= card => \E q \in SUBSET acsNodes : Cardinality(q) = c /\ \A x \in q, y \in acsNodes \ q : npTS[x] >= npTS[y]
                         <6>2. SUFFICES ASSUME TRUE PROVE \A c \in Nat : P(c) OBVIOUS
                         <6>3. P(0) BY <4>10, FS_EmptySet
                         <6>4. \A c \in Nat : P(c) => P(c+1)
                               <7>1. TAKE c \in Nat
                               <7>2. HAVE P(c)          \* PROVE P(c+1)
                               <7>3. HAVE c + 1 =< card \* PROVE \E q \in SUBSET acsNodes : Cardinality(q) = c + 1 /\ (\A x \in q, y \in acsNodes \ q : npTS[x] >= npTS[y])
                               <7>4. PICK q \in SUBSET acsNodes : Cardinality(q) = c /\ (\A x \in q, y \in acsNodes \ q : npTS[x] >= npTS[y])
                                     BY <7>2, <7>3
                               <7>6. PICK x \in (acsNodes \ q) : \A xx \in acsNodes \ q : npTS[x] >= npTS[xx]
                                     <8>1. Cardinality(acsNodes) >= c+1 BY <7>3, <4>10
                                     <8>2. Cardinality(q) = c BY <7>4
                                     <8> DEFINE Q == acsNodes \ q
                                     <8>3. Q # {} BY <8>1, <8>2, FS_Subset
                                     <8>4. IsFiniteSet(Q) BY <4>9, FS_Subset
                                     <8>4b. Q \in SUBSET acsNodes BY DEF TypeOK
                                     <8>4c. npTS \in [acsNodes -> Time] BY DEF TypeOK
                                     <8>5. \A xx \in Q : npTS[xx] \in Nat OBVIOUS
                                     <8>5a. \A xx \in Q : npTS[xx] \in Time BY DEF TypeOK
                                     <8>6. IsFiniteSet({npTS[xx] : xx \in Q}) BY ONLY <8>4, FS_Image
                                     <8>7. {npTS[xx] : xx \in Q} \subseteq Nat BY DEF TypeOK
                                     <8>8. {npTS[xx] : xx \in Q} \subseteq Time BY DEF TypeOK
                                     <8>8b. {npTS[xx] : xx \in Q} # {}
                                            BY ONLY <8>3, <8>4b, <8>4c
                                     <8>9. \E nn \in Nat : \A t \in Time : t <= nn
                                            BY <8>3, ConstantAssms, NatSubsetHasMax
                                     <8>10a. \E nn \in Nat : \A xxts \in {npTS[xx] : xx \in Q} : xxts =< nn
                                            BY ONLY <8>8, <8>9, ConstantAssms, NatSubsetHasMax
                                     <8>10b. \E nn \in Time : \A xxts \in {npTS[xx] : xx \in Q} : xxts =< nn
                                            BY ONLY <8>8, <8>9, ConstantAssms, NatSubsetHasMax
                                     <8>11. \E x \in Q : TRUE
                                         BY ONLY <8>3
                                     <8>12. IsFiniteSet(Time) /\ Time \subseteq Nat BY ConstantAssms
                                     <8>21. \A n \in acsNodes : \E t \in Time : npTS[n] = t BY DEF TypeOK
                                     <8>22b. PICK tt \in {npTS[xx] : xx \in Q} : \A ttt \in {npTS[xx] : xx \in Q} : ttt =< tt
                                       <9> DEFINE QTS == {npTS[xx] : xx \in Q}
                                       <9> HIDE DEF Q
                                       <9>1. QTS # {} BY <8>8b
                                       <9>2. QTS \in SUBSET Nat BY <8>7
                                       <9>3. IsFiniteSet(QTS) BY <8>6
                                       <9>4. \E tt \in QTS : \A x \in QTS : tt >= x BY ONLY <9>1, <9>2, <9>3, NatSubsetHasMax
                                       <9> PICK tt \in QTS : \A x \in QTS : tt >= x BY <9>4
                                       <9> WITNESS tt \in QTS
                                       <9> QED OBVIOUS
                                     <8>23. \E nn \in Q : npTS[nn] = tt BY ONLY <8>22b, <8>8, <8>3, TypeOK DEF TypeOK
                                     <8>24. PICK nn \in Q :  npTS[nn] = tt BY <8>23
                                     <8>25. WITNESS nn \in Q
                                     <8> QED BY <8>22b, <8>24
                               <7>7. q \cup {x} \in SUBSET acsNodes BY <7>4, <7>6
                               <7>8. WITNESS q \cup {x} \in SUBSET acsNodes
                               <7>9a. IsFiniteSet(q) BY <4>9, <7>4, FS_Subset
                               <7>9. Cardinality(q \cup {x}) = c + 1 BY FS_AddElement, <7>6, <7>4, <7>9a
                               <7>10. \A xx \in q \cup {x}, y \in acsNodes \ (q \cup {x}) : npTS[xx] >= npTS[y]
                                      <8> TAKE xx \in q \cup {x}, y \in acsNodes \ (q \cup {x})
                                      <8>1. CASE xx = x BY <8>1, <7>6
                                      <8>2. CASE xx \in q BY <8>2, <7>4
                                      <8>3. QED BY <8>1, <8>2
                               <7>q. QED BY <7>9, <7>10
                         <6>5. HIDE DEF P
                         <6>6. QED BY <6>3, <6>4, NatInduction
                     <5> QED BY <4>9, <4>10, <4>11, <2>10, FS_Subset, FS_CardinalityType, SubsetsAllCardinalities
                 <4>13. q \in FQuorums /\ \A x \in q, y \in acsNodes \ q : npTS[x] >= npTS[y] BY <4>12, <4>7 DEF FQuorums
                 <4>14. q \in FQuorums BY <4>12, <4>7 DEF FQuorums
                 <4>15. WITNESS q \in FQuorums
                 <4> QED BY <4>15, <4>13
           <3>4. QED BY <3>3
    <2>5. PICK fr \in Requests : ~\E qq \in F1Quorums : qq \subseteq acsNodes /\ \A nn \in qq : fr \in npRq[nn]
    <2>2. \A n \in fq : ts =< npTS[n] BY <2>12 DEF BatchTS, SubsetTS
    <2>3. \A n \in fq : \A nr \in npRq[n] : nr \in BatchRqs => \E nn \in acsNodes \ fq : nr \in npRq[n]
    <2>4. \A n \in acsNodes \ fq : \A r \in npRq[n] : r =< ts
    <2> QED BY <2>12, <2>2, <2>3, <2>4 DEF BatchRq, BatchRqs, FQuorums, F1Quorums, BatchTS, SubsetTS
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
