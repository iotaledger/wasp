root_path=$(git rev-parse --show-toplevel)
cd $root_path/iota-go/contracts/testcoin
iota move build 
iota client publish --gas-budget 1000000000 --skip-dependency-verification --json > publish_receipt.json
