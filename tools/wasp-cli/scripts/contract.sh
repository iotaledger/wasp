set -euox pipefail
IFS=$'\n\t'

if [ -f wasp-cli.json ]; then
	rm -i wasp-cli.json
fi

wasp-cli -d set utxodb true	
wasp-cli -d init
wasp-cli -w -d request-funds
wasp-cli -w -d chain deploy --chain=chain1 --committee=0,1,2,3 --quorum=3
wasp-cli -w -d chain deploy-contract wasmtimevm inccounter "inccounter SC" tools/cluster/tests/wasptest_new/wasm/inccounter_bg.wasm
wasp-cli -w -d chain post-request inccounter increment
