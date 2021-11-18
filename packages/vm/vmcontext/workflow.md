# Worklflow of request processing in the VM (very draft)

## General
Each UTXO and off-ledger request reaches the VM wrapped in `RequestData` interface. The `RequestData` wraps the following:
* The parsed output itself (as one of `iota.go` types) or off-ledger request in its entirety
* Gas budget
* if it is an output, the UTXO metadata:
    * Transaction ID
    * Output index in the transaction
    * Milestone index which confirmed the transaction
    * Milestone timestamp

###  Assumptions in the VM
* VM deals only with _correct_ UTXOs and off-ledger requests. If UTXO or off-ledger request does
  not parse as correct UTXO it should not reach the VM.
  The L1 node the Wasp it is connected to, is a **trusted party**. So, we should not expect malicious data
  (like wrong unparseable UTXO) from the L1 side. In these situations the node just panics.
* However, unexpected situations may happen due to asynchronicity, downtimes and so on. It must be processed gracefully
* chain's *alias output* never comes as `RequestData`. It is early recognized and comes to VM as `anchor`.
* if VM does not panic, it **always** produces a valid (confirmable) transaction essence and data state mutations
* A malicious data should always be expected from the off-ledger requests

So,
* the VM should provide reasonable reaction to any `RequestData` and any edge situation, or otherwise panic the node
* some `RequestData` may be skipped in the processing. It must be deterministic

_Question: is there a difference between syntactical and semantical correctness of separate UTXOs (not entire transactions)?_

### Edge situations
To be taken into account:
* *Replays*. `RequestData` may represent a duplicate i.e. it was already processed by the VM or it was even produced by the VM.
  The duplicate may come to the VM in ordinary course of events, also some unexpected or malicious way
* *No sender*. `sender` block in the UTXO may be absent, we cannot prevent it
* *Time unlockable* it means the received output can't be consumed based on time assumptions of the output and of the consensus time
* *Gas related* edge cases
* *Dust related* edge cases, for example not enough iotas to create new internal UTXO account
* some other?

#### Handling replays
Situation is output-type-specific. In general:
* the output produced by the VM itself may later come back to the VM as a `RequestData`.
  Can be recognized by `sender` being `self chain id`. *Extended outputs* may also be a self-posted requests, so VM must recognize those
  as such and act accordingly. One way of doing it is keeping a list of self posted requests in the state.
  Alternatively, a special flag or field in the output metadata can be introduced for this.
* UTXO may come again as a duplicate. It can only be prevented by performing a lookup into the `blocklog` receipts or in the `NFT` registry
* off-ledger request may be a duplicate (replay attack). It must be handled according to [replay-off-ledger.md](../../../documentation/temp/rfc/replay-off-ledger.md)

#### Handling wrong sender
If the UTXO have no sender block in general it is an error. The handling policy:
* accruing all iotas and native assets to the *common account* (owner's)
* consuming the output and destroying whenever relevant (extended, simple, unexpected alias output)
* alternatively: assigning UTXO as an asset owned by the owner. The owner would deal with it using special wallet functions of the `governance` contract

#### Time unlockable output
They should not reach VM. But if it reaches, it can easily be deterministically checked and respective `RequestData` ignored.

#### Gas and fee policy (temporary, to be discussed)
* Gas metering is always present, i.e. global gas variable is updated by `GasBurn` by the running SC
* Gas budget is always provided in the request
* View calls have a fixed gas budget, a constant set by chain. It should not be provided in the view call but it is used to cap the run.
* Gas checking is panic-ing when budget is exceeded and it is enabled
* Gas checking is always enabled. One of 2 options:
  * option I: fixed gas budget for each call. Fixed budget defined by chain level default, possibly contract level value. In this case gas budgets from requests are ignored
  * option II: dynamic gas budget for each call, taken from each request individually
* fees may be enabled or disabled
* if fees are disabled, tokens are not required to pay for gas. However, gas budgets (one of 2 options) are still enabled
* if fees are enabled, the real gas is taken from the caller's account according to the iota/gas ratio (see below tokens available for gas).
* iota/gas ratio is set in the governance contact. Gas market is also an option for the future, bad because of volatility

Question: what is the gas policy when processing NFT output? Probably fixed budget

### Asset transfers and caller's balance (temporary, to be discussed)
* assets in the request = iotas + native tokens
* each call carries some assets as transfer
* for UTXO request:
  * assets are coming with the output
  * total balances available for the target SC in the request as `Balances()` = `on-chain balances` + `incoming`
  * total balances available for gas is equal to on-chain balance of the sender (**this is different from what was before!!!**)
  * in the same request we may want have special property in the metadata `sender's assets`. It specifies balances to be taken from the
    `total assets in the UTXO` and credited to the sender's account.
  * `incoming` available for the target SC = `total assets in the UTXO` - `sender's assets`
* for off-ledger request:
  * attached assets just reference the on-ledger account of the sender. In the SC sandbox it is known as `incoming`
  * total balances available for the SC **and for the gas** in the off-ledger request = `on-chain balance`

Option: `sender's assets` metadata instead of being `any assets` may just be balance of the tokens used to pay for gas (usually iotas)

## Workflow per `RequestData` type

The VM is looping the slice of `[]requestdata.RequestData` (the batch) and calling `vm/request.go/RunTheRequest` method for each.
Here is described the workflow of how one `RequestData` is processed in the `RunTheRequest`.

**Note: dust provisions not yet covered**

### Off-ledger request
* check replay protection. Ignore duplicates
* charge `incoming` from the sender balance to the target
* run the request. Gas will be charged from the sender's balance. It will panic if not enough iotas or if exceeds gas budget

### Simple output
* accrue all assets to the chain's owner
* consume the UTXO and consolidate all assets with chain's L1 UTXO outputs
* no gas burned

### Extended output
* check if time unlockable. Ignore if not
* check if it has sender. If not, accrue all assets to the owner. Write respective receipt with apologies
* check if it is internal chain's account by checking sender's address and the flag of the internal account.
* If it is an internal account, ignore it all (or maybe assert consistency with the state?)
* check if it is a duplicate by looking up into the `blocklog` receipts
* otherwise it is a request.
* credit `sender's assets` to the sender's account
* `incoming` will become `assets on the output` - `sender's assets`
* run the request by calling target SC/entry point
* gas will be charged from the sender's account. In case of panic won't be refunded
* consume the output. Consolidate respective assets to chain's UTXOs. Produce respective outputs of chain's accounts

### NFT output
* check if time unlockable. Ignore if not
* check if the NFT ID is already owned by the chain. If yes, ignore
* check if it has sender. If not, accrue all assets to the owner, consume NFT and accrue it to the chain's owner. Write respective receipt with apologies
* credit `sender's assets` to the sender's account
* consume and produce new NFT output with altered request metadata and `assets on the output` - `sender's assets` as assets
  (they will be treated part of NFT and `incoming` will be amty. Alternatively, it may be accrued to the target contract through `incoming`)
* run the request by calling target contract/`special NFT entry point`
* gas will be charged from the sender's account.
* In case of gas panic what to do ???? Return? Assign to chain's owner? Ignore if has expiry option?

### Foundry output
* only can come if produced by the chain
* check cosnsistency and ignore

### Alias output and Unknown output
* if it comes, it means state or governance controller is set to the chain's address. Now the chain controls it
* if it is a replay, check consistency and ignore
* How to know to which target SC send it? The metadata is used for other purposes, need some generic metadata parsing
* check if time unlockable. Ignore if not

Probably the best strategy is to assign all unclear outputs to the chain's owner.
The `governance` contract should implement special wallet functions to deal with those unclear outputs: by sending them to sombody, destroying etc
