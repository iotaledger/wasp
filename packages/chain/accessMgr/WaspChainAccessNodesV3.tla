------------------------ MODULE WaspChainAccessNodesV3 -------------------------
(*
In this specification, we model a chain's management of access nodes.
Each node can add several nodes as access nodes.
That means it will accept queries from the access nodes and act as a server.
However, the access nodes have no information from their configuration
on what nodes consider them access nodes.
A protocol described here lets the access nodes know which
nodes will act as servers (accept the queries).

  - Nodes not related to a chain should get no information on the chains it runs.
  - TODO: ...
*)
EXTENDS Naturals
CONSTANT Nodes
CONSTANT Chains
CONSTANT MaxReboots  \* To have liveness to pass.
ASSUME assms ==
    /\ \E n \in Nodes : TRUE
    /\ \E c \in Chains : TRUE
    \* /\ MaxLC \in Nat

VARIABLE active     \* Activated chains for all the nodes.
VARIABLE access     \* For each node: what nodes are considered access nodes.
VARIABLE server     \* Access nodes uses these as servers for their queries.
VARIABLE reboots    \* Max number of remaining reboots.
VARIABLE msgs       \* Inflight messages.
vars == <<active, access, server, reboots, msgs>>

Msgs == [
    t: {"ACCESS"},         \* Only a single type of a message is needed here.
    src: Nodes,            \* Sender.
    dst: Nodes,            \* Receiver.
    access: SUBSET Chains, \* Sender says the receiver is considered access node for these chains.
    server: SUBSET Chains  \* The sender's estimation, that the receiver considers it access node for these chains.
]

TypeOK ==
    /\ active  \in [Nodes -> SUBSET Chains]
    /\ access  \in [Nodes -> [Chains -> SUBSET Nodes]]
    /\ server  \in [Nodes -> [Chains -> SUBSET Nodes]]
    /\ reboots \in 0..MaxReboots
    /\ msgs    \subseteq Msgs

--------------------------------------------------------------------------------
(*
User / node actions.
*)

\* A set of chains for which the node n has given access to m.
accessForChains(n, m) == {c \in Chains : c \in active[n] /\ m \in access[n][c]}
serverForChains(n, m) == {c \in Chains : c \in active[n] /\ m \in server[n][c]}
accessMsgs(n) == {[
        t      |-> "ACCESS",
        src    |-> n,
        dst    |-> dst,
        access |-> accessForChains(n, dst)',
        server |-> serverForChains(n, dst)'
    ] : dst \in (Nodes \ {n})}

\* We ignore the chains in access messages that we get, but don't have enabled.
\* That's to avoid a possibility to fill the memory with fake access notifications.
\* Therefore after enabling a chain we have to query for access again.
ChainActivate(n, c) ==
    /\ c \notin active[n]
    /\ active' = [active EXCEPT ![n] = @ \cup {c}]
    /\ UNCHANGED <<access, server, reboots>>
    /\ msgs' = msgs \cup accessMsgs(n)

ChainDeactivate(n, c) ==
    /\ c \in active[n]
    /\ active' = [active EXCEPT ![n] = @ \ {c}]
    /\ server' = [access EXCEPT ![n][c] = {}]   \* That's non-persistent info.
    /\ UNCHANGED <<access, reboots>>
    /\ msgs' = msgs \cup accessMsgs(n)

AccessNodeAdd(n, c, a) ==
    /\ a \notin access[n][c]
    /\ access' = [access EXCEPT ![n][c] = @ \cup {a}]
    /\ UNCHANGED <<active, server, reboots>>
    /\ \/ c \in    active[n] /\ msgs' = msgs \cup accessMsgs(n)
       \/ c \notin active[n] /\ UNCHANGED msgs

AccessNodeDel(n, c, a) ==
    /\ a \in access[n][c]
    /\ access' = [access EXCEPT ![n][c] = @ \ {a}]
    /\ UNCHANGED <<active, server, reboots>>
    /\ \/ c \in    active[n] /\ msgs' = msgs \cup accessMsgs(n)
       \/ c \notin active[n] /\ UNCHANGED msgs

Reboot(n) ==
    /\ reboots > 0
    /\ reboots' = reboots - 1
    /\ server' = [access EXCEPT ![n] = [c \in Chains |-> {}]] \* That's non-persistent info.
    /\ UNCHANGED <<active, access>>
    /\ msgs' = msgs \cup accessMsgs(n)

--------------------------------------------------------------------------------
(*
Handle the messages.
*)
sendMaybeAck(m, send) ==
    \* \/ msgs' = msgs \cup send \* TODO: Either use or remove.
    \/ msgs' = (msgs \ {m}) \cup send

RecvAccess(n) == \E m \in msgs:
    /\ m.t = "ACCESS" /\ m.dst = n
    /\ UNCHANGED <<active, access, reboots>>
    /\ server' = [server EXCEPT ![n] = [c \in Chains |-> IF c \in m.access
                                                           THEN server[n][c] \cup {m.src}
                                                           ELSE server[n][c] \ {m.src} ]]
    /\ \/ m.server # accessForChains(n, m.src) /\ sendMaybeAck(m, accessMsgs(n))
       \/ m.server = accessForChains(n, m.src) /\ sendMaybeAck(m, {})

--------------------------------------------------------------------------------
Init ==
    /\ active  = [n \in Nodes |-> {}]
    /\ access  = [n \in Nodes |-> [c \in Chains |-> {}]]
    /\ server  = [n \in Nodes |-> [c \in Chains |-> {}]]
    /\ reboots = MaxReboots
    /\ msgs    = {}

Next ==
    \/ \E n \in Nodes: Reboot(n)
    \/ \E n \in Nodes, c \in Chains: ChainActivate(n, c) \/ ChainDeactivate(n, c)
    \/ \E n, a \in Nodes, c \in Chains: AccessNodeAdd(n, c, a) \/ AccessNodeDel(n, c, a)
    \/ \E n \in Nodes: RecvAccess(n)

Fairness ==
    /\ SF_vars(\E n \in Nodes: RecvAccess(n))

Spec == Init /\ [][Next]_vars /\ Fairness

--------------------------------------------------------------------------------

LinkUp(n, c, a) == c \in active[n] /\ c \in active[a] /\ a \in access[n][c]
ServerGetsKnown ==
    \A n, a \in Nodes, c \in Chains:
        (n # a /\ LinkUp(n, c, a)) ~> (n \in server[a][c] \/ ~LinkUp(n, c, a))

THEOREM Spec =>
    /\ []TypeOK
    /\ ServerGetsKnown

================================================================================
