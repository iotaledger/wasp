---------- MODULE RBC -----------
(*
Here we are modelling the RBC as a black-box, only considering its properties.
It waits for a message to be broadcasted, and produces messages to be delivered
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
CONSTANTS Nodes, Faulty, Dealer
CONSTANT MsgP(_)    \* Predicate detecting the message to be broadcasted.
CONSTANT MsgS(_, _) \* Operator producing a message to be delivered.
VARIABLE msgs
(*
We have two cases here, depending on the correctness of the dealer:
  - If the Dealer is correct, then RBC has to deliver to all the Fair nodes, and maybe to some faulty nodes.
  - If dealer is Faulty, then two cases are possible:
      - Either all the Fair nodes (and maybe some faulty) will deliver a message;
      - Or no Fair node will deliver the message. The faulty nodes can still deliver it.
*)
Broadcast ==
    \E m \in msgs, someFaulty \in SUBSET Faulty: MsgP(m) /\
        \/ /\ Dealer \notin Faulty \* Deliver to all the fair nodes, and maybe to some faulty.
           /\ msgs' = msgs \cup {MsgS(m, n) : n \in Nodes \ someFaulty}
        \/ /\ Dealer \in Faulty \* Deliver either to all Fair nodes plus some Faulty, or just to some faulty.
           /\ \E deliverAt \in {(Nodes \ someFaulty)} \cup SUBSET Faulty:
                msgs' = msgs \cup {MsgS(m, n) : n \in deliverAt}

=================================
