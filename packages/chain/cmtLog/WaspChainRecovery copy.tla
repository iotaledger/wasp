--------------------------- MODULE WaspChainRecovery ---------------------------
(*
This specification is concerned with a Chain's committee recovery in the case,
when we have more than F nodes failed by crashing and probably loosing the
persistent storage (the SC state still has to be available).

The goal is to define rules for a node to recover and join the committee.
It is complicated to rejoin to the existing consensus instance because of
big number of protocols involved. Here we considering rejoin by proposing
to abandon the existing log entry and proceed to the next one.
*)
EXTENDS Integers, WaspByzEnv \* Defines CN, FN, AN, N, F, Q?F.

CONSTANT LogIndex
\* CONSTANTS Chain, Committee, AliasOutput \* TODO: Do we need these.
ASSUME LogIndexAssms == \A li \in LogIndex : 0..li \subseteq LogIndex
NoLogIndex == -1
OptLogIndex == LogIndex \cup {NoLogIndex}

VARIABLE storage    \* Persistent store.
VARIABLE running    \* Is the node running?
VARIABLE logIndex   \* The current LI the node works on.
VARIABLE consensus  \* Is a consensus now running at a node?
VARIABLE timeout    \* If a recovery timeout reached while running consensus?
VARIABLE msgs
vars == <<storage, running, logIndex, consensus, timeout, msgs>>

\* Log == {TRUE} \* Consider single log for now, otherwise use constants: Chain, Committee, AliasOutput.

Msg == UNION (
    \* [t: {"START_LOG"}, src: CN, ao: XXX], \* Can be modelled using NEXT_LI with li=NoLogIndex?
    [t: {"CONS_INP", "CONS_OUT"}, src: CN, li: LogIndex] \* Consensus started/done.
    [t: {"NEXT_LI"}, src: CN, li: OptLogIndex]
)

MsgNextLI(n, li) == [t |-> "NEXT_LI", src |-> n, li |-> li]
MsgQuorum(t, q, li) == \A qn \in q : \E m \in msgs : m.t = t /\ m.src = qn /\ m.li = li
MsgNextLIQuorum(q, li) == MsgQuorum("NEXT_LI", q, li)

TypeOK ==
    /\ storage   \in [CN -> OptLogIndex]
    /\ running   \in [CN -> BOOLEAN]
    /\ logIndex  \in [CN -> OptLogIndex]
    /\ consensus \in [CN -> BOOLEAN]
    /\ timeout   \in [CN -> BOOLEAN]


--------------------------------------------------------------------------------

(*
A crash can happen any time. It can involve disk loss.
*)
Crash(n) ==
    /\ running[n]
    /\ running'   = [running   EXCEPT ![n] = FALSE]
    /\ logIndex'  = [logIndex  EXCEPT ![n] = NoLogIndex]
    /\ consensus' = [consensus EXCEPT ![n] = FALSE]
    /\ timeout'   = [timeout   EXCEPT ![n] = FALSE]
    /\ \/ msgs' = { m \in msgs : m.src # n } \* Drop the node's messages,
       \/ UNCHANGED msgs                     \* or retain it.
    /\ \/ storage' = [storage EXCEPT ![n] = NoLogIndex] \* DB is lost on crash,
       \/ UNCHANGED <<storage>>                         \* or retained.

(*
A non-running node can start up. In this case it has to read its log index
from the persistent store, if any.
*)
Startup(n) ==
    /\ ~running[n]
    /\ running' = [running EXCEPT ![n] = TRUE]
    /\ logIndex' = [logIndex EXCEPT ![n] = storage[n]] \* Will read NoLogIndex, if DB was lost.
    /\ msgs' = msgs \cup {MsgNextLI(n, storage[n])}    \* We want to proceed to the next LI.
    /\ UNCHANGED <<storage, consensus, timeout>>

(*
A recovery timeout can happen while consensus is running.
*)
RecoveryTimeout(n) ==
    /\ running[n] /\ consensus[n] /\ ~timeout[n]
    /\ timeout' = [timeout EXCEPT ![n] = TRUE]
    /\ msgs' = msgs \cup {MsgNextLI(n, logIndex[n])}
    /\ UNCHANGED <<storage, running, logIndex, consensus>>

\* UponNextLINil ==
\*     /\ TRUE

UponQ1FNextLI(n, q) ==
    /\ n \in CN /\ q \in Q1F \* Parameter checks only, can be removed.
    /\ \E li \in LogIndex :
        /\ running[n]
        /\ li >= logIndex[n]
        /\ MsgNextLIQuorum(q, li)
        /\ logIndex'  = [logIndex  EXCEPT ![n] = li]
        /\ consensus' = [consensus EXCEPT ![n] = FALSE]
        /\ timeout'   = [timeout   EXCEPT ![n] = FALSE]
        /\ msgs' = msgs \cup {MsgNextLI(n, li)}

ConsensusStart(n) ==
    \E q \in QNF : ....

ConsensusDone(n) ==
    /\ logIndex[n] + 1 \in LogIndex                         \* Just to have a bounded model.
    /\ running[n] /\ consensus[n]                           \* We are running and participating.
    /\ \E q \in QNF : MsgQuorum("CONS_INP", q, logIndex[n]) \* Enough nodes participate in the consensus.
    /\ logIndex'  = [logIndex  EXCEPT ![n] = @ + 1]         \* Go to the next log index.
    /\ consensus' = [consensus EXCEPT ![n] = FALSE]         \* We don't participate in the next LI yet.
    /\ timeout'   = [timeout   EXCEPT ![n] = FALSE]         \* Timeout has not fired yet for the next LI.
    /\ storage'   = [storage EXCEPT ![n] = logIndex[n] + 1] \* Have to persist it.
    /\ msgs'      = msgs \cup {MsgNextLI(n, logIndex[n])}   \* Tell other, we want to go to the next LI.
    /\ UNCHANGED <<running>>

--------------------------------------------------------------------------------
Init ==
    /\ running   = [n \in CN |-> FALSE]
    /\ storage   = [n \in CN |-> NoLogIndex]
    /\ logIndex  = [n \in CN |-> NoLogIndex]
    /\ consensus = [n \in CN |-> FALSE]

Next ==
    \/ \E n \in CN : Crash(n) \/ Startup(n) \/ RecoveryTimeout(n)

Fairness ==
    WF_vars(\E n \in CN : Startup(n) \/ RecoveryTimeout(n))

Spec == Init /\ [][Next]_vars /\ Fairness

================================================================================
