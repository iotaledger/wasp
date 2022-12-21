------------------------ MODULE WaspChainAccessNodesV4 -------------------------
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

NOTE: A node has to track its logical clock for each peer independently, to
cope with byzantine nodes. They can send high LCs to overflow the receiver.
If the counters (or offsets) are tracked for each peer independently, then
overflow can only impact the communication with the byzantine node (which
makes no harm for the system as a whole). This is not modelled in this spec.
*)
EXTENDS Naturals
CONSTANT Nodes
CONSTANT Chains
CONSTANT MaxLC       \* To have model checking finite.
CONSTANT MaxReboots  \* To have liveness to pass.
ASSUME assms ==
    /\ \E n \in Nodes : TRUE
    /\ \E c \in Chains : TRUE
    /\ MaxLC \in Nat

VARIABLE active     \* Activated chains for all the nodes.
VARIABLE access     \* For each node: what nodes are considered access nodes.
VARIABLE server     \* Access nodes uses these as servers for their queries.
VARIABLE lClock     \* Logical clocks for all the nodes in each node.
VARIABLE reboots    \* Max number of remaining reboots.
VARIABLE msgs       \* Inflight messages.
vars == <<active, access, server, lClock, reboots, msgs>>

LC == 0..MaxLC                                       \* To have bounded model checking only.
LC_HaveNext(n, count) == lClock[n][n] + count \in LC \* To have bounded model checking only.

Msgs == [
    src: Nodes,            \* Sender.
    dst: Nodes,            \* Receiver.
    src_lc: LC,            \* Sender's logical clock, represents the version of the chains field.
    dst_lc: LC,            \* Last known logical clock of the destination node.
    access: SUBSET Chains, \* Access to these chains is granted by src to dst.
    server: SUBSET Chains  \* The src got this set of chains with dst_lc.
]

TypeOK ==
    /\ active  \in [Nodes -> SUBSET Chains]
    /\ access  \in [Nodes -> [Chains -> SUBSET Nodes]]
    /\ server  \in [Nodes -> [Chains -> SUBSET Nodes]]
    /\ lClock  \in [Nodes -> [Nodes -> LC]]
    /\ reboots \in 0..MaxReboots
    /\ msgs    \subseteq Msgs

--------------------------------------------------------------------------------
(*
User / node actions.
*)

\* A set of chains for which the node n has given access to m.
accessForChains(n, m) == {c \in Chains : c \in active[n] /\ m \in access[n][c]}
serverForChains(n, m) == {c \in Chains :                    m \in server[n][c]}
accessMsgs(n) == {[
        src    |-> n,
        dst    |-> dst,
        src_lc |-> lClock'[n][n],
        dst_lc |-> lClock'[n][dst],
        access |-> accessForChains(n, dst)',
        server |-> serverForChains(n, dst)'
    ] : dst \in (Nodes \ {n})}

\* We ignore the chains in access messages that we get, but don't have enabled.
\* That's to avoid a possibility to fill the memory with fake access notifications.
\* Therefore after enabling a chain we have to query for access again.
ChainActivate(n, c) ==
    /\ LC_HaveNext(n, 2)
    /\ c \notin active[n]
    /\ active' = [active EXCEPT ![n] = @ \cup {c}]
    /\ lClock' = [lClock EXCEPT ![n][n] = @+1]  \* Config has changed this way.
    /\ UNCHANGED <<access, server, reboots>>
    /\ msgs' = msgs \cup accessMsgs(n)

ChainDeactivate(n, c) ==
    /\ LC_HaveNext(n, 2)
    /\ c \in active[n]
    /\ active' = [active EXCEPT ![n] = @ \ {c}]
    /\ server' = [access EXCEPT ![n][c] = {}]   \* That's non-persistent info.
    /\ lClock' = [lClock EXCEPT ![n][n] = @+1]  \* Config has changed this way.
    /\ UNCHANGED <<access, reboots>>
    /\ msgs' = msgs \cup accessMsgs(n)

AccessNodeAdd(n, c, a) ==
    /\ LC_HaveNext(n, 2)
    /\ a \notin access[n][c]
    /\ access' = [access EXCEPT ![n][c] = @ \cup {a}]
    /\ lClock' = [lClock EXCEPT ![n][n] = @+1] \* Config has changed.
    /\ UNCHANGED <<active, server, reboots>>
    /\ \/ c \in    active[n] /\ msgs' = msgs \cup accessMsgs(n)
       \/ c \notin active[n] /\ UNCHANGED msgs

AccessNodeDel(n, c, a) ==
    /\ LC_HaveNext(n, 2)
    /\ a \in access[n][c]
    /\ access' = [access EXCEPT ![n][c] = @ \ {a}]
    /\ lClock' = [lClock EXCEPT ![n][n] = @+1] \* Config has changed.
    /\ UNCHANGED <<active, server, reboots>>
    /\ \/ c \in    active[n] /\ msgs' = msgs \cup accessMsgs(n)
       \/ c \notin active[n] /\ UNCHANGED msgs

Reboot(n) ==
    /\ reboots > 0
    /\ reboots' = reboots - 1
    /\ server' = [access EXCEPT ![n] = [c \in Chains |-> {}]]                    \* That's non-persistent info.
    /\ lClock' = [lClock EXCEPT ![n] = [m \in Nodes |-> IF n = m THEN 1 ELSE 0]] \* That's non-persistent info.
    /\ UNCHANGED <<active, access>>
    /\ msgs' = msgs \cup accessMsgs(n)

--------------------------------------------------------------------------------
(*
Handle the messages.
*)
sendMaybeAck(m, send) ==
    \* \/ msgs' = msgs \cup send \* TODO: Either use or remove.
    \/ msgs' = (msgs \ {m}) \cup send

(*
  - If we get msg.dst_lc > n.lc ----> update our local LC.
  -    additionally, if msg.servers != accessTo.., m.lc++
  - ...
  - ...
*)
RecvAccess(n) == \E m \in msgs:
    /\ m.dst = n
    /\ lClock' = [lClock EXCEPT ![n][n]     = IF @ > m.dst_lc THEN @ ELSE (IF accessForChains(n, m.src) = m.server THEN m.dst_lc ELSE m.dst_lc + 1),
                                ![n][m.src] = IF @ > m.src_lc THEN @ ELSE m.src_lc]
    /\ UNCHANGED <<active, access, reboots>>
    /\ \/ /\ m.src_lc > lClock[n][m.src] \* Update the servers, if message not outdated.
          /\ server' = [server EXCEPT ![n] = [c \in Chains |-> IF c \in m.access
                                                               THEN server[n][c] \cup {m.src}
                                                               ELSE server[n][c] \ {m.src} ]]
       \/ /\ m.src_lc =< lClock[n][m.src]
          /\ UNCHANGED <<server>>
    /\ \/ m.dst_lc < lClock[n][n]'                                         /\ sendMaybeAck(m, accessMsgs(n))
       \/ m.dst_lc = lClock[n][n]' /\ serverForChains(n, m.src) # m.access /\ sendMaybeAck(m, accessMsgs(n))
       \/ m.dst_lc = lClock[n][n]' /\ serverForChains(n, m.src) = m.access /\ sendMaybeAck(m, {})
       \/ m.dst_lc > lClock[n][n]'                                         /\ sendMaybeAck(m, {}) \* NOTE: Impossible.

--------------------------------------------------------------------------------
Init ==
    /\ active  = [n \in Nodes |-> {}]
    /\ access  = [n \in Nodes |-> [c \in Chains |-> {}]]
    /\ server  = [n \in Nodes |-> [c \in Chains |-> {}]]
    /\ lClock  = [n \in Nodes |-> [m \in Nodes |-> IF m = n THEN 1 ELSE 0]]
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
