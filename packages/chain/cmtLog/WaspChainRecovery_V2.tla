-------------------------- MODULE WaspChainRecovery_V2 -------------------------
(*
This specification is concerned with a Chain's committee recovery in the case,
when we have more than F nodes failed by crashing and probably loosing the
persistent storage (the SC state still has to be available).

The goal is to define rules for a node to recover and join the committee.
It is complicated to rejoin to the existing consensus instance because of
big number of protocols involved. Here we considering rejoin by proposing
to abandon the existing log entry and proceed to the next one. The approach
can be summarized as:

  - Do not participate in the LI after a restart, if the LogIndex for that consensus
    is the latest recorded in the persistent storage. We probably sent some messages
    in it, so participating after the restart can render us byzantine faulty.

  - Proceed to the next LI, if there is N-F nodes proposing that. Support such
    proposal if there is at least F+1 proposals (TODO: Do we need the F+1 case?).

  - RecoverTimeout is needed to break some cases. E.g. half of nodes are running
    consensus, and other half has restarted and asking for the next LI.

TODO: Consider ConsensusDone from the future LogIndexes.
TODO: Consider events: AOReceived, AORejected.

*)
EXTENDS Integers, WaspByzEnv \* Defines CN, FN, AN, N, F, Q?F.

CONSTANT MaxLogIndex
CONSTANT AliasOutputs
ASSUME LogIndexAssms == MaxLogIndex >= 0
NoLogIndex == -1
LogIndex == 0..MaxLogIndex
OptLogIndex == LogIndex \cup {NoLogIndex}
NextLIExist(li) == li < MaxLogIndex

VARIABLE storage    \* Persistent store.
VARIABLE logIndex   \* The current/latest LI the node works on.
VARIABLE ledgerCnf  \* The L1 Ledger: Confirmed till this index (inclusive).
VARIABLE ledgerRcv  \* Per-node: received till this index (exclusive).
VARIABLE ledgerExt  \* For each AO -- is it external or internal ("E", "I", "?").
VARIABLE msgs
vars == <<storage, logIndex, ledgerCnf, ledgerRcv, msgs>>


Msg == UNION {
    [t: {"CONS_IN"}, src: CN, li: LogIndex],    \* Consensus input.
    [t: {"NEXT_LI"}, src: CN, li: OptLogIndex]  \* Proposal to proceed with LI+1.
}
Ext == {"?", "I", "E"}

MsgNextLI(n, li) == [t |-> "NEXT_LI", src |-> n, li |-> li]
MsgConsIn(n, li) == [t |-> "CONS_IN", src |-> n, li |-> li]
MsgQuorum(t, q, li) == \A qn \in q : \E m \in msgs : m.t = t /\ m.src = qn /\ m.li = li

TypeOK ==
    /\ storage   \in [CN -> OptLogIndex]
    /\ logIndex  \in [CN -> OptLogIndex]
    /\ msgs \subseteq Msg


--------------------------------------------------------------------------------
(*
Here we model the L1 Ledger. We model it here as a linear chain of alias outputs.
We assume the initial AO is already present (chain creation is out of scope here).
*)

LedgerTypeOK ==
    /\ ledgerCnf \in AO
    /\ ledgerRcv \in [CN -> AO]
    /\ ledgerExt \in [AO -> Ext]

(*
The initial state for the ledger.
*)
LedgerInit ==
    /\ ledgerCnf = 0                \* AO at index 0 is already confirmed.
    /\ ledgerRcv = [n \in CN |-> 0] \* No AOs received by any nodes.
    /\ ledgerExt = [ao \in AO |-> "?"]


(*
Determines, if a node can receive an AO at the specified.
*)
LedgerRcv(n, chainIdx) ==
    /\ chainIdx <= ledgerCnf
    /\ chainIdx >= ledgerRcv[n]
    /\ ledgerRcv' = [ledgerRcv EXCEPT ![n] = chainIdx]
    /\ UNCHANGED <<ledgerCnf, ledgerExt>>

(*
TX confirmed in L1, posted by the current committee.
*)
LedgerCnfInternal ==
    \E ci \n ChainIdx, li \in LogIndex :
        /\ ci = ledgerCnf+1
        /\ \E q \in QNF:
               \A n \in QNF :
                   \E m \in msgs :
                       /\ m.t = "POST_TX"
                       /\ m.ci = ci
                       /\ m.li = li
                       /\ m.src = n
        /\ ledgerCnf' = ci
        /\ ledgerExt' = [ledgerExt EXCEPT ![ci] = "I"]
        /\ UNCHANGED <<storage, logIndex, ledgerRcv, msgs>>

(*
TX confirmed in L1, but that's by other committee, or some rotation.
NOTE: This covers also the LedgerCnfInternal, because
LedgerCnfInternal => LedgerCnfExternal, if ledgerExt would be dropped.
*)
LedgerCnfExternal ==
    \E ci \n ChainIdx, li \in LogIndex :
        /\ ci = ledgerCnf+1
        /\ ledgerCnf' = ci
        /\ ledgerExt' = [ledgerExt EXCEPT ![ci] = "E"]
        /\ UNCHANGED <<storage, logIndex, ledgerRcv, msgs>>

LedgerActions == LedgerCnfInternal \/ LedgerCnfExternal


--------------------------------------------------------------------------------
(*
TODO: Everything is incomplete bellow.

A node / chain / committee:
  - Can receive an AO from L1. They are delivered in order, but some can be skipped.
    Only confirmed AOs are received. We ignore reorgs and rejects in this model.
  - Consensus can be started upon receiving an AO and deciding a LI.
  - Consensus can be completed, timed-out or skipped.
*)

ReceivedAOFromL1(n) == n \in CN \* Parameter checks only, can be removed.
    /\ NextLIExist(storage[n])                              \* Just to have a bounded model.
    /\ logIndex' = [logIndex EXCEPT ![n] = storage[n]+1]    \* Will read NoLogIndex, if DB was lost.
    /\ msgs' = msgs \cup {MsgNextLI(n, storage[n])}         \* We want to proceed to the next LI.
    /\ UNCHANGED <<storage>>

(*
Support proceeding to the next round???
TODO: Do we need that? Maybe just wait for the timeout?
*)
UponQ1FNextLI(n, q) == n \in CN /\ q \in Q1F /\ \* Parameter checks only, can be removed.
    \E li \in LogIndex :
        /\ li+1 >= logIndex[n]                          \* Log index is not in the past.
        /\ MsgQuorum("NEXT_LI", q, li)                  \* Enough of proposals to support them.
        /\ msgs' = msgs \cup {MsgNextLI(n, li)}         \* Support the attempt to proceed.
        /\ UNCHANGED <<storage, logIndex>>

UponQNFNextLI(n, q) == n \in CN /\ q \in QNF /\ \* Parameter checks only, can be removed.
    /\ \E li \in OptLogIndex :
        /\ NextLIExist(li)                              \* Just to have a bounded model.
        /\ li+1 >= logIndex[n]                          \* Log index is not in the past.
        /\ MsgQuorum("NEXT_LI", q, li)                  \* Enough of proposals to proceed.
        /\ logIndex'  = [logIndex EXCEPT ![n] = li+1]   \* Proceed to the next LI.
        /\ storage'   = [storage  EXCEPT ![n] = li+1]   \* Save the LI for which we have sent consensus messages.
        /\ msgs'      = msgs \cup {MsgConsIn(n, li+1)}  \* Start participating in the consensus.

ConsensusDone(n, q) == n \in CN /\ q \in QNF /\ \* Parameter checks only, can be removed.
    /\ NextLIExist(logIndex[n])                             \* Just to have a bounded model.
    /\ MsgQuorum("CONS_IN", {n}, logIndex[n])               \* We have started to participate.
    /\ MsgQuorum("CONS_IN", q, logIndex[n])  \* Enough nodes participate in the consensus.
    /\ logIndex'  = [logIndex  EXCEPT ![n] = @ + 1]         \* Go to the next log index.
    /\ storage'   = [storage EXCEPT ![n] = logIndex[n]+1]   \* Have to persist it.
    /\ msgs'      = msgs \cup {MsgConsIn(n, logIndex[n]+1)} \* Start to participate in the next LI.

(*
A recovery timeout can happen while consensus is running.
*)
RecoveryTimeout(n) ==
    /\ MsgQuorum("CONS_IN", {n}, logIndex[n])               \* This node has contributed to the consensus.
    /\ msgs' = msgs \cup {MsgNextLI(n, logIndex[n])}        \* Propose to proceed to the next LI.
    /\ UNCHANGED <<storage, logIndex>>

(*
A crash can happen any time. It can involve disk loss.
*)
Crash(n) ==
    /\ logIndex' = [logIndex EXCEPT ![n] = NoLogIndex]
    /\ \/ msgs' = { m \in msgs : m.src # n } \* Drop the node's messages,
       \/ UNCHANGED msgs                     \* or retain them.
    /\ \/ storage' = [storage EXCEPT ![n] = NoLogIndex] \* DB is either lost on crash,
       \/ UNCHANGED storage                             \* or retained.

--------------------------------------------------------------------------------
Init ==
    /\ storage   = [n \in CN |-> NoLogIndex]
    /\ logIndex  = [n \in CN |-> NoLogIndex]
    /\ msgs      = { m \in Msg : m.src \in FN }

Next ==
    \E n \in CN :
        \/ ReceivedAOFromL1(n)
        \/ \E q \in Q1F : UponQ1FNextLI(n, q)
        \/ \E q \in QNF : UponQNFNextLI(n, q)
        \/ \E q \in QNF : ConsensusDone(n, q)
        \/ RecoveryTimeout(n)
        \/ Crash(n)

Fairness == WF_vars(
    \E n \in CN :
        \/ ReceivedAOFromL1(n)
        \/ \E q \in Q1F : q \subseteq CN /\ UponQ1FNextLI(n, q) \* Consider CN here.
        \/ \E q \in QNF : q \subseteq CN /\ UponQNFNextLI(n, q) \* Consider CN here.
        \/ \E q \in QNF : q \subseteq CN /\ ConsensusDone(n, q) \* Consider CN here.
        \/ RecoveryTimeout(n)
        \* No fairness on the Crash action.
)

Spec == Init /\ [][Next]_vars /\ Fairness

--------------------------------------------------------------------------------
ReachesLastLI ==
    <> \A n \in CN : running[n] => (logIndex[n] = MaxLogIndex)

AONotExtStable == (* Not overridden. *)
    [] (\A ao \in AO : \E e \in Ext \ {"?"} :  ledgerExt[i] = e => [] ledgerExt[i] = e)

THEOREM Spec =>
    /\ []TypeOK
    /\ ReachesLastLI
    /\ AONotExtStable

================================================================================
