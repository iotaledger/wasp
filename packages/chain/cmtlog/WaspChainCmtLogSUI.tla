---- MODULE WaspChainCmtLogSUI -------------------------------------------------
(******************************************************************************)
(*
This is a specification for a committee log adjusted to the SUI based L1.
Its purpose is to ensure self-stabilization of the consensus. I.e. crashed
nodes can reboot and join the consensus without storing its internal state
nor messages persistently.

Assume LastLI is the latest LogIndex on which the node could work.
  - A node increases LastLI to L on completion of consensus on L-1
  - or upon reception of consensus inputs at index L from F+1 nodes.

The algorithm in general:

  - A node can join the LastLI with ⊥, if LastLI=MinLI.
    This happens on boot.

  - A node can join the LastLI-1 or lower instance at any time with input ⊥.
    Others are already working on the LastLI, so we will not damage anything.

  - A node can join the LastLI after LastLI-1 is completed or a timeout exceeds.
    The node proposes input as follows:

      - If LastLI-1 resulted in AO_x, propose AO_x.
      - If LastLI-1 resulted in ⊥, propose latest AO from LocalView.
      - If LastLI-1 resulted in a timeout, propose ⊥.

Notes on modeling:

  - The L1 consensus is modeled as a single variable holding the log decisions made till now.
    Each node might see a different prefix of that log.
    The log is advanced when at least 1 L2 node proposes a TX.
    We assume only TXes signed by a quorum of a committee can be proposed.

  - The L2 consensus is modelled the same way, just a log is advanced when N-F nodes propose their input.

  - To make state space a bit less, we assume the decided AOs will come in some specific order.
    This is modelled by SomeAoOrder, which uses CHOOSE underneath.
*)
(******************************************************************************)
EXTENDS FiniteSets, Sequences, SequencesExt, Integers, WaspByzEnv, TLAPS
CONSTANT AO
CONSTANT LI \* Consensus Log Indexes, subset of Nat.

ASSUME AO # {}
ASSUME LI \subseteq Nat /\ 0 \in LI
ASSUME Cardinality(AO) >= Cardinality(LI)

VARIABLE l1AOs      \* A chain of already committed Anchor objects in L1.
VARIABLE l1View     \* Latest AO (by position) seen by a node.
VARIABLE l2MinLI    \* Minimal LogIndex a node can participate after a reboot.
VARIABLE l2LastLI   \* The latest LI a node considers to participate.
VARIABLE l2Input    \* Input proposed by nodes to the L2 consensus.
VARIABLE l2Decision \* L2 consensus decision, not necessary immediatelly known to the nodes.
VARIABLE l2Output   \* L2 consensus decision known by the nodes.
vars == <<l1AOs, l1View, l2MinLI, l2LastLI, l2Input, l2Decision, l2Output>>

Node == AN

BOT         == 0            \* ⊥ as an AO.
Pending     == -1           \* Pending, not yet proposed AO.
SomeAoOrder == SetToSeq(AO) \* We don't need to consider all the permulations.

TypeOK ==
    /\ l1AOs \in Seq(AO)
    /\ l1View \in [Node -> DOMAIN SomeAoOrder]
    /\ l2MinLI \in [Node -> LI]
    /\ l2LastLI \in [Node -> LI]
    /\ l2Input \in [Node -> [LI -> AO \cup {BOT, Pending}]]
    /\ l2Decision \in [LI -> AO \cup {BOT, Pending}]
    /\ l2Output \in [Node -> [LI -> AO \cup {BOT, Pending}]]


--------------------------------------------------------------------------------
(******************************************************************************)
(*                              Utilities                                     *)
(******************************************************************************)

(*
Checks if next follows ao in the predefined order of AOs.
*)
isNextAO(ao, next) ==
    \E i \in DOMAIN SomeAoOrder:
        /\ i+1 \in DOMAIN SomeAoOrder
        /\ SomeAoOrder[i] = ao
        /\ SomeAoOrder[i+1] = next


(*
Just to make expressions shorter.
*)
maxLI(a, b) == IF a >= b THEN a ELSE b


(*
Returns an AO which is reported as latest in the L1 at a particular node.
*)
lastSeenAO(n) == SomeAoOrder[l1View[n]]


(*
Partial action -- proposes an input, if not proposed yet.
*)
propose(n, li, v) ==
    /\ l2MinLI[n] <= li             \* Not before the MinLI.
    /\ l2Input[n][li] = Pending     \* Not proposed yet.
    /\ l2Input' = [l2Input EXCEPT ![n][li] = v]
    /\ UNCHANGED <<l1AOs, l1View, l2MinLI, l2LastLI, l2Decision, l2Output>>


(*
We use this to clear the node inputs when it drops the consensus instances
locally. This way it is not contributing to the result anymore, thus it cannot
be counted to the quorum, if decision has not be reached yet.
*)
dropInputsBefore(nodeInputs, beforeLI) ==
    [ li \in LI |-> IF li < beforeLI THEN Pending ELSE nodeInputs[li] ]


--------------------------------------------------------------------------------
(******************************************************************************)
(*                               Actions                                      *)
(******************************************************************************)

(*
A lagging node can receive more recent L1 information at any time.
The decisions are always received in order.
*)
AdvanceNodeView(n) ==
    \E i \in DOMAIN l1AOs :
        /\ i > l1View[n]
        /\ l1View' = [l1View EXCEPT ![n] = i]
        /\ UNCHANGED <<l1AOs, l2MinLI, l2LastLI, l2Input, l2Decision, l2Output>>


(*
A node advances its LastLI when it receives consensus inputs from F+1 nodes.
*)
AdvanceLastLI(n) ==
    \E li \in LI :
        /\ li > l2LastLI[n]
        /\ \E q \in Q1F : \A qn \in q: l2Input[qn][li] # Pending
        /\ l2LastLI' = [l2LastLI EXCEPT ![n] = li]
        /\ UNCHANGED <<l1AOs, l1View, l2MinLI, l2Input, l2Decision, l2Output>>


(*
A node provides ⊥ at MinLI any time.
*)
ProposeOnMinLI(n) ==
    LET li == l2MinLI[n]
    IN  propose(n, li, BOT)


(*
If a node sees LastLI instance started (either by itself, or by a qyorum of other nodes),
the node can propose ⊥ for the previous LI, if not yet provided.
*)
ProposeOnPrevLI(n) ==
    LET li == l2LastLI[n]
    IN  /\ li > l2MinLI[n]          \* MinLI will be handled by ProposeOnMinLI.
        /\ propose(n, li-1, BOT)    \* Propose ⊥ for the previous LI.


(*
When a node gets ⊥ output from LastLI-1, it provides own AO, as reported by the LocalView.
*)
ProposeOnPrevLIBot(n) ==
    LET li == l2LastLI[n]
    IN  /\ li > l2MinLI[n]                  \* MinLI will be handled by ProposeOnMinLI.
        /\ l2Output[n][li-1] = BOT          \* If we have ⊥ as an output from the previous LI.
        /\ propose(n, li, lastSeenAO(n))    \* Latest AO received from L1.


(*
When a node gets output from the LastLI-1, it provides it as an input to the LastLI.
*)
ProposeOnPrevLIOut(n) ==
    LET li == l2LastLI[n]
    IN  /\ li > l2MinLI[n]                      \* MinLI will be handled by ProposeOnMinLI.
        /\ l2Output[n][li-1] \in AO             \* If we have ⊥ as an output from the previous LI.
        /\ propose(n, li, l2Output[n][li-1])    \* Output from the consensus becomes input for the next one.


(*
A node can propose ⊥ for the LastLI after a timeout.
*)
ProposeOnTimeout(n) ==
    LET li == l2LastLI[n]
    IN  /\ li > l2MinLI[n]              \* MinLI will be handled by ProposeOnMinLI.
        /\ l2Input[n][li-1] # Pending   \* We have proposed something to previous LI.
        /\ propose(n, li, BOT)


(*
This action models the L2 consensus itself.
It decides when N-F nodes provide their inputs.
If there exist F+1 equal inputs, then the decision is equal to it.
Otherwise the decision is SKIP=⊥.
*)
DecideOnL2 ==
    \E li \in LI:
        LET existQ1N(ao) == \E q \in Q1F: \A qn \in q: l2Input[qn][li] = ao
            existQNF     == \E q \in QNF: \A qn \in q: l2Input[qn][li] # Pending
        IN  /\  existQNF
            /\  \/  \E inpAO, outAO \in AO :
                        /\ existQ1N(inpAO)
                        /\ isNextAO(inpAO, outAO)
                        /\ l2Decision' = [l2Decision EXCEPT ![li] = outAO]
                \/  /\ ~\E ao \in AO : existQ1N(ao)
                    /\ l2Decision' = [l2Decision EXCEPT ![li] = BOT]
            /\ UNCHANGED <<l1AOs, l1View, l2MinLI, l2LastLI, l2Input, l2Output>>


(*
The decisions are not necessary received by nodes in order by LI.
*)
ObtainL2Output(n) ==
    \E li \in LI :
        /\ l2MinLI[n] <= li \* We don't participate in LI<MinLI, so we cannot obtain the output.
        /\ l2Decision[li] # Pending
        /\ l2Output[n][li] = Pending
        /\ l2Output' = [l2Output EXCEPT ![n][li] = l2Decision[li]]
        /\ l2LastLI' = [l2LastLI EXCEPT ![n] = IF li+1 \in LI THEN li+1 ELSE @]
        /\ UNCHANGED <<l1AOs, l1View, l2MinLI, l2Input, l2Decision>>


(*
A node publish its consensus output if it sees the TX input as the latest AO in L1.
*)
PublishL2Output(n) ==
    \E li \in LI:
        /\ isNextAO(lastSeenAO(n), l2Output[n][li])
        /\ l1View[n] = Len(l1AOs)
        /\ l1AOs' = Append(l1AOs, l2Output[n][li])
        /\ UNCHANGED <<l1View, l2MinLI, l2LastLI, l2Input, l2Decision, l2Output>>


(*
It should be enough to have 2 instances running at each node.
We also drop our old inputs to model abort in participating old instances.
*)
DropOldInstances(n) ==
    /\ l2LastLI[n] > l2MinLI[n] + 1
    /\ l2MinLI' = [l2MinLI EXCEPT ![n] = l2LastLI[n] - 1]
    /\ l2Input' = [l2Input EXCEPT ![n] = dropInputsBefore(@, l2LastLI[n] - 1)]
    /\ UNCHANGED <<l1AOs, l1View, l2LastLI, l2Decision, l2Output>>


(*
On reboot a node sets its MinLI to the one for which it has not yet provided any input.
Also drop the old inputs, modeling this way abortion in the dropped consensus instances.
*)
Reboot(n) ==
    \E li \in LI:
        /\ \A i \in LI: i >= li => l2Input[n][li] = Pending
        /\ li > 0 /\ l2Input[n][li-1] # Pending
        /\ l2MinLI' = [l2MinLI EXCEPT ![n] = li]
        /\ l2LastLI' = [l2LastLI EXCEPT  ![n] = maxLI(@, li)]
        /\ l2Input' = [l2Input EXCEPT ![n] = dropInputsBefore(@, li)]
        /\ UNCHANGED <<l1AOs, l1View, l2Decision, l2Output>>


--------------------------------------------------------------------------------
(******************************************************************************)
(*                           The specification                                *)
(******************************************************************************)

Init ==
    /\ l1AOs = <<Head(SomeAoOrder)>>
    /\ l1View = [n \in Node |-> 1]
    /\ l2MinLI = [n \in Node |-> 0]
    /\ l2LastLI = [n \in Node |-> 0]
    /\ l2Input = [n \in Node |-> [li \in LI |-> Pending]]
    /\ l2Decision = [li \in LI |-> Pending]
    /\ l2Output = [n \in Node |-> [li \in LI |-> Pending]]


NextFair ==
    \/ \E n \in Node :
        \/ AdvanceNodeView(n)
        \/ AdvanceLastLI(n)
        \/ ProposeOnMinLI(n)
        \/ ProposeOnPrevLI(n)
        \/ ProposeOnPrevLIBot(n)
        \/ ProposeOnPrevLIOut(n)
        \/ ProposeOnTimeout(n)
        \/ ObtainL2Output(n)
        \/ PublishL2Output(n)
        \/ DropOldInstances(n)
    \/ DecideOnL2

NextFail ==
    \/ \E n \in Node :
        \/ Reboot(n)

Next == NextFair \/ NextFail


Spec == Init /\ [][Next]_vars /\ WF_vars(NextFair)


--------------------------------------------------------------------------------
(******************************************************************************)
(*                              Properties                                    *)
(******************************************************************************)

(*
MinLI cannot exceed the LastLI.
Otherwise the node is blocked.
*)
InvLI ==
    \A n \in Node: l2MinLI[n] <= l2LastLI[n]


(*
At least the latest LI has to decide, because there is no place for reboots
to happen because of the way we limit the state space.
*)
MaxLIWillDecide ==
    <> (\A li \in LI : (\A i \in LI: i <= li) => l2Decision[li] # Pending)


(*
If reboots stop happening, all remaining LIs will be decided.
The remaining LIs are those for which there is a quorum of nodes
at which LastLI does not exceed that LI.

For example, there can be some holes in the log with undecided LIs, but
only while the nodes keep restarting. After restarts stop occurring,
the LIs will be decided contiguously.
*)
NoRebootsAllDecide ==
    (<>[][NextFair]_vars) => <>[](
        \A li \in LI :
            (\E q \in QNF : \A qn \in q: l2LastLI[qn] <= li)
            => (l2Decision[li] # Pending)
    )


(*
Just a summary on what was model-checked.
*)
THEOREM
    Spec => /\ []TypeOK
            /\ []InvLI
            /\ MaxLIWillDecide
            /\ NoRebootsAllDecide
PROOF OMITTED \* Checked by the TLC.


================================================================================
