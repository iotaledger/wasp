---------------------------- MODULE WaspACSS -----------------------------------
EXTENDS Naturals, FiniteSets, TLC
CONSTANT Nodes
CONSTANT Dealer
CONSTANT Faulty
ASSUME NodesAssm == IsFiniteSet(Nodes)
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
Msgs == UNION {
    [t: {"DEAL_BC"},          deal: [Nodes -> BOOLEAN]],
    [t: {"DEAL"}, rcp: Nodes, deal: [Nodes -> BOOLEAN]],
    [t: {"OK", "READY", "OUTPUT"}, src: Nodes],
    [t: {"IMPLICATE", "RECOVER"}, src: Nodes, sec: BOOLEAN ]
}
TypeOK ==
    msgs \in SUBSET Msgs

---------- MODULE RBC -----------
(*
Here we are modelling the RBC as a blackbox, only considering its properties.
It waits for a message to be broadcasted, and produces messahes to be delivered
to particular nodes. The properties of the Uniform RBC:

URB-validity:
    If a process urb-delivers a message m, then message
    m as been previously urb-broadcast (by p_{m.sender}).
URB-integrity:
    A process urb-delivers a message m at most once.
    NOTE: We will ignore this property here, the consumer will have to handle it.
URB-termination-1:
    If a non-faulty process urb-broadcasts a
    message m, it urb-delivers the message m.
URB-termination-2:
    If a process urb-delivers a message m, then
    each non-faulty process urb-delivers the message m.
    This property sometimes called “uniform agreement”.
*)
CONSTANT MsgP(_)    \* Predicate detecting the message to be broadcasted.
CONSTANT MsgS(_, _) \* Operator producing a message to be delivered.
Broadcast ==
    \E msg \in msgs, someFaulty \in SUBSET Faulty: MsgP(msg) /\
        \/ /\ Dealer \in Faulty \* All fair or none of them.
           /\ \E deliverAt \in {(Nodes \ someFaulty)} \cup SUBSET Faulty:
                msgs' = msgs \cup {MsgS(msg, n) : n \in deliverAt}
        \/ /\ Dealer \notin Faulty
           /\ msgs' = msgs \cup {MsgS(msg, n) : n \in Nodes \ someFaulty}
=================================
rbc == INSTANCE RBC WITH
        MsgP <- LAMBDA m : m.t = "DEAL_BC",
        MsgS <- LAMBDA m, n : [t |-> "DEAL", rcp |-> n, deal |-> m.deal]

\* TODO: Move the crypto abstraction to a separate module?

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
    \E q \in QNF: \A qn \in q: \E m \in msgs : m.t = "OK" /\ m.src = qn
      /\ msgs' = msgs \cup {[t |-> "READY", src |-> n]}

\* >
\* > on receiving <READY> from f+1 parties:
\* >   send <READY> to all parties
\* >
HandleReadySupport(n) ==
    \E q \in QF1: \A qn \in q: \E m \in msgs : m.t = "READY"
      /\ msgs' = msgs \cup {[t |-> "READY", src |-> n]}

\* >
\* > on receiving <READY> from n-f parties:
\* >   if sᵢ is valid:
\* >     out = true
\* >     output sᵢ
\* >
HandleReadyQuorum(n) ==
    \E q \in QNF: \A qn \in q: \E m \in msgs : m.t = "READY"
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
    \E q \in QF1: \A qn \in q: \E m, md \in msgs:
      /\ m.t = "RECOVER"
      /\ md.t = "DEAL" /\ md.rcp = n
      /\ md.deal[m.src]
      /\ msgs' = msgs \cup {[t |-> "OUTPUT", src |-> n]}

--------------------------------------------------------------------------------
(*                                 SPEC                                       *)
Init ==
    msgs = UNION (
      {[t: {"OK", "READY", "OUTPUT"}, src: {n}] : n \in Faulty} \cup
      {[t: {"IMPLICATE", "RECOVER"}, src: {n}, sec: BOOLEAN ] : n \in Faulty}
    )

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
                    /\ \A n \in q: md.deal[n]
    IN enough ~> AllFairNodesOutput

\* Either all fair nodes output, or none of them.
FairNodeOutputImpliesAllFair ==
    (\E m \in msgs, n \in Fair: m.t = "OUTPUT" /\ m.src = n) ~> AllFairNodesOutput

\* Check the negative case: faulty dealer with to many faulty shares means
\* no fair nodes will evet output.
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

================================================================================
