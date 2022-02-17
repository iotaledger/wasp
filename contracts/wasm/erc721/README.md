# ERC721 as IOTA smart contract

### Deploy new chain (optional)

```shell
$ ./wasp-cli chain deploy --committee=0 --quorum=1 --chain=wasptest --description="ArgentinaHub"
# You can replace the amount of committees and quorum by your needs
```

### Deposit IOTA tokens to the chain
```shell
$ ./wasp-cli chain deposit IOTA:1000
```

### Deploy contract
```shell
./wasp-cli --verbose chain deploy-contract wasmtime erc721 "Argentina Hub" contracts/wasm/erc721/pkg/nft_bg.wasm string n string ArgHub string s string ARH
# n: Name
# s: Symbol
```

### Mint new erc721
```shell
./wasp-cli --verbose chain post-request erc721 mint string tokenid int <token-id>
string tokenuri string <token-uri>
#Example
# <token-id> = 73798465
# <token-uri> = "<IPFS_LINK>"
```

### Transfer ownership of the erc721
```shell
./wasp-cli --verbose chain post-request erc721 transferFrom string from agentid <from-agentid> string to agentid <to-agentid> string tokenid int <token-id>
# Example
# <from-agentid> = A/1urERbzXL1iraoW2jMvkLdFuMmPS711BYtA8u4L6sS53::00000000 
# <to-agentid> = A/111111111111111111111111111111111::00000000
# <token-id> = 73798465
```

### Approves another agentid to transfer the given token
```shell
./wasp-cli --verbose chain post-request erc721 approve string to agentid <agentid> string tokenid int <tokenid>
# Example
# <agentid> = A/1urERbzXL1iraoW2jMvkLdFuMmPS711BYtA8u4L6sS53::00000000
# <tokenid> = 73798465
```

### Balance Of
```shell
./wasp-cli chain call-view erc721 balanceOf string account agentid <agentid> | ./wasp-cli decode string amount int
# Example
# <agentid> = A/1urERbzXL1iraoW2jMvkLdFuMmPS711BYtA8u4L6sS53::00000000
```

### Name of contract
```shell
./wasp-cli chain call-view erc721 name | ./wasp-cli decode string name string 
```

### Symbol of contract
```shell
./wasp-cli chain call-view erc721 name | ./wasp-cli decode string name string 
```

### Check agent id if it's approved (1 = true ; 0 = false)
```shell
./wasp-cli chain call-view erc721 isApproved string tokenid int <token-id> string operator agentid <agent-id> | ./wasp-cli decode string approved int
# Example
# <token-id> = 73798465
# <agent-id> = A/12uApMh48Nq9EZGB8idco4W5NpZtyu6vv4sXpZuP5FUKs::00000000
```

### Get token URI
```shell
./wasp-cli chain call-view erc721 tokenURI string tokenid int <token-id> | ./wasp-cli decode string uri string
# Example
# <token-id> = 73798465
```

### Set approval for all (1 = true ; 0 = false)
```shell
./wasp-cli --verbose chain post-request erc721 setApprovalForAll string operator agentid <agent-id> string approved int 1
# Example
# <agent-id> = A/12uApMh48Nq9EZGB8idco4W5NpZtyu6vv4sXpZuP5FUKs::00000000
```
