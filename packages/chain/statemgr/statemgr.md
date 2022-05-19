# State manager

## Overview

The role of the state manager is to keep internal state of Wasp node (called `solidState`) in sync with the state, which is stored in L1. It receives state updates in the form of alias output (together with output ID) from L1 and stores the newest one (with highest state index) in `stateOutput`. It also receives new state with its approving output ID from consensus (another component of the same Wasp node), however before accepting it as a new `solidState`, the state manager ensures that it is approved by L1, by waiting for the alias output with respective output ID to be received from L1. In addition, the state manager regularly polls L1 for the latest state output.

![State manager diagram](statemgr_diagram.png)

## Need to catch up

When the new `stateOutput` is received from L1 it may result in `solidState` being behind it, meaning that `solidState.index < stateOutput.stateIndex`. This means that state manager current state is older than current state of the chain in L1 and the state manager needs to catch up --- synchronise its state.

If the Wasp node has some performance, communication or other issues, it may result in `solidState` being many states behind the newest `stateOutput`. When `solidState.index == stateOutput.stateIndex`, state manager is synchronised.  `solidState.index` may never be larger than `stateOutput.stateIndex`.

## Write ahead log (WAL)

After obtaining (calculating) next state, before posting it to L1 for acceptance and before sending it to the state manager, the consensus stores the last block, which was used to obtain the state, locally in the structure, called WAL. This is done to avoid a dead lock, when the new state is posted and accepted by L1, but the node crashes before the state manager had been able to store the block after the state synchronisation. When the new `stateOutput` is received from L1, the WAL is checked and every existing block with index from `solidState.index+1` to `stateOutput.stateIndex` is retrieved.

## Synchronisation

For exact code, see `packages/chain/statemgr/syncmgr.go`, especially `doSyncActionIfNeeded` method.

To synchronise the state to some index `n`, the state manager must have correct blocks with indices from `solidState.index+1` to `n` and it must contain information, needed to approve the state index `n`. Such information is the commitment of state index `n`, which consists of:
* `blockHash` --- hash of the block, which is applied to state index `n-1` to obtain state index `n`;
* `stateCommitment` --- commitment of the state index `n`.

In a perfect case `n == stateOutput.stateIndex`. However, this is not necessary: alias output with state index `n` might have earlier been received by the state manager, providing the commitment of state index `n` and `n` might be less than current `stateOutput.stateIndex`. Thus, upon receiving new alias output (`output`) from L1, the state manager stores the two fields of the commitment in respectively `approval[output.stateIndex].blockHash` and `approval[output.stateIndex].stateCommitment`.

This allows state manager to catch up the current state in smaller steps. Say, `solidState.index == 10`, `stateOutput.stateIndex == 40` and alias outputs with indices `20` and `30` have been received earlier. Thus instead of synchronising from index `11` straight to index `40`, state manager is capable of synchronising from `11` to `20`, then from `21` to `30` and finally from `31` to `40`. This avoids large memory usage by storing many blocks at once. Moreover, synchronisation cannot be done, if `solidState.index` is more than `10 000` smaller than `stateOutput.stateIndex`. This constant is hardcoded, but is not based on any tests. It is used to avoid crashes due to memory shortage.

The synchronisation algorithm is a cycle through indices, that have to be synchronised, starting with index `i := solidState.index+1`. Note that for synchronisation to be needed, `solidState.index` must be smaller than `stateOutput.stateIndex` and thus the initial `i <= stateOutput.stateIndex`. The iteration of the cycle is as follows:
1.  If block index `i` has already been retrieved from WAL, it is accepted as the only possible correct candidate. WAL is written by the node itself and thus it is assumed that the node cannot be malicious to itself.
1.  If block index `i` was not found in WAL, (at most) 5 random other nodes are requested. The request is asynchronous. Upon receiving the block index `i` from any node, it is stored as a possible candidate. This step is skipped, if not enough time has passed since the last request of block index `i`.
1.  If there are no block candidates for index `i`, then the synchronisation algorithm cannot proceed and is terminated. Hopefully on the next run of the algorithm, at least one block with index `i` will be present.
1.  If there is at least one block candidate for index `i`, then it is checked, if among the candidates there is an approved block. For this to be true, alias output with state index `i` must have already been received from L1 and the block with hash value `approval[i].blockHash` must exist among the block candidates. If this is the case, then, because of the algorithm, for every index from `solidState.index+1` to `i` there is at least one block candidate. Now an attempt is made to form the correct sequence of blocks with indices from `solidState.index+1` to `i` as retrieved block with hash `approval[i].blockHash` is certainly a correct one (it is approved by alias output, received from L1). Every block (as well as the retrieved one with hash `approval[i].blockHash`) contains information about previous state (index `i-1` in the case of the retrieved block) commitment. Thus the hash of the block, used to create the previous state, may be obtained and the sequence of blocks is formed as follows:

    1.  `blocks[i]` is the retrieved block with hash `approval[i].blockHash`.
    2.  `blocks[i-1]` is a block from block candidates with index `i-1`, which has a hash `blocks[i].previousL1Commitment.blockHash`, if such block exists.
    3.  `blocks[i-2]` is a block from block candidates with index `i-2`, which has a hash `blocks[i-1].previousL1Commitment.blockHash`, if such block exists.
    4.  etc., until `blocks[solidState.index+1]` is obtained.

    If some blocks in a sequence from `blocks[solidState.index+1]` to `blocks[i]` are missing, then the synchronisation algorithm cannot continue and is terminated as some correct blocks are missing. Hopefully on the next run of the algorithm, all the correct blocks will be present among the candidates. If however every block of such sequence is successfully retrieved, then an attempt to create a new state is made:
    1.  `solidState` is taken as a base for `newSolidState`.
    1.  if the commitment of `newSolidState` is the same as `blocks[solidState.index+1].previousL1Commitment.stateCommitment`, then `blocks[solidState.index+1]` is applied to `newSolidState`.
    1.  if the commitment of `newSolidState` (now with `blocks[solidState.index+1]` applied) is the same as `blocks[solidState.index+2].previousL1Commitment.stateCommitment`, then `blocks[solidState.index+2]` is applied to `newSolidState`.
    1.  etc., until all the blocks up to `blocks[i]` are applied.
    1.  finally it is checked, if the obtained `newSolidState` is what was expected: if its commitment matches `approval[i].stateCommitment`.

    If any of these steps fail, then something went really wrong, therefore all the candidate blocks of all the indices are removed and synchronisation is restarted. If however `newSolidState` is formed without errors, then this `newSolidState` together with all the `blocks` is committed to the DB and all the block candidates for indices `solidState.index+1` to `i` are removed as they are no longer needed. Finally, `newSolidState` becomes new `solidState` and therefore state manager gets synchronised until state index `i`.

1.  If among the block candidates for index `i` there is no approved block, or if approval information for block index `i` is not known (alias output with state index `i` has never been received), then `i` is increased by `1` and the next iteration of synchronisation cycle starts from step #1. The last iteration is when `i == stateOutput.stateIndex`.
