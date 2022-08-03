-------------------------- MODULE WaspChainConsensus ---------------------------
(*
This specification models interaction of a consensus log with the L1 chain.
  - Here the "L1 chain" is considered as a chain of Alias Output TXes
    forming the chain's history/execution.
  - And "Consensus Log" stands for series of consensus round outputs, as it
    is common in the replicated state machine approach.

It is possible that:
  - Single entry in the Consensus Log gives a rise several Alias Outputs on L1.
    E.g. we have more outputs, that can fit to a TX, thus we are making several
    of them.
  - Some Consensus Entries can produce no Alias Output's at all. E.g. there is
    no messages to execute, or there is no agreement on the alias output to
    use as a basis. Such Consensus Log Entries are still stored ans maintained.
  - Alias Output indexes can proceed non-monotonously, e.g. a produced TX
    was rejected by L1, thus we have go back to the latest known unspent
    alias output and build on it.
*)
EXTENDS Naturals, FiniteSets
CONSTANT CNode     \* A set of correct node identifiers.
CONSTANT FNode     \* A set of faulty node identifiers.
CONSTANT LogSeq    \* Log Entry positions (as in an RSM).
CONSTANT StateSeq  \* Sequence of the state indexes.
CONSTANT OutputId  \* IDs of L1 outputs.

IsNatPrefix(seq) == \A i \in seq : 0..i \subseteq seq

ASSUME NodeAssms ==
    /\ IsFiniteSet(CNode \cup FNode) \* Sets are finite.
    /\ CNode # {}                    \* We have at least 1 correct node.
    /\ CNode \cap FNode = {}         \* Nodes are either correct or Faulty.
ASSUME LogSeqAssms ==
    IsNatPrefix(LogSeq)              \* LogSeq is always a prefix of Nat.
ASSUME StateSeqAssms ==
    IsNatPrefix(StateSeq)            \* StateSeq is always a prefix of Nat.
ASSUME OutputIdAssms ==
    \E i \in OutputId : TRUE         \* A set of Output Ids is not empty.

VARIABLE logConsensus   \* Global consensus on Base Alias Outputs to use for particular logSN \in LogSeq.
VARIABLE nodeLog        \* The log, as it is seen by a node.
VARIABLE nodeLastCnf    \* Last Alias Output whose confirmation was received from L1.
VARIABLE nodeLastOut    \* Alias Output considered latest by the node (confirmed by L1 or published by us).
VARIABLE l1Chain        \* Chain on the L1.
VARIABLE l1Spent        \* Already spent L1 outputs.
VARIABLE msgs           \* Already sent messages.
nodeVars == <<nodeLog, nodeLastCnf, nodeLastOut>>
l1Vars == <<l1Chain, l1Spent>>
vars == <<logConsensus, nodeVars, l1Vars, msgs>>


(*
Define node sets, quorums.
*)
ANode == CNode \cup FNode \* All Nodes / Any Node.
N == Cardinality(ANode)
F == Cardinality(FNode)
QNF == {q \in SUBSET ANode : Cardinality(q) = N-F}
QF1 == {q \in SUBSET ANode : Cardinality(q) = F+1}
ASSUME QuorumAssms == 3*F+1 <= N \* Byzantine quorum assumption.

NullOutputId == CHOOSE id : id \notin OutputId

(*
A log contains decided Chain State Index (stateIdx \in StateSeq) and
a base output to build upon (outputId \in OutputId).
*)
AliasOutput == [stateIdx: StateSeq, outputId: OutputId]
AOConflict == [log |-> "conflict"] \* Not enough of agreement in consensus, retry.
AOPending  == [log |-> "pending"]  \* Entry was not decided yet.

Msg == UNION {
    (*
    Message sent by the committee nodes to the L1 to publish new
    alias output (a transaction).
    *)
    [
        t: {"L1_TX_POST"},        \* Type of the message.
        aliasOutput: AliasOutput, \* Alias Output being posted.
        baseOutputId: OutputId,   \* Output ID of an alias output being spent.
        logSN: LogSeq             \* Just for checking, if a node has already sent a message.
    ],
    (*
    Messages sent by the L1 Nodes in response to the "L1_TX_POST" notifying
    TX confirmation or rejection. // TODO: Confirmed message is not needed.
    *)
    [
        t: {"L1_TX_CONFIRMED", "L1_TX_REJECTED"}, \* Type of the message.
        aliasOutput: AliasOutput, \* output that was confirmed/rejected.
        logSN: LogSeq             \* Just for checking, if a node has already sent a message.
    ],
    (*
    A node publishes consensus proposals to vote for base output ids (an
    output to be consumed in this TX) at particular log sequence numbers
    and particular chain state indexes (outSN). // TODO: outSN -> outSI ?
    *)
    [
        t: {"NODE_PROPOSAL"},
        n: ANode,
        logSN: LogSeq,
        baseAO: AliasOutput
    ]
}

TypeOK ==
    /\ logConsensus \in [LogSeq -> AliasOutput \cup {AOPending, AOConflict}]
    /\ nodeLog      \in [CNode -> [LogSeq -> AliasOutput \cup {AOPending, AOConflict}]]
    /\ nodeLastCnf  \in [CNode -> AliasOutput \cup {AOPending}]
    /\ nodeLastOut  \in [CNode -> AliasOutput \cup {AOPending}]
    /\ l1Chain      \in [StateSeq -> OutputId \cup {NullOutputId}]
    /\ l1Spent      \subseteq OutputId
    /\ msgs         \subseteq Msg

--------------------------------------------------------------------------------
(*                             ---=== L1 ===---                               *)

IsL1LastUnspent(stateIdx, outputId) == \* stateIdx \in StateSeq, outputId \in OutputId
    /\ l1Chain[stateIdx] = outputId
    /\ \A idx \in StateSeq : idx > stateIdx => l1Chain[idx] = NullOutputId
    /\ outputId \notin l1Spent

(*
TX can be confirmed, id its baseId output is the last unspent output.
The proposal message is then consumed.
*)
L1Confirm == \E m \in msgs: m.t = "L1_TX_POST" /\
    LET stateIdx == m.aliasOutput.stateIdx
        outputId == m.aliasOutput.outputId
    IN  /\ IsL1LastUnspent(stateIdx - 1, m.baseOutputId)
        /\ l1Chain[stateIdx - 1] = m.baseOutputId
        /\ l1Chain[stateIdx] = NullOutputId
        /\ l1Chain' = [l1Chain EXCEPT ![stateIdx] = outputId]
        /\ l1Spent' = l1Spent \cup {m.baseOutputId}
        /\ msgs' = (msgs \ {m}) \cup {[t |-> "L1_TX_CONFIRMED", aliasOutput |-> m.aliasOutput, logSN |-> m.logSN]}
        /\ UNCHANGED <<logConsensus, nodeVars>>

(*
A TX can be rejected for any reason, if it is not confirmed yet, e.g. request expiry.
*)
L1Reject == \E m \in msgs: m.t = "L1_TX_POST" /\
    LET stateIdx == m.aliasOutput.stateIdx
        outputId == m.aliasOutput.outputId
    IN  /\ IsL1LastUnspent(stateIdx - 1, m.baseOutputId)
        /\ l1Chain[stateIdx - 1] = m.baseOutputId
        /\ l1Chain[stateIdx] = NullOutputId
        /\ msgs' = (msgs \ {m}) \cup {[t |-> "L1_TX_REJECTED", aliasOutput |-> m.aliasOutput, logSN |-> m.logSN]}
        /\ UNCHANGED <<logConsensus, nodeVars, l1Vars>>

(*
Reorg in L1 means our confirmed messages become unconfirmed, but the
committee remembers it sent those transactions. The transactions in
this case are not rejected, they just disappear (i.e. other branch wins).
*)
L1ReorgBack == \E reorgFrom \in StateSeq :
    LET freedIds == {i \in OutputId : \E idx \in StateSeq : idx >= reorgFrom /\ l1Chain[idx] = i}
    IN  /\ reorgFrom > 0 \* Keep the initial TX stable.
        /\ l1Chain[reorgFrom] # NullOutputId
        /\ l1Chain' = [i \in StateSeq |-> CASE i < reorgFrom -> l1Chain[i]
                                            [] OTHER         -> NullOutputId]
        /\ l1Spent' = l1Spent \ freedIds
        /\ UNCHANGED <<logConsensus, nodeVars, msgs>>

(*
Additionally, other transactions might appear confirmed without
sending any messages on their confirmation. We only model up to 1
TX appearing from the "another branch" to decrease the state space.
*)
L1ReorgBranch == \E reorgFrom \in StateSeq, altOutId \in OutputId :
    LET freedIds == {i \in OutputId : \E idx \in StateSeq : idx >= reorgFrom /\ l1Chain[idx] = i}
    IN  /\ reorgFrom > 0 \* Keep the initial TX stable.
        /\ l1Chain[reorgFrom] # NullOutputId
        /\ altOutId \notin (l1Spent \ freedIds)
        /\ l1Chain' = [i \in StateSeq |-> CASE i < reorgFrom -> l1Chain[i]
                                            [] i = reorgFrom -> altOutId
                                            [] OTHER         -> NullOutputId ]
        /\ l1Spent' = (l1Spent \ freedIds) \cup {altOutId}
        /\ UNCHANGED <<logConsensus, nodeVars, msgs>>

(*
All the L1 Actions.
*)
L1Actions ==
    \/ L1Confirm
    \/ L1Reject
    \/ L1ReorgBack
    \/ L1ReorgBranch

(*
Fairness assumptions for the L1.
We don't assume the "bad" actions will happen infinitely.
*)
L1Fairness == SF_vars(L1Confirm)

--------------------------------------------------------------------------------
(*                          ---=== CHAIN ===---                               *)

IsNextPendingLogIdx(logSN) == \* logSN \in LogSeq
    /\ logConsensus[logSN] = AOPending
    /\ \A prev \in LogSeq : (prev < logSN) => (logConsensus[prev] # AOPending)

HaveEnoughProposals(logSN) == \* logSN \in LogSeq
    \E q \in QNF : \A qn \in q : \E m \in msgs :
        /\ m.t = "NODE_PROPOSAL"
        /\ m.n = qn
        /\ m.logSN = logSN

CanBeDecided(logSN, baseAO) == \* logSN \in LogSeq, baseAO \in AliasOutput
    \E q \in QNF : \A qn \in q : \E m \in msgs :
        /\ m.t = "NODE_PROPOSAL"
        /\ m.n = qn
        /\ m.logSN = logSN
        /\ m.baseAO = baseAO

(*
From time to time, a node receives confirmed and unspent outputs from the L1.
*)
NodeSyncAliasOutputs ==
    \E ao \in AliasOutput, n \in CNode:
        /\ IsL1LastUnspent(ao.stateIdx, ao.outputId)
        /\ nodeLastCnf' = [nodeLastCnf EXCEPT ![n] = ao]
        /\ nodeLastOut' = [nodeLastOut EXCEPT ![n] = ao] \* // TODO: Consider retaining the pending outputs.
        /\ UNCHANGED <<logConsensus, nodeLog, l1Vars, msgs>>

(*
A node proposes a base output ID for a particular log index.
*)
NodeProposal ==
    \E logSN \in LogSeq, n \in CNode :
        /\ IsNextPendingLogIdx(logSN)
        /\ nodeLastOut[n] # AOPending
        /\ ~\E m \in msgs : m.t = "NODE_PROPOSAL" /\ m.n = n /\ m.logSN = logSN
        /\ msgs' = msgs \cup {[
             t |-> "NODE_PROPOSAL",
             n |-> n,
             logSN |-> logSN,
             baseAO |-> nodeLastOut[n] \* Propose our currently known last output.
           ]}
        /\ UNCHANGED <<logConsensus, nodeVars, l1Vars>>

(*
A consensus on a value is reached, when it is proposed by N-F nodes.
We con't model the consensus algorithm, so here is a simple central variable for that.
If there is no single value proposed by N-F nodes, then the consensus decides on Null.
*)
ConsensusDecision ==
    \E logSN \in LogSeq:
        /\ IsNextPendingLogIdx(logSN)
        /\ \/ \E baseAO \in AliasOutput:
                /\ CanBeDecided(logSN, baseAO)
                /\ logConsensus' = [logConsensus EXCEPT ![logSN] = baseAO]
           \/ /\ HaveEnoughProposals(logSN)
              /\ \A ao \in AliasOutput: ~CanBeDecided(logSN, ao)
              /\ logConsensus' = [logConsensus EXCEPT ![logSN] = AOConflict]
        /\ UNCHANGED <<nodeVars, l1Vars, msgs>>

(*
Fair nodes eventually get the consensus decision.
*)
ConsensusOutput ==
    \E n \in CNode, logSN \in LogSeq :
        /\ nodeLog[n][logSN] = AOPending
        /\ logConsensus[logSN] # AOPending
        /\ nodeLog' = [nodeLog EXCEPT ![n][logSN] = logConsensus[logSN]]
        /\ UNCHANGED <<logConsensus, nodeLastCnf, nodeLastOut, l1Vars, msgs>>

(*
When a consensus is reached, a node can post the corresponding
transaction to L1 network.
*)
NodeTxPost ==
    \E n \in CNode, logSN \in LogSeq, newFreeOId \in (OutputId \ l1Spent) :
        /\ nodeLog[n][logSN] \in AliasOutput \* The conflict case is handled in NodeRecoverConflict.
        /\ \A next \in LogSeq : next > logSN => nodeLog[n][next] = AOPending
        /\ ~\E m \in msgs : m.t = "L1_TX_POST" /\ m.logSN = logSN                            \* By any node.
        /\ ~\E m \in msgs : m.t \in {"L1_TX_CONFIRMED", "L1_TX_REJECTED"} /\ m.logSN = logSN \* By any node.
        /\ LET newAO == [stateIdx |-> nodeLog[n][logSN].stateIdx + 1, outputId |-> newFreeOId]
           IN  /\ newAO.stateIdx \in StateSeq \* To limit TLC.
               /\ msgs' = msgs \cup {[
                    t            |-> "L1_TX_POST",
                    aliasOutput  |-> newAO,
                    baseOutputId |-> nodeLog[n][logSN].outputId,
                    logSN        |-> logSN
                  ]}
               /\ nodeLastOut' = [nodeLastOut EXCEPT ![n] = newAO] \* NOTE
               /\ UNCHANGED  <<logConsensus, nodeLog, nodeLastCnf, l1Vars>>

(*
If the consensus decided, that there is no agreement on a single alias output to
be used as a base for a transition, we have to do the next proposal.
*)
NodeRecoverConflict ==
    NodeProposal \* We just re-propose our last known info.

ChainActions ==
    \/ NodeSyncAliasOutputs
    \/ NodeProposal
    \/ ConsensusDecision
    \/ ConsensusOutput
    \/ NodeTxPost
    \/ NodeRecoverConflict

ChainFairness ==
    /\ WF_vars(ChainActions)

--------------------------------------------------------------------------------
Init == \E initId \in OutputId :
    /\ logConsensus = [ls \in LogSeq |-> AOPending]
    /\ nodeLog      = [n \in CNode |-> [ls \in LogSeq |-> AOPending]]
    /\ nodeLastCnf  = [n \in CNode |-> AOPending]
    /\ nodeLastOut  = [n \in CNode |-> AOPending]
    /\ l1Chain      = [idx \in StateSeq |-> IF idx = 0 THEN initId ELSE NullOutputId]
    /\ l1Spent      = {}
    /\ msgs         = {}
Next == L1Actions \/ ChainActions
Fairness == L1Fairness /\ ChainFairness
Spec == Init /\ [][Next]_vars /\ Fairness
--------------------------------------------------------------------------------

\*
\* TODO: All log is eventually decided.
\* TODO: L1 Decided Always In Sequence.
\*
NodeLogsAreFilled ==
    <> \A n \in CNode, logSN \in LogSeq : nodeLog[n][logSN] # AOPending

THEOREM Spec =>
    /\ []TypeOK
    /\ NodeLogsAreFilled
PROOF OMITTED \* Checked by the TLC.

================================================================================
