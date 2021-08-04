# How accounts work

Each ISCP chain keeps a ledger of on-chain account balances.

## Account ownership

An on-chain account is identified by an AgentID.

- The AgentID for accounts owned by L1 entities (regular IOTA wallets) looks like the following:

    ```yaml
    Hname: 0
    Address: "some address"
    ```

- The AgentID for accounts owned by L2 entities (Smart Contracts) :

    ```yaml
    Hname: "Hname of the entity"
    Address: "Address of the chain where the entity exists"
    ```

    _example_: the smart contract with hname `123` that exists on the chain with address `000`, can be identified on **any** chain by the following AgentID:

    ```yaml
    Hname: 123
    Address: 000
    ```

## The Accounts Contract

The `Accounts` contract manages what funds are owned by which accounts.

Internally there is a mapping of `Account (AgentID)` to `balances`, which can include normal IOTAs and/or any colored tokens.

### Entrypoints

The entrypoints available for this Smart Contract are:

**Views**:

- `balance` - get the account balance of a specific account

  parameters:

  - `ParamAgentID` - account's AgentID

  returns:

  - a map of [token_color] -> [amount]

- `totalAssets` - get the total colored balances controlled by the chain
  
  returns:

  - a map of [token_color] -> [amount]

- `accounts` - get a list of all accounts existing on the chain

    returns:

  - a list of accounts (AgentIDs)

**Calls**:

- deposit funds // TODO add link
- withdrawal funds // TODO add link
- harvest // TODO
