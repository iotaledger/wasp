--------------------------- MODULE WaspChainRecovery ---------------------------
(*
This specification is concerned with a Chain's committee recovery in the case,
when we have more than F nodes failed by crashing and probably loosing the
persistent storage (the SC state still has to be available).

The goal is to define rules for a node to recover and join the committee.
It is complicated to rejoin to the existing consensus instance because of
big number of protocols involved. Here we considering rejoin by proposing
to abandon the existing log entry and proceed to the next one. The approach
can be summarized as:

  - Do not participate in the LI after restart, if it is the latest logged
    as the current in the persistent storage. We probably sent some messages
    in it, so participating after the restart can render us byzantine faulty.

  - Proceed to the next LI, if there is N-F nodes proposing that. Support such
    proposal is there is at least F+1 proposals.

  - RecoverTimeout is needed to break some cases. E.g. half of nodes are running
    consensus, and other half has restarted and asking for the next LI.

*)
EXTENDS Integers, WaspByzEnv \* Defines CN, FN, AN, N, F, Q?F.

CONSTANT MaxLogIndex
ASSUME LogIndexAssms == MaxLogIndex >= 0
NoLogIndex == -1
LogIndex == 0..MaxLogIndex
OptLogIndex == LogIndex \cup {NoLogIndex}
NextLIExist(li) == li < MaxLogIndex

VARIABLE storage    \* Persistent store.
VARIABLE running    \* Is the node running?
VARIABLE logIndex   \* The current/latest LI the node works on.
VARIABLE consFrom   \* Participate in the consensus form the specified log index.
VARIABLE msgs
vars == <<storage, running, logIndex, consFrom, msgs>>


Msg == UNION {
    [t: {"CONS_IN"}, src: CN, li: LogIndex],    \* Consensus input.
    [t: {"NEXT_LI"}, src: CN, li: OptLogIndex]  \* Proposal to proceed with LI+1.
}

MsgNextLI(n, li) == [t |-> "NEXT_LI", src |-> n, li |-> li]
MsgConsIn(n, li) == [t |-> "CONS_IN", src |-> n, li |-> li]
MsgQuorum(t, q, li) == \A qn \in q : \E m \in msgs : m.t = t /\ m.src = qn /\ m.li = li

TypeOK ==
    /\ storage   \in [CN -> OptLogIndex]
    /\ running   \in [CN -> BOOLEAN]
    /\ logIndex  \in [CN -> OptLogIndex]
    /\ consFrom  \in [CN -> OptLogIndex]    \* TODO: Do we need this? Maybe just logIndex?
    /\ msgs \subseteq Msg


--------------------------------------------------------------------------------

(*
A non-running node can start up. In this case it has to read its log index
from the persistent store, if any.

NOTE: Here we increase LogIndex, but con't sen the CONS_IN message. This way we
just skip participation in the same consensus instance, we probably participated
before the crash.
*)
Startup(n) ==
    /\ NextLIExist(storage[n])                              \* Just to have a bounded model.
    /\ ~running[n]
    /\ running' = [running EXCEPT ![n] = TRUE]
    /\ logIndex' = [logIndex EXCEPT ![n] = storage[n]+1]    \* Will read NoLogIndex, if DB was lost.
    /\ consFrom' = [consFrom EXCEPT ![n] = storage[n]+1]    \* Only participate in the next LI.
    /\ msgs' = msgs \cup {MsgNextLI(n, storage[n])}         \* We want to proceed to the next LI.
    /\ UNCHANGED <<storage>>

(*
Support proceeding to the next round???
TODO: Do we need that? Maybe just wait for the timeout?
*)
UponQ1FNextLI(n, q) == n \in CN /\ q \in Q1F /\ \* Parameter checks only, can be removed.
    \E li \in LogIndex :
        /\ running[n]                                   \* Node is running.
        /\ li+1 >= logIndex[n]                          \* Log index is not in the past.
        /\ li+1 >= consFrom[n]                          \* Don't participate, if there is change for duplicate insts.
        /\ MsgQuorum("NEXT_LI", q, li)                  \* Enough of proposals to support them.
        /\ msgs' = msgs \cup {MsgNextLI(n, li)}         \* Support the attempt to proceed.
        /\ UNCHANGED <<storage, running, consFrom, logIndex>>

UponQNFNextLI(n, q) == n \in CN /\ q \in QNF /\ \* Parameter checks only, can be removed.
    /\ \E li \in OptLogIndex :
        /\ NextLIExist(li)                              \* Just to have a bounded model.
        /\ running[n]                                   \* Node is running.
        /\ li+1 >= logIndex[n]                          \* Log index is not in the past.
        /\ li+1 >= consFrom[n]                          \* Don't participate, if there is change for duplicate insts.
        /\ MsgQuorum("NEXT_LI", q, li)                  \* Enough of proposals to proceed.
        /\ logIndex'  = [logIndex  EXCEPT ![n] = li+1]  \* Proceed to the next LI.
        /\ msgs'      = msgs \cup {MsgConsIn(n, li+1)}  \* Start participating in the consensus.
        /\ UNCHANGED <<storage, running, consFrom>>

ConsensusDone(n, q) == n \in CN /\ q \in QNF /\ \* Parameter checks only, can be removed.
    /\ NextLIExist(logIndex[n])                             \* Just to have a bounded model.
    /\ running[n]                                           \* We are running.
    /\ MsgQuorum("CONS_IN", {n}, logIndex[n])               \* We have started to participate.
    /\ MsgQuorum("CONS_IN", q, logIndex[n])  \* Enough nodes participate in the consensus.
    /\ logIndex'  = [logIndex  EXCEPT ![n] = @ + 1]         \* Go to the next log index.
    /\ storage'   = [storage EXCEPT ![n] = logIndex[n] + 1] \* Have to persist it.
    /\ msgs'      = msgs \cup {MsgConsIn(n, logIndex[n])}   \* Start to participate in the next LI.
    /\ UNCHANGED <<running, consFrom>>

(*
A recovery timeout can happen while consensus is running.
*)
RecoveryTimeout(n) ==
    /\ running[n]                                           \* This node is running.
    /\ MsgQuorum("CONS_IN", {n}, logIndex[n])               \* This node has contributed to the consensus.
    /\ msgs' = msgs \cup {MsgNextLI(n, logIndex[n])}        \* Propose to proceed to the next LI.
    /\ UNCHANGED <<storage, running, consFrom, logIndex>>

(*
A crash can happen any time. It can involve disk loss.
*)
Crash(n) ==
    /\ running[n]
    /\ running'  = [running  EXCEPT ![n] = FALSE]
    /\ logIndex' = [logIndex EXCEPT ![n] = NoLogIndex]
    /\ consFrom' = [consFrom EXCEPT ![n] = NoLogIndex]
    /\ \/ msgs' = { m \in msgs : m.src # n } \* Drop the node's messages,
       \/ UNCHANGED msgs                     \* or retain them.
    /\ \/ storage' = [storage EXCEPT ![n] = NoLogIndex] \* DB is either lost on crash,
       \/ UNCHANGED storage                             \* or retained.

--------------------------------------------------------------------------------
Init ==
    /\ storage   = [n \in CN |-> NoLogIndex]
    /\ running   = [n \in CN |-> FALSE]
    /\ logIndex  = [n \in CN |-> NoLogIndex]
    /\ consFrom  = [n \in CN |-> NoLogIndex]
    /\ msgs      = { m \in Msg : m.src \in FN }

Next ==
    \E n \in CN :
        \/ Startup(n)
        \/ \E q \in Q1F : UponQ1FNextLI(n, q)
        \/ \E q \in QNF : UponQNFNextLI(n, q)
        \/ \E q \in QNF : ConsensusDone(n, q)
        \/ RecoveryTimeout(n)
        \/ Crash(n)

Fairness == WF_vars(
    \E n \in CN :
        \/ Startup(n)
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

================================================================================
