# Snaphot tool

The snapshot-related library code is located in `packages/snapshot`. The `tools/snapshot` directory contains code of the `snapshot` executable.

## What is *snapshot*?
*Snaphot* is the state of the chain stored separately. Node can start the chain from snapshot, i.e. it assumes current state of the chain is equal to the one provided and then syncronizes the state with other nodes.

## How we create a snapshot of the chain's state?

There are two possibilities:
- by backing up the database
- by creating a snapshot file

### Database backup

All the data of the chain is stored in the `rocksdb` key/value database. The name of the directory of the database is hex-encoded `chainID` of the chain. It contains full state of the chain with trie at some state index.

Upon activation of the chain in the node, the `state manager` automatically loads the state in the database and synchronizes it with the chain.

The directory can be backed up and restored with usual backup tools.

Note, that `state manager` updates the database when syncing the state and when state transition occurs, i.e. at least once per 10 sec. It means, for the backed up database to be consistent, all chain activity must be stopped.

Pros of database copy:
- simple, no need for special tools

Cons of database copy:
- database is 2-3 times larger than the state itself, because it also contains *trie* and database overhead.

### Snapshot file

Snapshot file is a file which contains only key/value pairs of the state. It does not contain any information, such as `trie` which can be computed from the state itself.

Snapshot file name is `<chainID>.<index>.snapshot`, where `<chainID>` is hex-encoded `chainID` of the chain, `<index>` is state index of the snapshot.

For example for chainID = `210fc7bb818639ac48a4c6afa2f1581a8b9525e2` snapshot file for the state index `314` will be `210fc7bb818639ac48a4c6afa2f1581a8b9525e2.314.snapshot`.

#### Create snapshot file

`snapshot -create <chainID>`

The command expect database of the chain in the current directory. The chain must be deactivated or the node must be stopped.

The command iterates the detabase and creates the snapshot file.

**Important.** The order in which key/value pairs in database are iterated, is **not-deterministic** in general. This means, that content of the snapshot file (and, therefore, digest/hash of it) is not guaranteed to be the same for the same source database.

The above, however, does not violate consisency of the database. It is alway guaranteed, that *root commitment* to the state is fully deterministic.

#### Scan snapshot file

`snapshot -scan <chainID>` scans the file, checks it formatting and extracts main parameters of the state stored in the snapshot itself (not in the file name):

* chainID
* state index
* state timestamp
* number of key/value pairs

#### Restore database from the snapshot file

`snapshot -restoredb <snapshot file>`

The command scans file and rebuilds chain's database from it. It may take some time. 
The command does the following:
* reads key/value pairs from file one by one and writes it as state mutations
* build complete trie of the state
* periodically commits and flushes updates to the database

#### Verify snapshot file against database

`snapshot -verify <snapshot file>`

The command assumes databae is already restored. The command does the follwing:
* reads key/value from the file one-by-one.
* for each key/value pair it retrieves *proof of inclusion* from the state and verifies it.

The command may be lengthy because proof generation and verification are expensive operations.

The `-verify` command is rarely needed practice if the database is restored from snapshot. 
It is mostly used for testing and benchmarking.

#### Validating the snapshot

**TODO not yet implemented** because it requires special WEB APIs to L1 and L2 (Wasp) nodes.

`snapshot -validate <chainID> <L1 API endpoint> <L2 API endpoint>`

The purpose of the command is to make sure the restored state in the database is indeed a valid state snapshot of the chain.

It will work the following way:
* restore the snapshot from file into the database
* take root commitment of the restored state
* ask the L2 node for the *proof of the past state* that the state root commitment with the specific state index of the restored state is commited in the chain state as a past state (the commitment should be  contained in the `blocklog` partition).
* validate the *proof of the past state* with the L1 node.

To perform the `-validate` operation we need the following WEB APIs:
* L2 (Wasp) `get proof of inclusion` (see `viewcontext.GetMerkleProof` call)
* L1: `get state commitment` which retrieves `anchor output` of the chain (`AliasOutput`) with the state commitment. The proof must be verified agains the state commitment stored in the anchor.    

## Benchmarks

The benchmarks were obtained by using `snapshot` tool on randomly 
generated state data (max key len 64 byte, max value 128 bytes): 
one with ~1 mil of key/value pairs, another with 10 mil of key/value pairs. 
On the laptop 4 core, 2.6 GHz with 32 MB RAM and SDD HD.


| Parameter                                                | ~1 mil records   | ~10 mil records |
|----------------------------------------------------------|------------------|-----------------|
| Chain database size (with trie)                          | 189 MB           | 1.91 GB         |
| Snapshot file size                                       | 89 MB            | 896 MB          |
| Database (with trie) restored in                         | 21 sec           | 3 min 51 sec    |
| Generate/verify proofs (total time)                      | 1 min 29 sec     | 18 min 16 sec   |
| Generate/verify proofs (speed, clear cache every 100000) | 11000 proofs/sec | 8800 proofs/sec |
| Create snapshot file from db                             | 7 sec            | 1 min 11 sec    |


