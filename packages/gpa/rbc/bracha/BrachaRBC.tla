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

The algorithms differs a bit. The latter supports predicates and also it don't
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

AN == CN \cup FN
N == Cardinality(AN)
F == Cardinality(FN)
ASSUME QuorumAssms == N > 3*F
Q1F == {q \in SUBSET AN : Cardinality(q) = F+1}
Q2F == {q \in SUBSET AN : Cardinality(q) = 2*F+1}

VARIABLE bcNode     \* The broadcaster node.
VARIABLE bcValue    \* Value broadcasted by a correct BC node.
VARIABLE predicate  \* Predicates received by the nodes.
VARIABLE output     \* Output for each node.
VARIABLE msgs       \* Messages that were sent.
vars == <<bcNode, bcValue, predicate, output, msgs>>

NotValue == CHOOSE v : v \notin Value
Msg == [t : {"INITIAL", "ECHO", "READY"}, src: AN, v: Value]
HaveInitialMsg(n, vs) == \E im \in msgs : im.t = "INITIAL" /\ im.src = n /\ im.v \in vs
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
Broadcast ==
    /\ bcNode \in CN \* We only care about the correct nodes.
    /\ ~HaveInitialMsg(bcNode, Value)
    /\ msgs' = msgs \cup {[t |-> "INITIAL", src |-> bcNode, v |-> bcValue]}
    /\ UNCHANGED <<bcNode, bcValue, predicate, output>>

RecvInitial(im) ==
    \E n \in CN :
        /\ predicate[n]
        /\ HaveInitialMsg(bcNode, {im.v})
        /\ ~HaveEchoMsg(n, Value)
        /\ msgs' = msgs \cup {[t |-> "ECHO", src |-> n, v |-> im.v]}
        /\ UNCHANGED <<bcNode, bcValue, predicate, output>>

RecvEcho(eq) ==
    \E n \in CN, v \in Value :
        /\ eq \in Q2F
        /\ \A qn \in eq : HaveEchoMsg(qn, {v})
        /\ ~HaveReadyMsg(n, Value)
        /\ msgs' = msgs \cup {[t |-> "READY", src |-> n, v |-> v]}
        /\ UNCHANGED <<bcNode, bcValue, predicate, output>>

RecvReadySupport(rq) ==
    \E n \in CN, v \in Value :
        /\ rq \in Q1F
        /\ \A qn \in rq : HaveReadyMsg(qn, {v})
        /\ ~HaveReadyMsg(n, Value)
        /\ msgs' = msgs \cup {[t |-> "READY", src |-> n, v |-> v]}
        /\ UNCHANGED <<bcNode, bcValue, predicate, output>>

RecvReadyOutput(rq) ==
    \E n \in CN, v \in Value :
        /\ rq \in Q2F
        /\ \A qn \in rq : HaveReadyMsg(qn, {v})
        /\ output[n] = NotValue
        /\ output' = [output EXCEPT ![n] = v]
        /\ UNCHANGED <<bcNode, bcValue, predicate, msgs>>

UpdatePredicate ==
    \E n \in CN, p \in BOOLEAN  :
        /\ predicate[n] => p \* Only monotonic updates are make the algorithm to terminate.
        /\ predicate' = [predicate EXCEPT ![n] = p]
        /\ UNCHANGED <<bcNode, bcValue, output, msgs>>

--------------------------------------------------------------------------------
Init ==
    /\ bcNode \in AN
    /\ \/ bcNode \in CN /\ bcValue \in Value
       \/ bcNode \in FN /\ bcValue = NotValue
    /\ predicate \in [CN -> BOOLEAN]
    /\ output = [n \in CN |-> NotValue]
    /\ msgs = [t : {"INITIAL", "ECHO", "READY"}, src: FN, v: Value]

Next ==
    \/ Broadcast
    \/ \E im \in msgs : RecvInitial(im)
    \/ \E eq \in Q2F  : RecvEcho(eq)
    \/ \E rq \in Q1F  : RecvReadySupport(rq)
    \/ \E rq \in Q2F  : RecvReadyOutput(rq)
    \/ UpdatePredicate

Fairness ==
    /\ WF_vars(Broadcast)
    /\ WF_vars(\E im \in msgs : im.src \in CN   /\ RecvInitial(im))
    /\ WF_vars(\E eq \in Q2F  : eq \subseteq CN /\ RecvEcho(eq))
    /\ WF_vars(\E rq \in Q1F  : rq \subseteq CN /\ RecvReadySupport(rq))
    /\ WF_vars(\E rq \in Q2F  : rq \subseteq CN /\ RecvReadyOutput(rq))
    /\ WF_vars(UpdatePredicate)

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

THEOREM Spec =>
    /\ []TypeOK
    /\ []Agreement
    /\ Validity
    /\ Totality
PROOF OMITTED \* Checked by the TLC.

================================================================================