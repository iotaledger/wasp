---------------------------- MODULE WaspACSS -----------------------------------
EXTENDS Naturals, FiniteSets, TLC, TLAPS
CONSTANT Nodes
CONSTANT Dealer
CONSTANT Faulty
ASSUME NodesAssm == IsFiniteSet(Nodes) /\ \E n \in Nodes : TRUE
ASSUME DealerAssm == Dealer \in Nodes
VARIABLE msgs
vars == <<msgs>>
--------------------------------------------------------------------------------
(*                               QUORUMS                                      *)
N == Cardinality(Nodes)
F == CHOOSE F \in 0..N :
              /\ 3*F+1 <= N                              \* F rule.
              /\ \A FF \in 0..N : 3*FF+1 <= N => FF <= F \* Max.
QNF == {q \in SUBSET Nodes : Cardinality(q) = N-F}
QF1 == {q \in SUBSET Nodes : Cardinality(q) = F+1}
Fair == Nodes \ Faulty
ASSUME FaultyAssm == Faulty \subseteq Nodes /\ Cardinality(Faulty) <= F
--------------------------------------------------------------------------------
(*                                 TYPES                                      *)
(*
Deal contains secret shares for all the nodes, but hidden in such way, that
only the corresponding node can decrypt its secret share. We con't model the
crypto part here, so a Deal is modelled by a function [Nodes -> BOOLEAN],
which says for each node, if its deal is correct.
*)
Deal == [Nodes -> BOOLEAN]
(*
A set of possible messages.

In all the message types:
  - t -- a type of the message.
  - src -- is a sender of a node.
  - rcp -- is a recipient of a node.
           If this field is omitted, a message is sent to all the nodes.

The following message types are used in this spec:
  - "DEAL_BC" -- a message sent by the dealer to the RBC to share its Deal.
  - "DEAL"    -- a deal received by a node from the Dealer via the RBC.
  - "OK", "READY" -- Bracha style synchronization.
  - "OUTPUT"  -- models the final decision by the node 'src".
  - "IMPLICATE", "RECOVER" -- handle the scenario with a dealer being faulty.
    The field "sec" here stands for the correctness of the "src"s node secret.
*)
Msgs == UNION {
    [t: {"DEAL_BC"},          deal: Deal],
    [t: {"DEAL"}, rcp: Nodes, deal: Deal],
    [t: {"OK", "READY", "OUTPUT"}, src: Nodes],
    [t: {"IMPLICATE", "RECOVER"}, src: Nodes, sec: BOOLEAN ]
}
TypeOK ==
    msgs \in SUBSET Msgs

rbc == INSTANCE RBC WITH
         MsgP <- LAMBDA m : m.t = "DEAL_BC",
         MsgS <- LAMBDA m, n : [t |-> "DEAL", rcp |-> n, deal |-> m.deal]

--------------------------------------------------------------------------------
(*                                ACTIONS                                     *)
\* >
\* > // dealer with input s
\* > sample random polynomial ϕ such that ϕ(0) = s
\* > C, S := VSS.Share(ϕ, f+1, n)
\* > E := [PKI.Enc(S[i], pkᵢ) for each party i]
\* >
Input ==
    \/ /\ Dealer \in Faulty
       /\ \E deal \in [Nodes -> BOOLEAN]:
            msgs' = msgs \cup {[t |-> "DEAL_BC", deal |-> deal]}
    \/ /\ Dealer \notin Faulty
       /\ msgs' = msgs \cup {[t |-> "DEAL_BC", deal |-> [n \in Nodes |-> TRUE]]}

\* >
\* > RBC(C||E)
\* >
RBC ==
    /\ rbc!Broadcast

\* >
\* > sᵢ := PKI.Dec(eᵢ, skᵢ)
\* > if decrypt fails or VSS.Verify(C, i, sᵢ) == false:
\* >   send <IMPLICATE, i, skᵢ> to all parties
\* > else:
\* >   send <OK>
\* >
HandleDeal(n) ==
    \E m \in msgs: m.t = "DEAL" /\ m.rcp = n /\
        \/ m.deal[n]  /\ msgs' = msgs \cup {[t |-> "OK",        src |-> n]}
        \/ ~m.deal[n] /\ msgs' = msgs \cup {[t |-> "IMPLICATE", src |-> n, sec |-> TRUE]}
        \* False IMPLICATE messages can only be produced by the Faulty nodes.
        \* Those messages were added in the Init state already.

\* >
\* > on receiving <OK> from n-f parties:
\* >   send <READY> to all parties
\* >
HandleOK(n) ==
    /\ \E q \in QNF: \A qn \in q: \E m \in msgs : m.t = "OK" /\ m.src = qn
    /\ msgs' = msgs \cup {[t |-> "READY", src |-> n]}

\* >
\* > on receiving <READY> from f+1 parties:
\* >   send <READY> to all parties
\* >
HandleReadySupport(n) ==
    /\ \E q \in QF1: \A qn \in q: \E m \in msgs : m.t = "READY" /\ m.src = qn
    /\ msgs' = msgs \cup {[t |-> "READY", src |-> n]}

\* >
\* > on receiving <READY> from n-f parties:
\* >   if sᵢ is valid:
\* >     out = true
\* >     output sᵢ
\* >
HandleReadyQuorum(n) ==
    /\ \E q \in QNF: \A qn \in q: \E m \in msgs : m.t = "READY" /\ m.src = qn
    /\ msgs' = msgs \cup {[t |-> "OUTPUT", src |-> n]}

\* >
\* > on receiving <IMPLICATE, j, skⱼ>:
\* >   sⱼ := PKI.Dec(eⱼ, skⱼ)
\* >   if decrypt fails or VSS.Verify(C, j, sⱼ) == false:
\* >     if out == true:
\* >       send <RECOVER, i, skᵢ> to all parties
\* >       return
\* >
HandleImplicate(n) ==
    \E m, md \in msgs:
      /\ m.t = "IMPLICATE"
      /\ md.t = "DEAL" /\ md.rcp = n
      /\ ~md.deal[m.src] \* Check if the deal was invalid.
      /\ msgs' = msgs \cup {[t |-> "RECOVER", src |-> n, sec |-> TRUE]}
      \* False RECOVER can only be produced by the Faulty nodes.
      \* Those messages were added in the Init state already.

\* >
\* >     on receiving <RECOVER, j, skⱼ>:
\* >       sⱼ := PKI.Dec(eⱼ, skⱼ)
\* >       if VSS.Verify(C, j, sⱼ): T = T ∪ {sⱼ}
\* >
\* >     wait until len(T) >= f+1:
\* >       sᵢ = SSS.Recover(T, f+1, n)(i)
\* >       out = true
\* >       output sᵢ
\* >
HandleRecover(n) ==
    \E q \in QF1:
      /\ \A qn \in q: \E m, md \in msgs:
           /\ m.t = "RECOVER" /\ m.src = qn \* F+1 RECOVER messages received.
           /\ md.t = "DEAL" /\ md.rcp = n   \* We have received our DEAL.
           /\ md.deal[m.src]                \* The sender of RECOVER had a correct DEAL.
      /\ msgs' = msgs \cup {[t |-> "OUTPUT", src |-> n]}

--------------------------------------------------------------------------------
(*                                 SPEC                                       *)
Init ==
    msgs = UNION {
      [t: {"OK", "READY", "OUTPUT"}, src: Faulty],
      [t: {"IMPLICATE", "RECOVER"}, src: Faulty, sec: BOOLEAN]
    }

NodeActions(n) ==
    \/ HandleDeal(n) \/ HandleOK(n) \/ HandleReadySupport(n) \/ HandleReadyQuorum(n)
    \/ HandleImplicate(n) \/ HandleRecover(n)
Next ==
    \/ Input \/ RBC
    \/ \E n \in Fair : NodeActions(n) \* Outputs by Faulty nodes done on Init.
Fairness ==
    /\ WF_vars(Input \/ RBC)
    /\ \A n \in Fair: /\ WF_vars(NodeActions(n))
Spec == Init /\ [][Next]_vars /\ Fairness

--------------------------------------------------------------------------------
(*                               PROPERTIES                                   *)

\* A helper definition, not a property on its own.
AllFairNodesOutput ==
    \A n \in Fair: \E m \in msgs: m.t = "OUTPUT" /\ m.src = n

\* That's a simplified property, to make sure the algorithm
\* works with a non-faulty dealer.
FairDealerAlwaysOK ==
    Dealer \notin Faulty => <>AllFairNodesOutput

\* If we have enough of good deals, all fair nodes will receive their deals.
FairNodesWillReceiveDeals ==
    LET enough == \E md \in msgs, q \in QNF:
                    /\ md.t = "DEAL"
                    /\ \A qn \in q: md.deal[qn]
    IN enough ~> AllFairNodesOutput

\* Either all fair nodes output, or none of them.
FairNodeOutputImpliesAllFair ==
    (\E m \in msgs, n \in Fair: m.t = "OUTPUT" /\ m.src = n) ~> AllFairNodesOutput

\* Check the negative case: faulty dealer with to many faulty shares means
\* no fair nodes will event output.
ToMuchFaultyDealsMeansNoOutput ==
    (\E md \in msgs, q \in QF1: md.t = "DEAL" /\ \A qn \in q: ~md.deal[qn])
    => [](\A n \in Fair: ~\E m \in msgs: m.t = "OUTPUT" /\ m.src = n)

THEOREM Spec =>
    /\ []TypeOK
    /\ FairDealerAlwaysOK
    /\ FairNodesWillReceiveDeals
    /\ FairNodeOutputImpliesAllFair
    /\ ToMuchFaultyDealsMeansNoOutput
PROOF OMITTED \* Checked by the TLC

--------------------------------------------------------------------------------
(*                              SOME PROOFS                                   *)
LEMMA Spec => []TypeOK
  <1>0. Faulty \subseteq Nodes BY FaultyAssm
  <1>1. Init => TypeOK  BY <1>0 DEF TypeOK, Init, Msgs
  <1>2. TypeOK /\ [Next]_vars => TypeOK'
    <2> SUFFICES ASSUME TypeOK, Next PROVE TypeOK' BY DEF TypeOK, vars
    <2>1. CASE Input
      <3>1. CASE Dealer \in Faulty BY <2>1, <3>1, DealerAssm DEF Input, TypeOK, Msgs, Deal
      <3>2. CASE Dealer \notin Faulty
        <4>1. [t |-> "DEAL_BC", deal |-> [n \in Nodes |-> TRUE]] \in Msgs BY DEF Msgs, Deal
        <4>q. QED BY <4>1, <2>1, <3>2, DealerAssm DEF Input, TypeOK, Msgs
      <3>q. QED BY <2>1, <3>1, <3>2 DEF Input
    <2>2. CASE RBC BY <2>2, <1>0 DEF RBC, rbc!Broadcast, TypeOK, Msgs
    <2>3. CASE \E n \in Fair : NodeActions(n)
      <3> PICK n \in Fair : NodeActions(n) BY <2>3
      <3> n \in Nodes BY <1>0 DEF Fair
      <3>1. CASE HandleDeal(n)
        <4> [t |-> "OK",        src |-> n              ] \in Msgs BY DEF Msgs
        <4> [t |-> "IMPLICATE", src |-> n, sec |-> TRUE] \in Msgs BY DEF Msgs
        <4>q. QED BY <3>1 DEF HandleDeal, TypeOK, Msgs
      <3>2. CASE HandleOK(n) BY <3>2 DEF HandleOK, TypeOK, Msgs
      <3>3. CASE HandleReadySupport(n) BY <3>3 DEF HandleReadySupport, TypeOK, Msgs
      <3>4. CASE HandleReadyQuorum(n) BY <3>4 DEF HandleReadyQuorum, TypeOK, Msgs
      <3>5. CASE HandleImplicate(n)
        <4> [t |-> "RECOVER", src |-> n, sec |-> TRUE] \in Msgs BY DEF Msgs
        <4>q. QED BY <3>5 DEF HandleImplicate, TypeOK, Msgs
      <3>6. CASE HandleRecover(n) BY <3>6 DEF HandleRecover, TypeOK, Msgs
      <3>q. QED BY <2>3, <3>1, <3>2, <3>3, <3>4, <3>5, <3>6 DEF NodeActions
    <2>q. QED BY <2>1, <2>2, <2>3 DEF Next
  <1>3. QED BY <1>1, <1>2, PTL DEF Spec, TypeOK

================================================================================
