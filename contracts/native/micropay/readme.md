# MicroPay: a micro payment POC

The package implements hardcoded ISC contract for continuos micro payment protocol.

The final implementation will be in Rust.

## Use case

Party `P` (the _provider_) provides services to party `C` (the _consumer_).

Services are provided in small pieces, permanently, for example every several minutes. Telecom services would be an
example. Upon delivery of a piece of services, `P` expects a payment from `C` according of the service contract. If `C`
doesn't pay as expected, `P` may stop providing services. The payment usually very small, normally no more that some
thousands of iotas per piece, a micro payment. Due to continuous flow of micro payments and micro services, the parties
has limited financial risk.

Due to the confirmation latencies it is not practical to settle each micro payment into the DLT, IOTA Tangle in this
case. Therefore, the approach in the MicroPay is for `P` to collect many payments into a batch and then settle it all
into DLT at once: the _off-ledger_ (a.k.a. _off-chain_, _off_tangle_) approach.

## Requirements/assumptions

* the off-ledger payments bear risk of not being able to settle them later due to shortage of funds
* to eliminate that risk the `P` require a sum from `C` to be locked at third trusted part, a smart contract.
  This `paymentWarrant` (another word?) is used to settle the payments,
* The contract guarantees settlement of _off-ledger_ payments up to the amount left in the `paymentWarrant`
  account.
* `C` can top up the `paymentWarrant`, however it is out of reach for `C` while the contract is in effect
* `C` may announce it stops using the services at any time. `P` has in `closeTimeout` must settle all of pending
  _off-ledger_ payments. It means closing the contract and releasing what has left in `paymentWarrant` back to `C`.
* If `P` fails to close the contract from its side in `closeTimeout`, the `paymentWarrant` it release automatically and
  claims for payments are not accepted. It means, if not settled in time, the paymenst may be lost for `P` 

