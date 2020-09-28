# Committees and deployment of the smart contract

## The committee

Smart contract instance is a _dApp_, a distributed application or program. 
It is run by a distributed network of Wasp nodes called the _committee_. 
A committee, small or big, is behind every smart contract. 

Network locations of the committee nodes are usually not a public information. 
The reason is to protect the committee node from DDoS and similar attacks. 
In the dashboards of _DonateWithFeedback_ and other PoC smart contracts we display it for demo purposes. 

For example, it shows the [instance](http://waspdev01.iota.cafe:10000/dwf) of _DonateWithFeedback_ 
is deployed on 5 nodes, _TokenRegistry_ on 10 and _FairAuction_ on 15. 
The committees of those 3 instances overlap. Any Wasp node can participate in many committees
of many smart contracts.

You may think a committee is a distributed processor which runs the smart contract program. 

The committee is specific for each SC instance, depending on how SC is _deployed_. 
It may be as small as 2 nodes, or it may have hundreds of nodes. 

The committee nodes run the program under the _distributed consensus_. 
In the ISCP each committee has specific number less or equal to the size of the committee called _quorum_ or 
_quorum factor_.  The _quorum_ is always strictly larger than the half of the size of the committee. 
Very often the quorum is equal to 2/3 plus one node.

The consensus guarantees that for each computed state update at least the _quorum_ of any committee nodes 
have 100% agreement on the computation result. 
The _state transition_ can occur only if the condition above is satisfied. 

For example, for the demo deployment of _FairAuction_ smart contract the quorum factor is set to 8. 
It means at least 8 out of 15 nodes must produce and sign with their private keys bit-by-bit equal computation 
results for that result to be accepted as a valid smart contact state update. 

After the quorum agrees on the result, the opinion of the rest of nodes is unimportant: 
those nodes may be down, faulty or malicious. 
The quorum already has consensus on which requests to process and in which order, to where and 
how to move tokens and update the values of the smart contract state. The quorum agrees on the same view to many things,
like the UTXO ledger, which one committee node is a leader for the round, what is the timestamp consistent 
with the local clocks of each node and which node will receive rewards (if the smart contract instance requires fees).

After the committee runs the program to process the request to the smart contract, it comes to 
the consensus on the result and posts the result transaction to the Tangle.

In order to fake the result, for example to steal iotas (to move them to the attacker’s address), at least 
a quorum of committee nodes must be corrupted. It is not feasible in most practical use cases with 
committee size and quorum large enough.

### Selecting the committee

The committee is crucial for the security of the smart contract. 

The topic of committee formation is very broad, and we do not intend to cover it all here.

The policies and protocol for committee selection may be centralized or it may be based on 
permissionless participation of nodes in an open market of nodes.

Centralized selection of committee nodes is common for corporate setting or in the consortium (in latter case
it actually will essentially be decentralized).

Committee may be selected in the open market where everyone participates in the parmissionless manner. 
In this case nodes may be selected (by whoever is forming the committee) based on committment of the node 
owners to SLA and other criteria. Trust to such a committee will be ensured by requiring stakes from the committee
nodes and setting up automatic punishment, including complete slashing of the stake, 
for not complying to SLA and other misbehavior.    

Note that at this stage we intentionally leave the committee selection policies and protocols outside of the 
ISCP itself because we want variety here. It is an important scope of further development of ISCP. 

## Deploying a smart contract 

Let's say we have committee nodes already selected (see above). 

We have to provide the following input to the deployment process:
- the _program hash_ (see on program hash [here](dwf.md)), which points to the program binary  
- the sorted list of committee nodes

## Deploy a new instance of _DonateWithFeedback_
We will deploy _DonateWithFeedback_ using `wwallet` program. 

For the _DonateWithFeedback_ the program hash is equal to `5ydEfDeAJZX6dh6Fy7tMoHcDeh42gENeqVDASGWuD64X`. 
It points to the code statically linked with the Wasp node.
In the particular case of PoC smart contracts the `wwallet` knows and derives the program hash from the `dwf` keyword
in the command line.  

The list of committee nodes must be listed in the respective `dwf` section of `wwallet.json`, in the current directory. 
Download example of `wwallet.json` via link in the demo dashboard.

The balance of the wallet must have at least 2 iotas in order to deploy smart contract. One of these iotas will be
minted (colored) to a new non-fungible token transferred to the smart contract address. 
It will remain there for its lifetime. Another iota will be returned back to the owner's balance after 
SC instance will be fully intialized.

The following command will deploy an instance of _DonateWithFeedback_ smart contact on the committee 
specified in the `wwallet.json`:

`wwallet dwf admin deploy` 

If successful, the program will print the address of the newly deployed instance among other things. 
From now on anyone can send `donate` requests to that address. 

## Deployment process in detail 

### The owner
The initiator of the deployment process is the owner of the wallet (the private key) and hence the tokens in the
wallet's address. The owner of the wallet becomes the `owner of the SC` as it performs the deployment 
using his/her wallet.
This ensures that for the smart contract the owner is authenticated and therefore known. 
 
The owner address of the SC becomes a privileged user of the smart contract. 
It is the only address which can send _protected requests_ to the smart contract. 
For the _DonateWithFeedback_ smart contract, the SC owner is the only one who can withdraw 
iotas from its account by sending `withdraw` request. Requests send by any other sender will have no effect.

Note that an owner in general doesn't _own_ anything which belongs to SC account. 
It is just a privileged user, recognized as such at the base protocol. 
The smart contract program may or may not treat it as privileged, depending on its algorithms.
Ownership of the smart contract can be transferred (this feature not implemented yet in PoC).     

### DKG and the origin transaction

Deployment of the smart contract consists of two main steps: 
- _distributed key generation_ (DKG) 
- creating the _origin transaction_.

#### DKG 

The purpose of the DKG is to generate a set of private keys among committee nodes in a distributed and secure manner. 
The process is triggered by the owner of the smart contract during deployment and then it is 
performed between committee nodes, in a distributed and leaderless manner. 

The result is a set of private keys, 
the master public key and the smart contract address, controlled by the quorum of private keys 
(partial keys or key shares). 

The quorum parameter T is an input parameter of the DKG process. The private keys are stored in each node's registry. 
In the future we plan to use hardware instances with IOTA Stronghold in it.

The base scenario of DKG is a completely decentralized one. 
It results in private keys only known by the committee nodes.

However, in many practical situations it may be needed for example to have possibility make backup copies 
of critical data or even take certain decisions on behalf of the committee 
(like move the SC instance to another committee). These situations may be common when we use 
centralized consensus on the committee during deployment.  
  
In these alternative situations DKG process may provide master private key or even backup of 
all private keys to the owner of the SC instance in secure manner. 
In any case, these settings are parameters of the protocol known to each Wasp node which is 
participating in DKG, so private keys can only be shared with the consensus of all nodes.  

#### The _origin transaction_. Color of the smart contract 

After the private keys and the smart contract address are generated and stored by the committee nodes, 
the _origin transaction_ of the smart contract must be created on the Tangle. 

The _origin transaction_ is a value transaction with certain data payload (a smart contract transaction) 
signed by the owner of the smart contract with its private key. 

With the origin transaction the owner mints 1 new token and transfers it to the smart contract address.
The resulting _non-fungible_ (unique) token takes the hash of the origin transaction as its _color code_.

The new colored token is the _smart contract token_. The _smart contract token_ remains in the smart contract 
address as long as the smart contract is using this address, i.e. normally for the lifetime.

The color code of the _smart contract token_ is equal to the hash of the origin 
transaction. It is called _color of the smart contract_.

You may find _color_ of the instance in the demo dashboard of the _DonateWithFeedback_ smart contract. 
You will also find exactly 1 token with the color equal to the smart contract color 
in the balances of the smart contract.

The non-fungible _smart contract token_ guarantees there is exactly 1 transaction 
and exactly 1 UTXO with this token on the ledger globally. 
This property prevents _forking_ the chain of state updates, i.e. there's always unique and objective SC state. 

The color of the smart contract remains the same during the lifetime of the smart contract. 
In some situations a smart contract with its state may be moved to another address. 
In that case the address of the smart contract will change, and the _smart contract token_ will be moved to it. 
So, the color of the smart contract (the color of the token) will remain the same for its lifetime 
even if the address will change. Therefore, the color of the instance is an ultimate identity of 
it on the global ledger.

The owner of the smart contract mints the smart contract token and therefore becomes authenticated and singled 
out participant in the smart contract since the origin. 
This makes the owner to be a privileged user of the smart contract: the builtin logic of the smart contract will 
process privileged request codes only if sent from the owner’s address.

