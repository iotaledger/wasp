------------------------ MODULE WaspRecoveryTermination ------------------------
EXTENDS WaspByzEnv
CONSTANT MAX_LI
ASSUME MAX_LI \in Nat

VARIABLE recovered  \* Nodes proposed to recover with the specified LI.
VARIABLE consDone   \* Consensus completed with these indexes.
VARIABLE msgs       \* Sent messages.
vars == <<recovered, consDone>>

LI == 1..MAX_LI
NoLI == 0
OptLI == LI \cup {NoLI}

TypeOK ==
    /\ recovered \in [CN -> OptLI]
    /\ consDone \in TRUE \* TODO: ...


Init ==
    /\ recovered \in [CN -> OptLI]
Next == TRUE
Fair == TRUE
Spec == Init /\ [][Next]_vars /\ Fair

================================================================================
