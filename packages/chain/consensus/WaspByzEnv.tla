------------------------------ MODULE WaspByzEnv -------------------------------
(*
Commonly used node and quorum definitions.
Should be used in other modules by extending this one.
*)
EXTENDS FiniteSets, Naturals
CONSTANT CN
CONSTANT FN
ASSUME NodesAssms ==
    /\ IsFiniteSet(CN)
    /\ IsFiniteSet(FN)
    /\ CN \cap FN = {}
    /\ CN # {}

AN  == CN \cup FN       \* All nodes.
N   == Cardinality(AN)  \* Number of nodes in the system.
F   == Cardinality(FN)  \* Number of faulty nodes.
Q1F == {q \in SUBSET AN : Cardinality(q) = F+1}     \* Contains >= 1 correct node.
Q2F == {q \in SUBSET AN : Cardinality(q) = 2*F+1}   \* Contains >= F+1 correct nodes.
QNF == {q \in SUBSET AN : Cardinality(q) = N-F}                 \* Max quorum.
QXF == {q \in SUBSET AN : Cardinality(q) = ((N+F) \div 2) + 1}  \* Intersection is F+1.
ASSUME QuorumAssms == N > 3*F

================================================================================
