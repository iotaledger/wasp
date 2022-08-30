------------------------------- MODULE BrachaRBC -------------------------------
(*
Specification of the Bracha Reliable Broadcast algorithm with a support for
a predicate. Predicates can be updated on the fly, but their sequence has
to be monotonic for each node. I.e. if a predicate was true before, it cannot
become false later.

The original version of this RBC can be found here (see "FIG. 1. The broadcast primitive"):

    Gabriel Bracha. 1987. Asynchronous byzantine agreement protocols. Inf. Comput.
    75, 2 (November 1, 1987), 130–143. DOI:https://doi.org/10.1016/0890-5401(87)90054-X

This specification follows its presentation from (see "Algorithm 2 Bracha’s RBC [14]"):

    Sourav Das, Zhuolun Xiang, and Ling Ren. 2021. Asynchronous Data Dissemination
    and its Applications. In Proceedings of the 2021 ACM SIGSAC Conference on Computer
    and Communications Security (CCS '21). Association for Computing Machinery,
    New York, NY, USA, 2705–2721. DOI:https://doi.org/10.1145/3460120.3484808

The algorithms differ a bit. The latter supports predicates and also it don't
imply sending ECHO messages upon receiving T+1 READY messages. The pseudo-code
from the Das et al.:

     1: // only broadcaster node
     2: input 𝑀
     3: send ⟨PROPOSE, 𝑀⟩ to all
     4: // all nodes
     5: input 𝑃(·) // predicate 𝑃(·) returns true unless otherwise specified.
     6: upon receiving ⟨PROPOSE, 𝑀⟩ from the broadcaster do
     7:     if 𝑃(𝑀) then
     8:         send ⟨ECHO, 𝑀⟩ to all
     9: upon receiving 2𝑡 + 1 ⟨ECHO, 𝑀⟩ messages and not having sent a READY message do
    10:     send ⟨READY, 𝑀⟩ to all
    11: upon receiving 𝑡 + 1 ⟨READY, 𝑀⟩ messages and not having sent a READY message do
    12:     send ⟨READY, 𝑀⟩ to all
    13: upon receiving 2𝑡 + 1 ⟨READY, 𝑀⟩ messages do
    14:     output 𝑀

In the above 𝑡 is "Given a network of 𝑛 nodes, of which up to 𝑡 could be malicious",
thus that's the parameter F in the specification bellow.
*)
EXTENDS FiniteSets, Naturals
CONSTANT CN
CONSTANT FN
CONSTANT Value
ASSUME NodesAssms ==
    /\ IsFiniteSet(CN)
    /\ IsFiniteSet(FN)
    /\ CN \cap FN = {}
    /\ CN # {}
ASSUME ValueAssms == \E v \in Value : TRUE

AN  == CN \cup FN       \* All nodes.
N   == Cardinality(AN)  \* Number of nodes in the system.
F   == Cardinality(FN)  \* Number of faulty nodes.
Q1F == {q \in SUBSET AN : Cardinality(q) = F+1}     \* Contains >= 1 correct node.
Q2F == {q \in SUBSET AN : Cardinality(q) = 2*F+1}   \* Contains >= F+1 correct nodes.
QNF == {q \in SUBSET AN : Cardinality(q) = N-F}                 \* Max quorum.
QXF == {q \in SUBSET AN : Cardinality(q) = ((N+F) \div 2) + 1}  \* Intersection is F+1.
ASSUME QuorumAssms == N > 3*F

VARIABLE bcNode     \* The broadcaster node.
VARIABLE bcValue    \* Value broadcasted by a correct BC node.
VARIABLE predicate  \* Predicates received by the nodes.
VARIABLE output     \* Output for each node.
VARIABLE msgs       \* Messages that were sent.
vars == <<bcNode, bcValue, predicate, output, msgs>>

NotValue == CHOOSE v : v \notin Value
Msg == [t : {"PROPOSE", "ECHO", "READY"}, src: AN, v: Value]
HaveProposeMsg(n, vs) == \E pm \in msgs : pm.t = "PROPOSE" /\ pm.src = n /\ pm.v \in vs
HaveEchoMsg   (n, vs) == \E em \in msgs : em.t = "ECHO"    /\ em.src = n /\ em.v \in vs
HaveReadyMsg  (n, vs) == \E rm \in msgs : rm.t = "READY"   /\ rm.src = n /\ rm.v \in vs

TypeOK ==
    /\ msgs \subseteq Msg
    /\ bcNode \in AN
    /\ \/ bcNode \in CN /\ bcValue \in Value
       \/ bcNode \in FN /\ bcValue = NotValue
    /\ predicate \in [CN -> BOOLEAN ]
    /\ output \in [CN -> Value \cup {NotValue}]

--------------------------------------------------------------------------------
(*
Actions.
*)

(*
>    1: // only broadcaster node
>    2: input 𝑀
>    3: send ⟨PROPOSE, 𝑀⟩ to all
*)
Broadcast ==
    /\ bcNode \in CN \* We only care on the behaviour of the correct nodes.
    /\ ~HaveProposeMsg(bcNode, Value)
    /\ msgs' = msgs \cup {[t |-> "PROPOSE", src |-> bcNode, v |-> bcValue]}
    /\ UNCHANGED <<bcNode, bcValue, predicate, output>>

(*
>    4: // all nodes
>    5: input 𝑃(·) // predicate 𝑃(·) returns true unless otherwise specified.

NOTE: Additionally we allow to update a predicate monotonically.
NOTE: Implementation for the predicate update has been removed from the
      code, just to simplify it. Its use was removed by dropping the adkg/das.
*)
UpdatePredicate ==
    \E n \in CN, p \in BOOLEAN  :
        /\ predicate[n] => p \* Only monotonic updates are make the algorithm to terminate.
        /\ predicate' = [predicate EXCEPT ![n] = p]
        /\ UNCHANGED <<bcNode, bcValue, output, msgs>>

(*
>    6: upon receiving ⟨PROPOSE, 𝑀⟩ from the broadcaster do
>    7:     if 𝑃(𝑀) then
>    8:         send ⟨ECHO, 𝑀⟩ to all
*)
RecvPropose(pm) ==
    \E n \in CN :
        /\ predicate[n]
        /\ HaveProposeMsg(bcNode, {pm.v})
        /\ ~HaveEchoMsg(n, Value)
        /\ msgs' = msgs \cup {[t |-> "ECHO", src |-> n, v |-> pm.v]}
        /\ UNCHANGED <<bcNode, bcValue, predicate, output>>

(*
>    9: upon receiving 2𝑡 + 1 ⟨ECHO, 𝑀⟩ messages and not having sent a READY message do
>   10:     send ⟨READY, 𝑀⟩ to all
*)
RecvEcho(eq) ==
    \E n \in CN, v \in Value :
        /\ eq \in QXF
        /\ \A qn \in eq : HaveEchoMsg(qn, {v})
        /\ ~HaveReadyMsg(n, Value)
        /\ msgs' = msgs \cup {[t |-> "READY", src |-> n, v |-> v]}
        /\ UNCHANGED <<bcNode, bcValue, predicate, output>>

(*
>   11: upon receiving 𝑡 + 1 ⟨READY, 𝑀⟩ messages and not having sent a READY message do
>   12:     send ⟨READY, 𝑀⟩ to all
*)
RecvReadySupport(rq) ==
    \E n \in CN, v \in Value :
        /\ rq \in Q1F
        /\ \A qn \in rq : HaveReadyMsg(qn, {v})
        /\ ~HaveReadyMsg(n, Value)
        /\ msgs' = msgs \cup {[t |-> "READY", src |-> n, v |-> v]}
        /\ UNCHANGED <<bcNode, bcValue, predicate, output>>

(*
>   13: upon receiving 2𝑡 + 1 ⟨READY, 𝑀⟩ messages do
>   14:     output 𝑀
*)
RecvReadyOutput(rq) ==
    \E n \in CN, v \in Value :
        /\ rq \in Q2F
        /\ \A qn \in rq : HaveReadyMsg(qn, {v})
        /\ output[n] = NotValue
        /\ output' = [output EXCEPT ![n] = v]
        /\ UNCHANGED <<bcNode, bcValue, predicate, msgs>>

--------------------------------------------------------------------------------
(*
The specification.
*)

Init ==
    /\ bcNode \in AN
    /\ \/ bcNode \in CN /\ bcValue \in Value
       \/ bcNode \in FN /\ bcValue = NotValue
    /\ predicate \in [CN -> BOOLEAN]
    /\ output = [n \in CN |-> NotValue]
    /\ msgs = [t : {"PROPOSE", "ECHO", "READY"}, src: FN, v: Value]

Next ==
    \/ Broadcast
    \/ UpdatePredicate
    \/ \E pm \in msgs : RecvPropose(pm)
    \/ \E eq \in QXF  : RecvEcho(eq)
    \/ \E rq \in Q1F  : RecvReadySupport(rq)
    \/ \E rq \in Q2F  : RecvReadyOutput(rq)

Fairness ==
    /\ WF_vars(Broadcast)
    /\ WF_vars(UpdatePredicate)
    /\ WF_vars(\E pm \in msgs : pm.src \in CN   /\ RecvPropose(pm))
    /\ WF_vars(\E eq \in QXF  : eq \subseteq CN /\ RecvEcho(eq))
    /\ WF_vars(\E rq \in Q1F  : rq \subseteq CN /\ RecvReadySupport(rq))
    /\ WF_vars(\E rq \in Q2F  : rq \subseteq CN /\ RecvReadyOutput(rq))

Spec == Init /\ [][Next]_vars /\ Fairness

--------------------------------------------------------------------------------
(*
Properties. Their formulations are taken from the Das et al. paper.
*)

HaveOutput(n) == output[n] # NotValue

(*
Agreement. If an honest node outputs a message 𝑀′ and another
honest node outputs a message 𝑀′′, then 𝑀′ = 𝑀′′.
*)
Agreement ==
    \A n1, n2 \in CN :
        HaveOutput(n1) /\ HaveOutput(n2) => output[n1] = output[n2]

(*
Validity. If the broadcaster is honest, all honest nodes eventually
output the message 𝑀.
*)
Validity ==
    bcNode \in CN ~> []\A n \in CN : output[n] = bcValue

(*
Totality. If an honest node outputs a message, then every honest
node eventually outputs a message.
*)
Totality ==
    \A v \in Value :
        (\E n \in CN : output[n] = v) ~> []\A n \in CN : output[n] = v

(*
Additionally: We can only receive a single message of a particular type
from a correct peer. Thus, we can ignore the following messages and
prevent an adversary from sending us a lot of messages to fill our memory.
*)
SingleValueFromPeerPerMsgType ==
    \A m1, m2 \in msgs : (
        /\ m1.src \in CN
        /\ m1.src = m2.src
        /\ m1.t = m2.t
    ) => m1.v = m2.v

THEOREM Spec =>
    /\ []TypeOK
    /\ []Agreement
    /\ []SingleValueFromPeerPerMsgType
    /\ Validity
    /\ Totality
PROOF OMITTED \* Checked by the TLC.

================================================================================
