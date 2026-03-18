#set document(
  title: [Distributed Algorithms in ISC]
)
#show link: set text(blue)
#show figure: set place(
  clearance: 5em,
)
#set heading(numbering: "1.")

// all links to code pointing to some commit hash, to avoid broken links in the future when the code changes.
#let codelink(dest, body) = link("https://github.com/iotaledger/wasp/blob/baa9674e79aa0e5cc7cfe5a1260cb7c4cb2f86ab" + dest, body)

#title()

A core aspect of how ISC works is through the use of distributed algorithms
(i.e. algorithms that run on multiple nodes in a network and require
communication between them). These algorithms (also sometimes called
_protocols_) are used for various purposes, such as reaching consensus on the
state of the system, managing access to resources, and synchronizing data
between nodes.

A recommended reading to understand the design and implementation of distributed
algorithms in general is the book _Fault-tolerant message-passing distributed
systems_ by Michel Raynal @Raynal2018-ws. It is not necessary to read the whole
book, but it is recommended to read at least the first few chapters to
understand the basics of distributed algorithms and the different types of
well-known algorithms that exist.

#outline()

= GPA

The distributed algorithms in ISC are implemented using a framework called GPA,
which stands for "General Purpose Algorithm". GPA provides a common interface
and structure for implementing distributed algorithms in ISC. It has the
following design principles:

- *Pseudo-functional:* Message-passing is implemented as methods that receive
  some `input` (from the same node) or `message` (from some other node), and
  return some messages (if any) to be sent to other nodes. It is not really
  functional, because the state of the algorithm is stored in the struct that
  implements the GPA interface, and is mutated by the methods. However, it is
  functional in spirit because the methods do not have side effects on other parts
  of the system, and only mutate the state of the algorithm itself.

- *Single-threaded:* Each GPA instance and all its subcomponents (if any) run in
  the same goroutine as the caller, and all `input` and `message` calls are
  blocking. This eliminates the need to worry about concurrency issues. In
  practice this makes the code more complex, because Go is designed for concurrent
  programming, and it is often more natural to write code that uses goroutines and
  channels.

- *Common testing framework:* All GPA implementations can be tested using the
  same testing tool, called
  #codelink("/packages/gpa/test_context.go")[`TestContext`].

At any point in time there may be several GPA instances running on the same
node, and they can be organized in a hierarchical structure. All nodes have the
same hierarchical structure of GPA instances, and each GPA instance communicates
with the corresponding GPA instance on other nodes. @gpas shows an example of
this behavior.

// https://viewer.diagrams.net/?tags=%7B%7D&lightbox=1&highlight=0000ff&edit=_blank&layers=1&nav=1&dark=auto#R%3Cmxfile%3E%3Cdiagram%20name%3D%22Page-1%22%20id%3D%22t0HHPHxeYJFIHd6IrH3T%22%3E7Vtdc6IwFP01PnaHJKD4aLXtPuzudMaHbR8zkkK6SJwQ%2FOiv3yBBBKy1JYG2%2BtJJTpIL3HO4NzfSHhrP13ccL4LfzCNhD1reuocmPQhd25V%2FU2CTAc5gmAE%2Bp14GgQKY0heiQEuhCfVIXJooGAsFXZTBGYsiMhMlDHPOVuVpTywsX3WBfVIDpjMc1tG%2F1BOBeizHKvCfhPpBfmVgqZE5zicrIA6wx1Z7ELrpoTFnTGSt%2BXpMwtR3uV%2BydbevjO5ujJNInLJg%2FjAM715WVxMh7uN4%2BWSPH5OrA1YUFItN7gPOksgjqRmrh65XARVkusCzdHQlSZdYIOah7AHZXBIuqPTfKKR%2BJDHB0gmZzSUOE2XzjxSKREZqQC4i673rq0e4I2xOBN%2FIKcGel1GujlVBCUQKy83Yqq%2BEB3NisFKEv7NdOE02lN%2Fe4UMns0u8mpLqTmUJn5EjtmDN%2BanZqeoyLgLmswiHNwV6XaanmPOLpa7fkvJMhNiolwsngpUpk3fJNw9q%2FbbzmHZ%2BOHl3st4fnGx2PW%2BUvmKyG7GIZMgtTb23HVevK%2BY%2BEUeeGR0mnJMQC7os%2B1Q7ef0LeY3Is7skD2qPXrU4JfNKTOIkBh8LVPU45VbClFsJU7apMIXMO6sH%2B6E0fe3RpWz6aXMyneaovOu9AUP%2BhE7Fn64pf9rm%2FTkaT1tyE4Km3DTQGGCHZxhggXWY73YirHthrxl7r7yt7bA3bC8%2FwpbyIzIWz4Fl3lufMEHaxiI%2F0F9cdpgh7b4xP%2Bnfxr6vCP%2Bg1E4owu3WinCANGYq4Jxjqup3maqAfeGvIX%2BDTvlz2ttrmKrF7dZqcdA%2Fi72G01oxDgZfea9R9ZO5ahzoLOh2v2ycU5h9jfGWwuzwwl9D%2FmCX%2FMEWikzTJXk1TZoryY%2BJ%2BxunSXMlOdRfanaYJs2V5FBnNQm7iLIejoPt%2Brxzj4UgPD0MkFsAy5VoLDj7R8YsZFyiHnnCyVbzsYyhIg%2B1sxDHMZ3lsIq34PR4mxUGncVbnWUl%2Bp5EkjUVD3vtLB1DR3WLdJx2imx8OIvDt9L4qbJpehqhlt4zKmkuTvUGlYqk75RNZCJQqwrx1Qw5VsXQsGIoe8Kaoa2Kd8%2FTQNg6vzqxL8JuT9hNj2kO6xFVTlqRpUnYCLUsbJ1f5IBOKpwzVXbjyvhEZSP3g8ruv2HItLJ1fgoB6tv2i7KNKbvpmcFpyrZ1KbtmyLSyv%2FxnIl0J%2B7iuP%2FcWuHYOpG0LLLvFd%2FHZ9OKfC9DNfw%3D%3D%3C%2Fdiagram%3E%3C%2Fmxfile%3E
#figure(
  [ #image("gpa.png", width:80%) ],
  caption: [Example: many GPA instances running in two nodes.],
  placement: auto,
) <gpas>

@gpalist shows all implementations of the GPA interface in ISC.

#figure(
  rect[
    #set align(left)
    #set list(marker: [--])

    - #codelink("/packages/chains/chains.go")[`Chains`] (container for all chains managed by the node)
      - #codelink("/packages/chains/accessmanager/access_manager.go")[`AccessMgr`] (which nodes has access to which chain? who can access you? inform other nodes which chains you run, share information to access nodes)
        - #codelink("/packages/chains/accessmanager/dist/access_manager_dist.go")[`GPA:AccessMgr`]
      - #codelink("/packages/chain/node.go")[`chainNodeImpl`] (one for each chain managed by the node -- implements `chain.Chain` -- main container for all objects & logic needed to run a chain)
        - #codelink("/packages/chain/chainmanager/chain_manager.go")[`GPA:ChainMgr`]
          - #codelink("/packages/chain/committeelog/cmt_log.go")[`GPA:CommitteeLog`] (one for each committee the node participates in -- node might be lagging or ahead; for a node to be in sync decide which Consensus instances to work on, and propose the base anchor)
        - #codelink("/packages/chain/statemanager/state_manager.go")[`StateMgr`]
          - #codelink("/packages/chain/statemanager/gpa/state_manager_gpa.go")[`GPA:StateMgr`] (make sure that the node has all the blocks it needs e.g. after reboot)
        - #codelink("/packages/chain/mempool/mempool.go")[`Mempool`]
          - #codelink("/packages/chain/mempool/distsync/dist_sync.go")[`GPA:Mempool`] (share requests between nodes)
        - #codelink("/packages/chain/consensus/consensusrunner/gr.go")[`consensusrunner`] (one for each currently running Consensus instance)
          - #codelink("/packages/chain/consensus/cons.go")[`GPA:Consensus`]
            - #codelink("/packages/chain/distsign/dss.go")[`GPA:DSS`] (DSS / Distributed Schnorr Signature - signs the produced block)
              - #codelink("/packages/gpa/asyncdistkeygen/nonce/nonce.go")[`GPA:NonceDKG`] (Nonce distributed key generation)
                - #codelink("/packages/gpa/acss/acss.go")[`GPA:ACSS`] (Asynchronous Complete Secret Sharing -- one for each node in the committee)
                  - #codelink("/packages/gpa/rbc/bracha/bracha.go")[`GPA:RBC`] (Reliable Broadcast -- Bracha's algorithm)
            - #codelink("/packages/gpa/acs/acs.go")[`GPA:ACS`] (Asynchronous Common Subset -- from the Honey Badger paper)
              - #codelink("/packages/gpa/rbc/bracha/bracha.go")[`GPA:RBC`] (Reliable Broadcast -- Bracha's algorithm -- one for each node in the committee)
              - #codelink("/packages/gpa/aba/mostefaoui/mostefaoui.go")[`GPA:ABA`] (Asynchronous Binary Agreement -- Mostefaoui's algorithm -- one for each node in the committee)
                - #codelink("/packages/gpa/cc/blssig/blssig.go")[`GPA:CC`] (Common Coin based on a BLS Threshold signatures -- one for each round)
  ],
  caption: [Hierarchy of GPA implementations in ISC. Some of the listed components are not GPA implementations,
  but rather "containers" or "managers". Actual GPA implementations are marked with `GPA:`.],
  placement: auto,
) <gpalist>

== The GPA interface

All GPA instances are objects that implement the GPA interface, which
models any distributed algorithm that:

- may accept inputs, and in consequence send some messages to peers
- may receive messages, and in consequence send some messages to peers
- will eventually produce some output

The difference between an "input" and a "message" is that the former is
generated by the same node (e.g. a timer, or a request from another component),
while the latter is generated by other nodes in the committee.

```go
type GPA interface {
	// must be called to pass some input to the algorithm
	Input(inp Input) OutMessages
	// must be called whenever a message is received from a peer
	Message(msg Message) OutMessages
	// returns the output of the algorithm, if it has produced one
	Output() Output
}
```

== `AckHandler` and `OwnHandler`

TODO

== Alternative implementation

There is an experimental alternative implementation of some of the distributed
algorithms called
#link("https://github.com/dessaya/wasp/tree/actors/packages/actors")[actors],
in which the single-threaded and pseudo-functional design principles of GPA are
lifted, and the algorithms are implemented using Go's concurrency primitives
such as goroutines and channels. As a result, the code tends to be more
straightforward and easier to read. The Consensus protocol and all its
subcomponents are implemented using actors, although it has not been thoroughly
tested and so it is currently unsuitable for production use.

= `Chains` and `chainNodeImpl`

The `Chains` component is the main container for all chains managed by the node.
It is responsible for creating and managing the lifecycle of chains, and for
providing access to them. It contains:

- `AccessMgr`: responsible for managing access to chains
- one `chainNodeImpl` for each chain managed by the node, which contains:
  - `ChainMgr`, responsible for managing the lifecycle of the chain
  - `StateMgr`, responsible for managing the chain state (aka the DB)
  - `Mempool`, responsible for managing the chain requests
  - one `Consensus` instance for each currently running Consensus instance

= `AccessMgr`

This component is responsible for keeping track of trusted nodes and which nodes
have access to which chains. There is exactly one `AccessMgr` per Wasp node (as
opposed to one per chain), and it is owned by `Chains`.

Each node can designate other nodes as *access nodes* for the chains it manages.
However, the access nodes have no way of knowing from their own configuration
alone which nodes consider them access nodes. The purpose of `AccessMgr` is to
solve this problem: it runs a distributed protocol that informs each access node
which nodes will act as *servers* for it.

The design goals are:

- Nodes not related to a chain should get no information about it.
- The list of server nodes should be _transient_ (not persisted) to avoid state
  desynchronization.
- The algorithm must work in an asynchronous setting.

== Architecture

The #codelink("/packages/chains/accessmanager/access_manager.go")[`AccessMgr`]
struct is the outer wrapper. It runs a single goroutine
(#codelink("/packages/chains/accessmanager/access_manager.go#L119")[`run`]) that
processes events from several
channels#footnote[#codelink("/packages/util/pipe/pipe.go")[Pipes] are used
instead of channels, to avoid blocking goroutines.]:

- The set of trusted peers changes
- The access node list for a particular chain is updated
- A chain is deactivated on this node

Each of these translates into a GPA `Input` that is fed into the inner
`GPA:AccessMgr` instance. Network messages from peers are similarly received
via a channel and passed as GPA `Message` calls. Outgoing messages produced by the
GPA are serialized and sent over the peering network.

== `AccessMgr` GPA

The inner GPA component is implemented in
#codelink("/packages/chains/accessmanager/dist/access_manager_dist.go")[`access_manager_dist.go`]
and has a formal TLA+ specification in
#codelink("/packages/chains/accessmanager/dist/WaspChainAccessNodesV4.tla")[`WaspChainAccessNodesV4.tla`].

=== State

The GPA maintains two maps:

- *Per-node state*
  (#codelink("/packages/chains/accessmanager/dist/access_manager_dist.go#L35")[`accessMgrNode`]):
  for each trusted peer, tracks:
  - `accessFor`: the set of chains for which _we_ have granted access to that
    peer (i.e. we act as server).
  - `serverFor`: the set of chains for which _the peer_ has granted access to us
    (i.e. they act as server).
  - `ourLC` / `peerLC`: logical clocks used to detect stale messages and drive
    convergence.

- *Per-chain state*
  (#codelink("/packages/chains/accessmanager/dist/access_manager_dist.go#L36")[`accessMgrChain`]):
  for each active chain, tracks:
  - `access`: which nodes have been granted access.
  - `server`: which nodes have confirmed they will serve us (i.e. the nodes that
    consider _us_ an access node for this chain).

=== Inputs

The GPA accepts three kinds of inputs, corresponding to TLA+ spec actions:

+ *`inputAccessNodes`*
  (#codelink("/packages/chains/accessmanager/dist/input_access_nodes.go")[`input_access_nodes.go`]):
  the access node list for a chain has been updated. The first reception of this
  input for a chain implicitly _activates_ the chain (corresponding to
  `ChainActivate` in the spec). Subsequent receptions correspond to
  `AccessNodeAdd` / `AccessNodeDel`. For each trusted peer, a
  #codelink("/packages/chains/accessmanager/dist/msg_access.go")[`msgAccess`]
  message is sent with the updated access grants.

+ *`inputChainDisabled`*
  (#codelink("/packages/chains/accessmanager/dist/input_chain_disabled.go")[`input_chain_disabled.go`]):
  the chain has been deactivated. All peers are notified that access is revoked
  (`ChainDeactivate` in the spec).

+ *`inputTrustedNodes`*
  (#codelink("/packages/chains/accessmanager/dist/input_trusted_nodes.go")[`input_trusted_nodes.go`]):
  the set of trusted peers has changed. New peers are initialized and sent the
  current access state; removed peers are disconnected and their server status is
  cleared. This corresponds to `Reboot` in the spec.

=== Message exchange

All peer-to-peer communication uses a single message type,
#codelink("/packages/chains/accessmanager/dist/msg_access.go")[`msgAccess`],
which carries four fields:

- `senderLClock`: the sender's logical clock (version of its access grants).
- `receiverLClock`: the last known logical clock of the receiver (acts as an
  acknowledgment).
- `accessForChains`: the set of chains for which the sender grants access to the
  receiver.
- `serverForChains`: the set of chains for which the sender believes the receiver
  grants access to the sender (echo / ack).

=== Convergence

The protocol converges through a simple echo/ack mechanism driven by logical
clocks (#codelink("/packages/chains/accessmanager/dist/access_manager_dist.go#L337")[`handleMsgAccess`]).
When a node receives a message, it checks whether:

+ The peer's access information matches what we already know (`serverFor`).
+ The echoed server information matches our current access grants (`accessFor`).
+ The logical clocks are consistent (not outdated and not exceeding ours).

If all conditions hold, the state is converged for this peer and no reply is
needed. Otherwise, a reply is sent with the node's current state, incrementing
the logical clock if necessary. This ensures that after all in-flight messages
are delivered, both sides agree on who serves whom. The liveness property proved
in the TLA+ spec (`ServerGetsKnown`) states that if node _n_ has granted access
to node _a_ for chain _c_, and both have the chain active, then eventually _a_
will learn that _n_ is a server for it.

When the set of server nodes for a chain changes, a callback
(#codelink("/packages/chains/accessmanager/dist/access_manager_dist.go#L38")[`serversUpdatedCB`])
is invoked, which propagates the updated server list to the rest of the system
(e.g. so the node knows where to send its queries). Note that the use of
callbacks by the GPA component violates the pseudo-functional design principle
of the GPA framework.

= `ChainMgr`

TODO

== `CommitteeLog`

TODO

= `Consensus`

In ISC, the "Consensus" is the protocol that the nodes in a committee run to
agree on the next block to be added to the chain.

The ISC Consensus is loosely based on the Honey Badger BFT protocol
#cite(label("10.1145/2976749.2978399")). It is highly recommended to read the
original paper, although we will describe the main ideas and components here.

Some further documentation about ISC Consensus can also be found in
#link("https://iotafoundation.slack.com/files/U013DPD51RV/F07NZ7J651U/isc__consensus__draft__v4__-_hackmd.pdf")[this
draft document] by Karolis Petrauskas, the main author of the Consensus code.
Even if the document corresponds to a previous version of the ISC protocol, most
of the ideas and components are still relevant.

In the Honey Badger paper, it is stated that it is an "asynchronous BFT"
protocol. BFT stands for
#link("https://en.wikipedia.org/wiki/Byzantine_fault")[Byzantine Fault
Tolerant], which means that the protocol is resilient to a certain number `f` of
faulty nodes in the committee of `n` nodes. Asynchronous means that the protocol
guarantees liveness without making any timing assumptions (i.e. network delays,
ordering of messages, etc).

@cons shows a high-level view of the ISC Consensus protocol. The
#codelink("/packages/chain/consensus/consensusrunner/gr.go")[`consensusrunner`]
component is the main orchestrator that communicates with other components shown
in blue (State Manager, Mempool, etc.), and controls the
#codelink("/packages/chain/consensus/cons.go")[Consensus
GPA] instance.

#let nmark(body) = circle(fill: red, stroke: none, width: 0.5cm, height: 0.5cm)[ #align(center+horizon)[ #text(fill: white)[ #body ]]]

#let mark(body, dx, dy) = place(top+left, dx: dx, dy: dy)[ #nmark[ #body ]]

#figure(
  box(width: 100%)[
    #image("WaspConsensusInstance.drawio.png", width: 100%)
    #mark([1], 7cm, 0.5cm)
    #mark([2], 10.5cm, 6.5cm)
    #mark([3], 7.5cm, 6.5cm)
    #mark([4], 4.5cm, 5cm)
    #mark([5], 13.5cm, 5.5cm)
    #mark([6], 7.5cm, 9cm)
    #mark([7], 4.5cm, 12.5cm)
    #mark([8], 10cm, 12.5cm)
    #mark([9], 7.5cm, 15cm)
    #mark([10], 3cm, 17cm)
    #mark([11], 12.5cm, 17cm)
    #mark([12], 7.5cm, 19cm)
  ],
  caption: [High-level view of the consensus protocol],
  placement: auto,
) <cons>

A Consensus instance (with its corresponding `consensusrunner`) is spawned
whenever the `ChainMgr` determines that a new block needs to be computed and
added to the chain. In order to compute the new block we need:

- The *base L1 anchor*#footnote[Instead of "anchor",
  some old documentation may mention an "Alias Output" or "AO", which corresponds to
  a previous version of ISC.] (i.e. the latest chain state)
- The current *timestamp*
- The latest *L1 parameters* such as the base coin address, reference gas price, etc.
- A *list of requests* to execute (sometimes called a "batch")
- A way to *sign* the produced block, so that it can be anchored on L1

#let mark(body) = box[#nmark[ #body ]]

#mark[1] When spawning the Consensus, the `ChainMgr` supplies the *proposed base
anchor* as the first input to the Consensus GPA, although after the Consensus is
executed, the *decided base anchor*
#codelink("/packages/chain/consensus/batchproposal/batch_proposal_set.go#L39")[may
be different] if not every node in the committee proposes the same base anchor.

#mark[2] The `ChainMgr` also
#codelink("/packages/chain/consensus/consensusrunner/gr.go#L310")[provides]
the *timestamp* as an input to the Consensus GPA.

#mark[3] In order to obtain the L1 parameters, the `consensusrunner`
#codelink("/packages/chain/consensus/consensusrunner/gr.go#L441")[asks]
the `NodeConnection` component to provide it. When this information is ready, it
is provided as an input to the Consensus GPA.

#mark[4] In similar fashion, the `consensusrunner`:
- asks the `StateMgr` to provide the chain state corresponding to the proposed
  base anchor, and when it is ready, it is provided as an input to the Consensus
  GPA.
- asks the `Mempool` to provide a list of proposed request IDs, and when it is
  ready, it is provided as an input to the Consensus GPA.

#mark[5] While the Consensus inputs are being fetched, the Consensus GPA
#codelink("/packages/chain/consensus/cons.go#L195")[starts]
an instance of the #link(<dss>)[`DSS` protocol], which is ultimately responsible
for *signing* the produced block. The DSS protocol is designed so that `n - f`
honest nodes are enough to sign the block. Each node proposes a set of nodes
(shown as "index proposal" in @cons) and after collecting the proposals from all
peers, the DSS protocol will decide on an aggregated set of nodes that will sign
the block.

#mark[6] With all needed inputs in place, an
#link(<acs>)[ACS protocol is started]. ACS stands for "Asynchronous Common
Subset", and it is the main component of the ISC Consensus protocol. It is
responsible for taking all the inputs from the different nodes in the committee,
and deciding (agreeing) on a common subset of them, so that every honest node in
the committee will end up with the same subset of inputs.

The output of ACS will be either a valid
#codelink("/packages/chain/consensus/batchproposal/aggregated_batch_proposals.go")[aggregated
batch] to compute a new block, or a *skip* action meaning that no block will be
produced (e.g. if there are no valid requests). In case of a decided skip
action, no further steps are needed by the Consensus.

#mark[7] Given the decided base anchor and request refs, the `consensusrunner`:
- asks the `StateMgr` to provide the chain state corresponding to the decided
  base anchor
- asks the `Mempool` to provide the requests corresponding to the
  decided request refs

#mark[8] In order to compute the new block, an *entropy* or *randomness* value
is needed. The entropy value is a `[]byte` that is used as a source of
randomness for the block computation, and it is also included in the block
itself. The entropy value is
#codelink("/packages/chain/consensus/cons.go#L521")[computed]
as the BLS signature of the base anchor's reference. This way, the entropy value
is unique to each anchor version, and unpredictable by any party that is not
part of the committee.

#mark[9] From all the gathered inputs, the Consensus
#codelink("/packages/chain/consensus/cons.go#L570")[creates]
a `VMTask` that is
#codelink("/packages/chain/consensus/consensusrunner/gr.go#L446")[used]
to run a VM instance. The output of the VM is either a valid *block*, or a
*skip* action. In case of a decided skip action, no further steps are needed by
the Consensus.

#mark[10] As a result of the VM run there is a new chain state that needs to be
persisted in the DB. The `StateMgr` is
#codelink("/packages/chain/consensus/consensusrunner/gr.go#L437")[asked]
to perform this action, and when it is done, the new state is ready to be
anchored on L1.

#mark[11] The produced block is wrapped into a *L1 transaction* that needs to be
*signed* by the committee. The TX bytes is
#codelink("/packages/chain/consensus/cons.go#L607")[passed]
as an input to `DSS`, along with the sets of nodes proposed by each peer.

#mark[12]
#codelink("/packages/chain/consensus/cons.go#L644")[Finally],
the TX is signed and the Consensus is finished.

== `DSS` <dss>

The
#codelink("/packages/chain/distsign/dss.go")[`DSS`
(Distributed Schnorr Signature) protocol] is responsible for signing the L1
transaction that anchors the new block. It is designed so that the honest nodes
in the committee can sign the block even in the presence of `f` faulty nodes.
The signing algorithm is described in #cite(label("10.5555/646038.678297")).

To generate a distributed signature, the protocol requires one longterm
distributed secret (generated as part of the chain creation process),
and one random secret to be used only once (generated for each signing operation).
Each participant then issues a partial signature that can be broadcast to
the whole group. Once one has collected enough
partial signatures, it is possible to compute the distributed signature.

1. #codelink("/packages/chain/distsign/dss.go#L89")[Start a `NonceDKG`]
   GPA instance, in order to generate the distributed random secret.
   See @noncedkg for more information about this protocol.
   `NonceDKG` will produce an intermediate output (a list of proposed honest nodes),
   and a final output (the random secret to be used for DSS).

2. When the `NonceDKG` produces the intermediate output (the list of proposed nodes),
   forward it as an
   #codelink("/packages/chain/distsign/dss.go#L153")[intermediate
   output of `DSS`].

3. #codelink("/packages/chain/distsign/dss.go#L250")[Wait
   for input] containing:
   - sets of nodes proposed by each peer
   - TX bytes to be signed,
   Then forward the sets of nodes as an input to `NonceDKG`.

4. When `NonceDKG`
   #codelink("/packages/chain/distsign/dss.go#L156")[produces]
   its final output (the random secret), create a partial signature of the TX
   bytes, and broadcast it to all peers.

5. When
   #codelink("/packages/chain/distsign/dss.go#L237")[enough
   partial signatures are collected], compute the final distributed signature,
   and output it as the result of `DSS`.

== `NonceDKG` <noncedkg>

#codelink("/packages/gpa/asyncdistkeygen/nonce/nonce.go")[`NonceDKG`]
is a distributed key generation protocol that is used to generate the
random secret needed for the `DSS` protocol. `NonceDKG` is
adapted from #cite(label("cryptoeprint:2021/1591")).

1. Each node #codelink("/packages/gpa/asyncdistkeygen/nonce/nonce.go#L127")[generates]
   a random secret value and broadcasts it to all peers using the
   #link(<acss>)[`ACSS` protocol] (which guarantees that only honest nodes can
   recover the shared secrets). Thus, each `NonceDKG` instance manages `n`
   instances of `ACSS`, one for broadcasting, and `n - 1` for receiving the
   secrets from peers.

2. As soon as the secrets are
   #codelink("/packages/gpa/asyncdistkeygen/nonce/nonce.go#L189")[received
   from `n - f` peers], these peers are proposed as the set of nodes that
   will sign the block. This is an intermediate output of `NonceDKG`.

3. Wait for input containing the sets of nodes proposed by each peer. Then
   #codelink("/packages/gpa/asyncdistkeygen/nonce/nonce.go#L270")[produce
   the final output], containing:
   - an aggregated set of nodes
   - the random secret to be used for `DSS` (computed as the sum of the secrets
     proposed by the nodes in the aggregated set)

== `ACSS` <acss>

#codelink("/packages/gpa/acss/acss.go")[`ACSS`]
is an implementation of the Asynchronous Complete Secret Sharing protocol, which
is used as a building block for the `NonceDKG` protocol.
`ACSS` is adapted from @kate2019briefnoteasynchronousverifiable #cite(label("cryptoeprint:2021/159")).

"Secret Sharing" means that a secret value is split into `n` shares, and
distributed among `n` nodes in such a way that any subset of `f+1` shares
can be used to reconstruct the secret, but any subset of `f` shares cannot
learn anything about the secret.

"Complete" means that the protocol guarantees one ot two outcomes:
1. All honest nodes can reconstruct the same secret.
2. They collectively agree that the secret cannot be reconstructed (e.g.
   because the dealer is faulty and did not share the secret correctly).

The algorithm pseudocode is
#codelink("/packages/gpa/acss/acss.go#L10")[shown]
at the top of the source code. The main idea is that the "dealer" (the node that
wants to share the secret) splits the secret into `n` shares, and encrypts each
share using the public key of the corresponding peer. Then it broadcasts all the
shares using the #link(<rbc>)[`RBC` protocol]. After that, the nodes attempt to
decrypt the shares they receive, and inform each other about the success or
failure. After some rounds of communication, the nodes can decide whether they
have enough valid shares to reconstruct the secret, or if they collectively
agree that the secret cannot be reconstructed.

Note: for a receiver node, the output of `ACSS` is the received share. The
actual secret is never reconstructed inside `ACSS`.

== `RBC` <rbc>

#codelink("/packages/gpa/rbc/bracha/bracha.go")[`RBC`]
is an implementation of Bracha's Reliable Broadcast protocol. Its main purpose
is to allow a node to broadcast a message to all peers in such a way that all
honest nodes will receive the same message. `RBC` first appeared in
@BRACHA1987130, although our implementation is mainly taken from
#cite(label("10.1145/3460120.3484808")).

The pseudocode is
#codelink("/packages/gpa/rbc/bracha/bracha.go#L21")[shown]
at the top of the source code. The main idea is that the sender node broadcasts
the message using a simple "unreliable" broadcast, and then the nodes exchange
"echo" and "ready" messages to ensure that all honest nodes receive the same
message.

== `ACS` <acs>

#codelink("/packages/gpa/acs/acs.go")[`ACS`]
is an implementation of the Asynchronous Common Subset protocol, which is the
main component of the ISC Consensus protocol. It is responsible for taking all
the inputs from the different nodes in the committee, and deciding (agreeing) on
a common subset of them, so that every honest node in the committee will end up
with the same subset of inputs. `ACS` is mainly adapted from
#cite(label("10.1145/2976749.2978399")).

The pseudocode is
#link("https://github.com/iotaledger/wasp/blob/baa9674e79aa0e5cc7cfe5a1260cb7c4cb2f86ab/packages/gpa/acs/acs.go#L14")[shown]
at the top of the source code.

1. Each node $i$ receives an input $v_i$, and broadcasts it to all peers using
   the #link(<rbc>)[`RBC` protocol]. Thus, there are `n` instances of `RBC` running
   in parallel, one for each node in the committee.

2. Each node also starts `n` instances of the #link(<aba>)[`ABA` protocol].
   Whenever `RBC`$""_j$ outputs $v_j$, the node provides input `1` to
   `ABA`$""_j$. This essentially means that the node $i$ "votes" for including
   $v_j$ in the common subset.

3. When `n - f` instances of `ABA` have input 1, provide input 0 to all `ABA`s
   without input. This ensures that all `ABA` instances will eventually produce
   an output.

4. Once all `ABA` instances have produced an output, the ones that have output 1
   correspond to the inputs that are included in the common subset. The node can
   then output the common subset of inputs.

== `ABA` <aba>

#codelink("/packages/gpa/aba/mostefaoui/mostefaoui.go")[`ABA`]
is an implementation of the Asynchronous Binary Agreement protocol. Its purpose
is to allow a group of nodes to agree on a common binary value (0 or 1).
`ABA` first appeared in #cite(label("10.1145/2785953")), although our
implementation is mainly taken from #cite(label("10.1145/2976749.2978399")).

The pseudocode is
#codelink("/packages/gpa/aba/mostefaoui/mostefaoui.go#L23")[shown]
at the top of the source code.
The input to `ABA` is a binary value (0 or 1) that the node wants to propose.
The output is a binary value that is agreed upon by all honest nodes in the committee.
The protocol proceeds in rounds, and in each round the nodes exchange their
proposed values and a #link(<cc>)[common coin] value (which is a random value
that is the same for all nodes in the round).

The protocol ensures that if all honest nodes propose the same value, then that
value will be the output. If there is disagreement, the common coin is used to
break ties and ensure progress.

== `CC` <cc>

#codelink("/packages/gpa/cc/blssig/blssig.go")[`CC`]
is an implementation of the Common Coin protocol based on BLS threshold
signatures. A "Common Coin" is a random binary value (0 or 1) that is the same
for all peers. This protocol is described in Appendix C of #cite(label("10.1145/2976749.2978399")).

1. Each node takes as an input a "session ID" (a `[]byte`), so that many `CC`
   instances can run in parallel without interfering with each other. When
   used as part of `ABA`, the session ID includes the Consensus LogIndex and
   the `ABA` round number.

2. Each node computes a partial BLS signature of the session ID using its share of
   the distributed key, and broadcasts it to all peers.

3. When enough partial signatures are collected, the final BLS signature is reconstructed,
   and the output `0` or `1` is the least significant bit of the signature.

#bibliography("refs.bib", title: "References")
