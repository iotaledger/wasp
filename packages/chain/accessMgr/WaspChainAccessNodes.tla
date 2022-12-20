------------------------- MODULE WaspChainAccessNodes --------------------------
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
CONSTANT Nodes
ASSUME \E n \in Nodes : TRUE

VARIABLE enabled    \* Where is the chain enabled.
VARIABLE access     \* For each node: what nodes are considered access nodes.
VARIABLE server     \* Access nodes uses these as servers for their queries.
VARIABLE msgs       \* Inflight messages.
vars == <<enabled, access, server, msgs>>

Msgs == UNION {
    [t: {"ENABLED", "DISABLED"}, src: Nodes, dst: Nodes],
        \* src enabled/disabled dst as an access node.
    [t: {"NODE_UP"}, src: Nodes, dst: Nodes]
        \* Send by a node to all the others to get information
        \* on who considers it a access node.
}

TypeOK ==
    /\ enabled \in [Nodes -> BOOLEAN]
    /\ access  \in [Nodes -> SUBSET Nodes]
    /\ server  \in [Nodes -> SUBSET Nodes]
    /\ msgs    \subseteq Msgs
--------------------------------------------------------------------------------
(*
User / node actions.
*)

ChainEnable(n) ==
    /\ ~enabled[n]
    /\ enabled' = [enabled EXCEPT ![n] = TRUE]
    /\ msgs' = msgs \cup [t: {"ENABLED"}, src: {n}, dst: access[n]]
    /\ UNCHANGED <<access, server>>

ChainDisable(n) ==
    /\ enabled[n]
    /\ enabled' = [enabled EXCEPT ![n] = FALSE]
    /\ server'  = [access EXCEPT ![n] = {}] \* That's non-persistent info.
    /\ UNCHANGED <<access, msgs>>

AccessNodeAdd(n, a) ==
    /\ a \notin access[n]
    /\ access' = [access EXCEPT ![n] = @ \cup {a}]
    /\ \/ enabled[n] /\ msgs' = msgs \cup {[t |-> "ENABLED", src |-> n, dst |-> a]}
       \/ ~enabled[n] /\ UNCHANGED msgs
    /\ UNCHANGED <<enabled, server>>

AccessNodeDel(n, a) ==
    /\ a \in access[n]
    /\ access' = [access EXCEPT ![n] = @ \ {a}]
    /\ \/ enabled[n] /\ msgs' = msgs \cup {[t |-> "DISABLED", src |-> n, dst |-> a]}
       \/ ~enabled[n] /\ UNCHANGED msgs
    /\ UNCHANGED <<enabled, server>>

Reboot(n) ==
    /\ server'  = [access EXCEPT ![n] = {}] \* That's non-persistent info.
    /\ msgs' = msgs \cup [t: {"NODE_UP"}, src: {n}, dst: Nodes]
    /\ UNCHANGED <<enabled, access>>

--------------------------------------------------------------------------------
(*
Handle the messages.
*)
MaybeAck(m) ==
    \/ UNCHANGED msgs
    \/ msgs' = msgs \ {m}
SendMaybeAck(send, m) ==
    \/ msgs' = msgs \cup send
    \/ msgs' = (msgs \ {m}) \cup send

RecvEnabled(n) == \E m \in msgs:
    /\ enabled[n] /\ m.t = "ENABLED" /\ m.dst = n
    /\ server' = [server EXCEPT ![n] = @ \cup {m.src}]
    /\ MaybeAck(m)
    /\ UNCHANGED <<enabled, access>>
RecvDisabled(n) == \E m \in msgs:
    /\ enabled[n] /\ m.t = "DISABLED" /\ m.dst = n
    /\ server' = [server EXCEPT ![n] = @ \ {m.src}]
    /\ MaybeAck(m)
    /\ UNCHANGED <<enabled, access>>
RecvNodeUp(n) == \E m \in msgs:
    /\ enabled[n] /\ m.t = "NODE_UP" /\ m.dst = n
    /\ \/ m.src \in    access[n] /\ SendMaybeAck({[t |-> "ENABLED", src |-> n, dst |-> m.src]}, m)
       \/ m.src \notin access[n] /\ MaybeAck(m)
    /\ UNCHANGED <<enabled, access, server>>

--------------------------------------------------------------------------------
Init ==
    /\ enabled = [n \in Nodes |-> FALSE]
    /\ access  = [n \in Nodes |-> {}]
    /\ server  = [n \in Nodes |-> {}]
    /\ msgs    = {}
Next ==
    \/ \E n \in Nodes: ChainEnable(n) \/ ChainDisable(n) \/ Reboot(n)
    \/ \E n, a \in Nodes: AccessNodeAdd(n, a) \/ AccessNodeDel(n, a)
    \/ \E n \in Nodes: RecvEnabled(n) \/ RecvDisabled(n) \/ RecvNodeUp(n)
Fairness ==
    /\ WF_vars(\E n \in Nodes: ChainEnable(n))
    /\ WF_vars(\E n, a \in Nodes: AccessNodeAdd(n, a))
    /\ \A n \in Nodes: WF_vars(RecvEnabled(n) \/ RecvDisabled(n) \/ RecvNodeUp(n))
Spec == Init /\ [][Next]_vars /\ Fairness
--------------------------------------------------------------------------------

LinkUp(n, a) == enabled[n] /\ enabled[a] /\ a \in access[n]
ServerGetKnown ==
    \A n, a \in Nodes: LinkUp(n, a) ~> (n \in server[a] \/ ~LinkUp(n, a))

THEOREM Spec =>
    /\ []TypeOK
    /\ ServerGetKnown

================================================================================
