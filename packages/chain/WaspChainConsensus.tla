-------------------------- MODULE WaspChainConsensus ---------------------------
(*
This specification models interaction of a consensus log with the L1 chain.

  - Here the "L1 chain" is considered as a chain of Alias Output TXes
    forming the chain's history/execution at L1. We consider TX rejections
    and L1 reorganizations in this spec.

  - And "Consensus Log" stands for a series of consensus round outputs, as it
    is common in the replicated state machine approach. In the case of L1 reorg
    the consensus log is not truncated in any way, it just picks other alias
    output as a basis for the next entry in the consensus log.

It is possible that:

  - A single entry in the Consensus Log gives a rise several Alias Outputs on L1.
    E.g. we have more outputs, that can fit to a TX, thus we are making several
    of them. We don't model this in the current specification.

  - Some Consensus Log Entries can produce no Alias Output's at all. E.g. there is
    no messages to execute, or there is no agreement on the alias output to
    use as a basis. Such Consensus Log Entries are still stored and maintained.

  - Alias Output indexes can proceed non-monotonously in the consensus log, e.g.
    a produced TX was rejected by L1, thus we have go back to the latest known
    unspent alias output and build on it.

See TODO marks bellow for open questions.

*)
EXTENDS Naturals, FiniteSets
CONSTANT CNode       \* A set of correct node identifiers.
CONSTANT FNode       \* A set of faulty node identifiers.
CONSTANT LogSeq      \* Consensus Log Entry positions (as in an RSM).
CONSTANT StateSeq    \* Sequence of the state indexes of the chain.
CONSTANT OutputId    \* IDs of L1 outputs.
CONSTANT MaxL1Faults \* Keep number of faults bounded to keep ability to check liveness.

IsNatPrefix(seq) == \A i \in seq : 0..i \subseteq seq

ASSUME NodeAssms ==
    /\ IsFiniteSet(CNode \cup FNode) \* Sets are finite.
    /\ CNode # {}                    \* We have at least 1 correct node.
    /\ CNode \cap FNode = {}         \* Nodes are either correct or faulty.
ASSUME LogSeqAssms ==
    IsNatPrefix(LogSeq)              \* LogSeq is always a prefix of Nat.
ASSUME StateSeqAssms ==
    IsNatPrefix(StateSeq)            \* StateSeq is always a prefix of Nat.
ASSUME OutputIdAssms ==
    \E i \in OutputId : TRUE         \* A set of Output Ids is not empty.
ASSUME L1Assms ==
    MaxL1Faults \in Nat

VARIABLE logConsensus   \* Global consensus on Base Alias Outputs to use for particular `logSN \in LogSeq'.
VARIABLE nodeLog        \* The log, as it is seen by a node, it can lag from the global `logConsensus'.
VARIABLE nodeLastOut    \* Alias Output considered latest by the node (confirmed by L1 or published by us).
VARIABLE l1Chain        \* Chain on the L1.
VARIABLE l1Spent        \* Already spent L1 outputs.
VARIABLE l1FaultsLeft   \* To limit possible number of faults and making liveness checking possible.
VARIABLE msgs           \* Already sent messages.
nodeVars == <<nodeLog, nodeLastOut>>
l1Vars == <<l1Chain, l1Spent, l1FaultsLeft>>
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

(*
A void value for OutputIds.
*)
NullOutputId == CHOOSE id : id \notin OutputId

(*
An alias output is the main building block of a chain in L1. Here we
only consider its identifier (outputId) and the chain state index it
represents (stateIdx).
*)
AliasOutput == [stateIdx: StateSeq, outputId: OutputId]
AOConflict  == [ao |-> "conflict"] \* Not enough of agreement in consensus, retry.
AONull      == [ao |-> "null"]     \* Entry was not decided yet.

(*
A set of possible messages.
*)
Msg == UNION {
    (*
    Message sent by the committee nodes to the L1 to publish new alias output (a transaction).
    *)
    [
        t: {"L1_TX_POST"},        \* Type of the message.
        aliasOutput: AliasOutput, \* Alias Output being posted.
        baseOutputId: OutputId,   \* Output ID of an alias output being spent in this TX.
        logSN: LogSeq             \* Just for checking, if a node has already sent a message.
    ],
    (*
    Messages sent by the L1 Nodes in response to the "L1_TX_POST" notifying
    TX confirmation or rejection.
    *)
    [
        t: {"L1_TX_CONFIRMED", "L1_TX_REJECTED"}, \* Type of the message.
        aliasOutput: AliasOutput, \* An output that was confirmed/rejected.
        logSN: LogSeq             \* Just for checking, if a node has already sent a message.
    ],
    (*
    A node publishes consensus proposals to vote for a base alias output (an output
    to be consumed in this TX) at particular log sequence number (logSN).
    *)
    [
        t: {"NODE_PROPOSAL"}, \* Type of the message.
        n: ANode,             \* A node which sent the proposal.
        logSN: LogSeq,        \* The proposal is for this log sequence number.
        baseAO: AliasOutput   \* The actual proposal.
    ]
}

(*
The usual type correctness invariant. Some notes:

  - An entry for the log is AONull, if it is not decided or known yet. Later
    it either set to a valid AliasOutput, or set to AOConflict, if there is no
    enough nodes voting for the same AliasOutput to use as an input for the TX.

*)
TypeOK ==
    /\ logConsensus \in [LogSeq -> AliasOutput \cup {AONull, AOConflict}]
    /\ nodeLog      \in [CNode -> [LogSeq -> AliasOutput \cup {AONull, AOConflict}]]
    /\ nodeLastOut  \in [CNode -> AliasOutput \cup {AONull}]
    /\ l1Chain      \in [StateSeq -> OutputId \cup {NullOutputId}]
    /\ l1Spent      \subseteq OutputId
    /\ l1FaultsLeft \in 0..MaxL1Faults
    /\ msgs         \subseteq Msg

--------------------------------------------------------------------------------
(*
`^\center{\textbf{L1}}^'
*)

(*
A predicate, indicating if the supplied alias output is the last one in the chain.
*)
IsL1LastUnspent(stateIdx, outputId) == \* stateIdx \in StateSeq, outputId \in OutputId
    /\ l1Chain[stateIdx] = outputId
    /\ \A idx \in StateSeq : idx > stateIdx => l1Chain[idx] = NullOutputId
    /\ outputId \notin l1Spent

(*
A set of OutputIds in the L1 Chain starting with the index `from'.
*)
L1ChainOIdsFrom(from) ==
    {oid \in OutputId : \E idx \in StateSeq : idx >= from /\ l1Chain[idx] = oid}

(*
That's partial action. Should be part of all the L1 actions that are considered faulty.
It is here to limit number of faulty steps and make liveness checking possible.
*)
L1FaultyStep ==
    /\ l1FaultsLeft > 0
    /\ l1FaultsLeft' = l1FaultsLeft - 1

(*
TX can be confirmed, if its baseOutputId is the last unspent output in the chain.
The proposal message is then consumed, to avoid re-proposals in the case of rejection or reorg.
*)
L1Confirm == \E m \in msgs: m.t = "L1_TX_POST" /\
    LET stateIdx == m.aliasOutput.stateIdx
        outputId == m.aliasOutput.outputId
    IN  /\ IsL1LastUnspent(stateIdx - 1, m.baseOutputId)
        /\ l1Chain[stateIdx - 1] = m.baseOutputId
        /\ l1Chain[stateIdx] = NullOutputId
        /\ l1Chain' = [l1Chain EXCEPT ![stateIdx] = outputId] \* New chain element added.
        /\ l1Spent' = l1Spent \cup {m.baseOutputId}           \* Previous one is consumed.
        /\ msgs' = (msgs \ {m}) \cup {[
             t           |-> "L1_TX_CONFIRMED",
             aliasOutput |-> m.aliasOutput,
             logSN       |-> m.logSN
           ]}
        /\ UNCHANGED <<logConsensus, l1FaultsLeft, nodeVars>>

(*
A TX can be rejected for any reason if it is not confirmed yet,
e.g. because of request expiry or other reason. We model this as
an environment, so don'n try to consider particular reasoning.
*)
L1Reject == L1FaultyStep /\ \E m \in msgs: m.t = "L1_TX_POST" /\
    LET stateIdx == m.aliasOutput.stateIdx
        outputId == m.aliasOutput.outputId
    IN  /\ IsL1LastUnspent(stateIdx - 1, m.baseOutputId)
        /\ l1Chain[stateIdx - 1] = m.baseOutputId
        /\ l1Chain[stateIdx] = NullOutputId
        /\ msgs' = (msgs \ {m}) \cup {[
             t           |-> "L1_TX_REJECTED",
             aliasOutput |-> m.aliasOutput,
             logSN       |-> m.logSN
           ]}
        /\ UNCHANGED <<logConsensus, nodeVars, l1Chain, l1Spent>>

(*
Reorg in L1 means our confirmed messages become unconfirmed, but the
committee remembers it has sent those transactions. The transactions in
this case are not rejected, they just disappear (i.e. other branch wins).

We keep the initial alias output stable, not a subject for reorgs. That's
because we model things in a particular chain, and reorg covering the
initial alias output would mean destruction of the entire chain.
*)
L1ReorgBack == L1FaultyStep /\ \E reorgFrom \in StateSeq :
    /\ reorgFrom > 0                     \* Keep the initial TX stable.
    /\ l1Chain[reorgFrom] # NullOutputId \* Only go back with reorg.
    /\ l1Chain' = [i \in StateSeq |-> CASE i < reorgFrom  -> l1Chain[i]
                                        [] i >= reorgFrom -> NullOutputId]
    /\ l1Spent' = l1Spent \ (L1ChainOIdsFrom(reorgFrom) \cup {l1Chain[reorgFrom-1]})
    /\ UNCHANGED <<logConsensus, nodeVars, msgs>>

(*
Additionally, other transactions might appear confirmed without
sending any messages on their confirmation. We only model up to 1
TX appearing from the "another branch" to decrease the state space.
*)
L1ReorgBranch == L1FaultyStep /\ \E reorgFrom \in StateSeq, altOutId \in OutputId :
    LET retainedSpent == l1Spent \ L1ChainOIdsFrom(reorgFrom)
    IN  /\ reorgFrom > 0                     \* Keep the initial TX stable.
        /\ l1Chain[reorgFrom] # NullOutputId \* Only go back with reorg.
        /\ altOutId \notin retainedSpent     \* Select unused OID for a TX in other branch.
        /\ l1Chain' = [i \in StateSeq |-> CASE i < reorgFrom -> l1Chain[i]
                                            [] i = reorgFrom -> altOutId
                                            [] i > reorgFrom -> NullOutputId]
        /\ l1Spent' = retainedSpent
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
(*
`^\center{\textbf{CONSENSUS}}^'

We model consensus here as a single "centralized" variable. Thus, we only use
the consensus abstraction, and don't bother with its implementation. In practice
the consensus will be implemented by the ACS or the entire HoneyBadgerBFT.
*)

(*
Predicate indicating, if the provided Log Sequence Number is the next one to fill.
*)
IsNextPendingLogIdx(logSN) == \* logSN \in LogSeq
    /\ logConsensus[logSN] = AONull
    /\ \A prev \in LogSeq : (prev < logSN) => (logConsensus[prev] # AONull)

(*
Predicate indicating, if there is enough votes casted for a specific log sequence number.
This does'n check, if there is a single value, that could be decided.
*)
HaveEnoughProposals(logSN) == \* logSN \in LogSeq
    \E q \in QNF : \A qn \in q : \E m \in msgs :
        /\ m.t = "NODE_PROPOSAL"
        /\ m.n = qn
        /\ m.logSN = logSN

(*
Predicate indicating, if `baseAO' can be decided for the specified entry in the log.
There should be N-F votes for the base Alias Output for it to be decided.
*)
CanBeDecided(logSN, baseAO) == \* logSN \in LogSeq, baseAO \in AliasOutput
    \E q \in QNF : \A qn \in q : \E m \in msgs :
        /\ m.t = "NODE_PROPOSAL"
        /\ m.n = qn
        /\ m.logSN = logSN
        /\ m.baseAO = baseAO

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

ConsensusActions == ConsensusDecision
ConsensusFairness == SF_vars(ConsensusActions)

--------------------------------------------------------------------------------
(*
`^\center{\textbf{COMMITTEE NODE}}^'

Here we model actions of the nodes in the chain's committee. The main goal
is to model node's interaction with the L1 Chain and the Consensus.
The committee has to handle rejects and reorgs in L1, agree on the next
log entry by choosing an alias output to use as an input for the next TX.

The main idea is that if the nodes consider different alias outputs to use,
after some time they will get synchronized, because L1 here the the source
of this information, and it will be synchronized to all the correct nodes
after some time.
*)

(*
Predicate indicating if `logSN' is the first pending slot in the node's `n' log.
*)
IsNextPendingNodeLogIdx(n, logSN) == \* n \in CNode, logSN \in LogSeq
    /\ nodeLog[n][logSN] = AONull
    /\ \A prev \in LogSeq : (prev < logSN) => (nodeLog[n][prev] # AONull)

(*
This predicate indicates, if `logSN' is the last decided entry in the node `n' log.
*)
IsLastDecidedNodeLogSN(n, logSN) ==
    /\ nodeLog[n][logSN] # AONull
    /\ \A next \in LogSeq : next > logSN => nodeLog[n][next] = AONull

(*
From time to time, a node receives confirmed and unspent outputs from the L1.
This action makes the nodes resynchronize, if something bad happened.
We can have several scenarios here:

  - An Alias Output was just posted to L1, it was confirmed and we got it
    here from L1. That's the main "sunny day" scenario.

  - It is possible that we have posted a TX (and stored it in nodeLastOut to
    support the pipelining), but L1 rejected it. We will get the older Alias
    Output here and will have to agree with other nodes to go back one step.

      - A race condition is possible here: it is possible that we have posted a
        TX, it was not rejected, and not yet confirmed. We will get te previous
        Alias Output as the last unspent during that time and will see it in our
        log of decisions. That makes it unclear, if we have to go back, or wait
        for other confirmations. `^\textbf{//TODO: What to do with it?}^'.

      - One possible solution is for a node to post old TX'es in the descending
        order until it will receive no rejection. This way the chain can know,
        if its local state is outdated and the one reported by L1 node is the
        correct one.

  - It is possible that we will receive newer alias output (if we are lagging
    with our consensus proposals and decisions), or other unseen alias outputs
    (in the case of reorg). In these cases, we don't see these outputs in our
    log, so we just take them and propose in the next step.

`^\textbf{//TODO: This action makes synchronization quite strict. It only
pushes the single globally known last unspent TX to all the nodes. Such an
assumption can be inappropriate in reality. Older/outdated outputs can be
pushed by some nodes, not to mentioning the byzantine ones.}^'
*)
NodeSyncFromL1 ==
    \E n \in CNode, ao \in AliasOutput :
        /\ IsL1LastUnspent(ao.stateIdx, ao.outputId)
        /\ nodeLastOut' = [nodeLastOut EXCEPT ![n] = ao]
        /\ UNCHANGED <<logConsensus, nodeLog, l1Vars, msgs>>

(*
A node proposes a base output ID for a particular log index.
It makes the proposals only for the first pending log index, thus
proposals are done in lock step with the decisions from the consensus.

That doesn't mean we wait for something from L1. In the success scenario
we will always propose `nodeLastOut[n]' as an alias output to use as a base.
This variable set only when a transaction is posted by this node, or a
confirmed output is received from L1. Thus, we mark this node as already sent
the proposal, by making `nodeLastOut[n] = AONull' to avoid repeating
consensus with the same output.
*)
NodeConsensusProposal ==
    \E n \in CNode, logSN \in LogSeq :
        /\ IsNextPendingNodeLogIdx(n, logSN) \* Consider first undecided log entry.
        /\ nodeLastOut[n] # AONull           \* Wait until we have opinion on the last Alias UTXO.
        /\ ~\E m \in msgs :                  \* That's our first proposal.
             m.t = "NODE_PROPOSAL" /\ m.n = n /\ m.logSN = logSN
        /\ msgs' = msgs \cup {[
             t      |-> "NODE_PROPOSAL",
             n      |-> n,
             logSN  |-> logSN,
             baseAO |-> nodeLastOut[n] \* Propose our currently known last unspent output.
           ]}
        /\ nodeLastOut' = [nodeLastOut EXCEPT ![n] = AONull]
        /\ UNCHANGED <<logConsensus, nodeLog, l1Vars>>


(*
Fair nodes eventually get the consensus decision.
The decisions can be received out of order.

As a result of the consensus, the node only stores its result in its
persistent memory. The actual execution is usually made separately in
the state machine replication approach. Here the execution is posting
a transaction to L1.
*)
NodeConsensusLearned ==
    \E n \in CNode, logSN \in LogSeq :
        /\ nodeLog[n][logSN] = AONull
        /\ logConsensus[logSN] # AONull
        /\ nodeLog' = [nodeLog EXCEPT ![n][logSN] = logConsensus[logSN]]
        /\ UNCHANGED <<logConsensus, nodeLastOut, l1Vars, msgs>>

(*
When a consensus is reached, a node can post the corresponding
transaction to L1 network. That's the mentioned execution of the RSM.
*)
NodePostL1ChainTx ==
    \E n \in CNode, logSN \in LogSeq, newFreeOId \in (OutputId \ l1Spent) :
        /\ IsLastDecidedNodeLogSN(n, logSN)  \* Consider latest log entry only.
        /\ nodeLog[n][logSN] \in AliasOutput \* The conflict case is handled in NodeRecoverConflict.
        /\ ~\E m \in msgs :                  \* We haven't proposed a TX yet.
             /\ m.t \in {"L1_TX_POST", "L1_TX_CONFIRMED", "L1_TX_REJECTED"}
             /\ m.logSN = logSN
        /\ LET newAO == [stateIdx |-> nodeLog[n][logSN].stateIdx + 1,
                         outputId |-> newFreeOId]
           IN  /\ newAO.stateIdx \in StateSeq             \* Just to limit the TLC state search.
               /\ newFreeOId # nodeLog[n][logSN].outputId \* Don't reuse the last unspent OId as well.
               /\ msgs' = msgs \cup {[
                    t            |-> "L1_TX_POST",
                    aliasOutput  |-> newAO,
                    baseOutputId |-> nodeLog[n][logSN].outputId,
                    logSN        |-> logSN
                  ]}
               /\ nodeLastOut' = \* Consider the posted AO as the latest unspent.
                    [nodeLastOut EXCEPT ![n] = newAO]
               /\ UNCHANGED  <<logConsensus, nodeLog, l1Vars>>

(*
If the consensus decided, that there is no agreement on a single alias output to
be used as a base for a transition, we have to do the next proposal for the
consensus. It it should be so that with each such attempt more nodes will become
synchronized and thus will propose the same base alias output with higher probability.
*)
NodeRecoverConflict ==
    NodeConsensusProposal \* We just re-propose our last known info.

(*
All the node actions.
*)
NodeActions ==
    \/ NodeSyncFromL1
    \/ NodeConsensusProposal
    \/ NodeConsensusLearned
    \/ NodePostL1ChainTx
    \/ NodeRecoverConflict

(*
All the actions should eventually happen, if enabled.
*)
NodeFairness ==
    /\ SF_vars(NodeActions)

--------------------------------------------------------------------------------
(*
`^\center{\textbf{The Specification}}^'
*)
Init == \E initId \in OutputId :
    /\ logConsensus = [ls \in LogSeq |-> AONull]
    /\ nodeLog      = [n \in CNode |-> [ls \in LogSeq |-> AONull]]
    /\ nodeLastOut  = [n \in CNode |-> AONull]
    /\ l1Chain      = [idx \in StateSeq |-> IF idx = 0 THEN initId ELSE NullOutputId]
    /\ l1Spent      = {}
    /\ l1FaultsLeft = MaxL1Faults
    /\ msgs         = {}
Next == L1Actions \/ ConsensusActions \/ NodeActions
Fairness == L1Fairness /\ ConsensusFairness /\ NodeFairness
Spec == Init /\ [][Next]_vars /\ Fairness
--------------------------------------------------------------------------------
(*
`^\center{\textbf{Properties}}^'
*)

(*
The chain on L1 should be always continuous. It can be truncated by reorgs, but then
other state indexes will be proposed in order based on the new basis.
*)
L1ChainIsContinuous ==
    \A stateIdx \in StateSeq:
        l1Chain[stateIdx] # NullOutputId =>
            \A prevIdx \in StateSeq: prevIdx < stateIdx => l1Chain[prevIdx] # NullOutputId

(*
If there is an output confirmed on the L1 chain, then there was a decision
at least on some nodes to propose a value for that state index. That's not
the case for the initial state `l1Chain[0]'.
*)
ProposalForEachConfirmedIndex ==
    \A stateIdx \in StateSeq:
        (stateIdx > 0 /\ l1Chain[stateIdx] # NullOutputId) => (
            \E n \in CNode, logSN \in LogSeq:
                /\ nodeLog[n][logSN] \in AliasOutput
                /\ nodeLog[n][logSN].stateIdx = stateIdx-1
        )

(*
It might happen that L1 will always reject our TXes, or do some reorgs. Nevertheless
we have to keep deciding on the next transaction to propose. That is, we have to fill
the consensus log on all correct nodes.
*)
NodeLogsAreFilled ==
    <> \A n \in CNode, logSN \in LogSeq : nodeLog[n][logSN] # AONull


(*
AUX: Decision on the first log entry will always be made, because we don't
consider reorgs for the initial state (that would mean deletion of a chain).
*)
AlwaysDecidesOnTheFirst ==
    <>[](
        /\ logConsensus[0] # AONull
        /\ logConsensus[0].outputId = l1Chain[0]
        /\ logConsensus[0].stateIdx = 0
    )

(*
Initial state [0] stays not reverted in this spec, so we have to agree on the next [1] TX.
The state index [1] can be reverted after being confirmed and the TX itself can be
rejected, thus we can't check if `1 \in StateSeq ~> l1Chain[1] # NullOutputId'.
*)
AlwaysProposeTheSecond ==
    <> \E m \in msgs :
         /\ m.t = "L1_TX_POST"
         /\ m.aliasOutput.stateIdx = 1
         /\ m.baseOutputId = l1Chain[0]

(*
Here we just note all the invariants and temporal properties, that were checked using TLC.
*)
THEOREM Spec =>
    /\ []TypeOK
    /\ []L1ChainIsContinuous
    /\ []ProposalForEachConfirmedIndex
    /\ NodeLogsAreFilled
    /\ AlwaysDecidesOnTheFirst
    /\ AlwaysProposeTheSecond
PROOF OMITTED \* Checked by the TLC.

================================================================================
