# the deployer's EVM account private key
# the account should be funded
# (some fund has been deposited from L1 to L2, you can use `wasp-cli deposit`)
EVM_PRIV_KEY="<EVM_PRIV_KEY>"
# the deployer's EVM account address
EVM_ADDR="0x7321F89DA1F5266948c65319ef4C7F25c0b4ded5"

forge create --broadcast out/weth9/WETH9.sol:WETH9 \
             --rpc-url https://api.evm.lb-0.h.testnet.iota.cafe/v1/chain/evm \
             --private-key $(EVM_PRIV_KEY) \
             --legacy
# [⠊] Compiling...
# No files changed, compilation skipped
# Deployer: 0x7321F89DA1F5266948c65319ef4C7F25c0b4ded5
# Deployed to: 0x349F71aE4cC64D5c6Def4eAf964467AAFDc3Da10
# Transaction hash: 0x6f1cb8d9aab46820b7bc0704d65619133d617c68792c9eb602c45f8e9f97a71e

# WETH9 address from the WETH9 deploying command
WETH9_ADDR="0x349F71aE4cC64D5c6Def4eAf964467AAFDc3Da10"

# use uniswap's deploying script
# refer https://bobanetwork.medium.com/deploying-uniswap-v3-a-developers-guide-to-navigating-the-complexities-269dde21bff6
# git clone https://github.com/Uniswap/deploy-v3.git --depth=1
# cd deploy-v3
# yarn && yarn build
# 
# yarn start -pk <EVM_PRIV_KEY> \
#             -j https://api.evm.lb-0.h.testnet.iota.cafe/v1/chain/evm \ 
#             -w9 $(WETH9_ADDR) \
#             -ncl IOTA \
#             -o $(EVM_ADDR) \
#             -c 1
# 
# # result
# ➜  deploy-v3 git:(main) ✗ yarn start -pk <EVM_PRIV_KEY> -j https://api.evm.lb-0.h.testnet.iota.cafe/v1/chain/evm -w9 0x349F71aE4cC64D5c6Def4eAf964467AAFDc3Da10 -ncl IOTA -o 0x7321F89DA1F5266948c65319ef4C7F25c0b4ded5 -c 1
# yarn run v1.22.11
# $ yarn build
# $ ncc build index.ts -o dist -m
# ncc: Version 0.33.1
# ncc: Compiling file index.js into CJS
# ncc: Using typescript@4.2.3 (local user-provided)
# 1405kB  dist/index.js
# 1405kB  [2331ms] - ncc 0.33.1
# $ cat shebang.txt dist/index.js > dist/index.cmd.js && mv dist/index.cmd.js dist/index.js
# $ node dist/index.js -pk <EVM_PRIV_KEY> -j https://api.evm.lb-0.h.testnet.iota.cafe/v1/chain/evm -w9 0x349F71aE4cC64D5c6Def4eAf964467AAFDc3Da10 -ncl IOTA -o 0x7321F89DA1F5266948c65319ef4C7F25c0b4ded5 -c 1
# Step 1 complete [
#   {
#     message: 'Contract UniswapV3Factory deployed',
#     address: '0xda499309692E7beD5eF2f98257349c2D9a448617',
#     hash: '0x4587c3ffdac52d8c14b1e364786df1e0599a08ea4c9e5aaf8c6f06b57b07cd4d'
#   }
# ]
# state.v3CoreFactoryAddress:  0xda499309692E7beD5eF2f98257349c2D9a448617
# Step 2 complete [
#   {
#     message: 'UniswapV3Factory added a new fee tier 1 bps with tick spacing 1',
#     hash: '0x47e5ac0fff71297a1c3ae95f0b5bece2bc990ebe2d261905949561ebc18dd8eb'
#   }
# ]
# Step 3 complete [
#   {
#     message: 'Contract UniswapInterfaceMulticall deployed',
#     address: '0xbC2fdfF29fe15384389143A28D4aa0E4765bdf7B',
#     hash: '0xcf791a2a930b64f4989be64b2d6310f7fd541ad951344b31ebe9fd90e93a35f0'
#   }
# ]
# Step 4 complete [
#   {
#     message: 'Contract ProxyAdmin deployed',
#     address: '0x254852B9422e202c190B068dB0e4A3e4341d31De',
#     hash: '0x6501ddd909064c6dc355fa1630f320c63872409a0a4a52301405ec399355df05'
#   }
# ]
# Step 5 complete [
#   {
#     message: 'Contract TickLens deployed',
#     address: '0x6093fca2f7b10C1e4052a0Ac3B418A0219B1AA2C',
#     hash: '0x3b2449afe4834970f818d223b68d1823aa122fc98d5f8a46247c7e28c6ef3f98'
#   }
# ]
# Step 6 complete [
#   {
#     message: 'Library NFTDescriptor deployed',
#     address: '0x93eF7d628281ab3692d95f4514cbA307f7aF7AFE',
#     hash: '0x3dcd2856c19fd72149b63a687028a66bc77b425fa58e119a329a695a75a913ba'
#   }
# ]
# Step 7 complete [
#   {
#     message: 'Contract NonfungibleTokenPositionDescriptor deployed',
#     address: '0x68C3FDdd5911B55231b488a856876911CB64a878',
#     hash: '0x13f21b08cddb6b4bb4166c6706f8e8f422d6b35f69268e17088875639768a2dc'
#   }
# ]
# Step 8 complete [
#   {
#     message: 'Contract TransparentUpgradeableProxy deployed',
#     address: '0x2392B2d3Cd2Fe9c6120967168B23E815c964FE0e',
#     hash: '0x0194a50f7625f6b63d50d413d78a81c7c41dded99e3ee6450b87d7afb3e684df'
#   }
# ]
# Step 9 complete [
#   {
#     message: 'Contract NonfungiblePositionManager deployed',
#     address: '0xe6861c155cb8630372e76762deA595682c89DB32',
#     hash: '0xebd8c3c13abf5d40aa604e1663ff6fa2fc2e58f0cd5a0f0ce47e6afc187d1cb9'
#   }
# ]
# Step 10 complete [
#   {
#     message: 'Contract V3Migrator deployed',
#     address: '0x5385F4807a5E65C9ec799ab0eE0dccD02CedF6ac',
#     hash: '0xd755352b743ff6dd7dc3cf35f5253a01ed7490836e3db91714465f39e3752b16'
#   }
# ]
# Step 11 complete [
#   {
#     message: 'UniswapV3Factory owned by 0x7321F89DA1F5266948c65319ef4C7F25c0b4ded5 already'
#   }
# ]
# Step 12 complete [
#   {
#     message: 'Contract UniswapV3Staker deployed',
#     address: '0x5FB5cC94AEB0DE107C0c4958D38BAd958827D70d',
#     hash: '0xe40391b7ae59f06fdd7b84cd8688ac49e0e08dc8f313af314cfffe16429994b9'
#   }
# ]
# Step 13 complete [
#   {
#     message: 'Contract QuoterV2 deployed',
#     address: '0xd26EcB8DD244f9589c005e55FCacc8f0ebE2BF1b',
#     hash: '0x7d6ea0bcbb83eb0e2b217863958f85ca5f709beb9e296d9c3a201312faa126ea'
#   }
# ]
# Step 14 complete [
#   {
#     message: 'Contract SwapRouter02 deployed',
#     address: '0xadEDAbdb44A883C0103C582e4e66F9879E98AB3e',
#     hash: '0x1840c351a11610cd18b61819e58effa6fe9f1943538236ce617585dec1c560f6'
#   }
# ]
# Step 15 complete [
#   {
#     message: 'ProxyAdmin owned by 0x7321F89DA1F5266948c65319ef4C7F25c0b4ded5 already'
#   }
# ]
# Deployment succeeded
# [[{"message":"Contract UniswapV3Factory deployed","address":"0xda499309692E7beD5eF2f98257349c2D9a448617","hash":"0x4587c3ffdac52d8c14b1e364786df1e0599a08ea4c9e5aaf8c6f06b57b07cd4d"}],[{"message":"UniswapV3Factory added a new fee tier 1 bps with tick spacing 1","hash":"0x47e5ac0fff71297a1c3ae95f0b5bece2bc990ebe2d261905949561ebc18dd8eb"}],[{"message":"Contract UniswapInterfaceMulticall deployed","address":"0xbC2fdfF29fe15384389143A28D4aa0E4765bdf7B","hash":"0xcf791a2a930b64f4989be64b2d6310f7fd541ad951344b31ebe9fd90e93a35f0"}],[{"message":"Contract ProxyAdmin deployed","address":"0x254852B9422e202c190B068dB0e4A3e4341d31De","hash":"0x6501ddd909064c6dc355fa1630f320c63872409a0a4a52301405ec399355df05"}],[{"message":"Contract TickLens deployed","address":"0x6093fca2f7b10C1e4052a0Ac3B418A0219B1AA2C","hash":"0x3b2449afe4834970f818d223b68d1823aa122fc98d5f8a46247c7e28c6ef3f98"}],[{"message":"Library NFTDescriptor deployed","address":"0x93eF7d628281ab3692d95f4514cbA307f7aF7AFE","hash":"0x3dcd2856c19fd72149b63a687028a66bc77b425fa58e119a329a695a75a913ba"}],[{"message":"Contract NonfungibleTokenPositionDescriptor deployed","address":"0x68C3FDdd5911B55231b488a856876911CB64a878","hash":"0x13f21b08cddb6b4bb4166c6706f8e8f422d6b35f69268e17088875639768a2dc"}],[{"message":"Contract TransparentUpgradeableProxy deployed","address":"0x2392B2d3Cd2Fe9c6120967168B23E815c964FE0e","hash":"0x0194a50f7625f6b63d50d413d78a81c7c41dded99e3ee6450b87d7afb3e684df"}],[{"message":"Contract NonfungiblePositionManager deployed","address":"0xe6861c155cb8630372e76762deA595682c89DB32","hash":"0xebd8c3c13abf5d40aa604e1663ff6fa2fc2e58f0cd5a0f0ce47e6afc187d1cb9"}],[{"message":"Contract V3Migrator deployed","address":"0x5385F4807a5E65C9ec799ab0eE0dccD02CedF6ac","hash":"0xd755352b743ff6dd7dc3cf35f5253a01ed7490836e3db91714465f39e3752b16"}],[{"message":"UniswapV3Factory owned by 0x7321F89DA1F5266948c65319ef4C7F25c0b4ded5 already"}],[{"message":"Contract UniswapV3Staker deployed","address":"0x5FB5cC94AEB0DE107C0c4958D38BAd958827D70d","hash":"0xe40391b7ae59f06fdd7b84cd8688ac49e0e08dc8f313af314cfffe16429994b9"}],[{"message":"Contract QuoterV2 deployed","address":"0xd26EcB8DD244f9589c005e55FCacc8f0ebE2BF1b","hash":"0x7d6ea0bcbb83eb0e2b217863958f85ca5f709beb9e296d9c3a201312faa126ea"}],[{"message":"Contract SwapRouter02 deployed","address":"0xadEDAbdb44A883C0103C582e4e66F9879E98AB3e","hash":"0x1840c351a11610cd18b61819e58effa6fe9f1943538236ce617585dec1c560f6"}],[{"message":"ProxyAdmin owned by 0x7321F89DA1F5266948c65319ef4C7F25c0b4ded5 already"}]]
# Final state
# {"v3CoreFactoryAddress":"0xda499309692E7beD5eF2f98257349c2D9a448617","multicall2Address":"0xbC2fdfF29fe15384389143A28D4aa0E4765bdf7B","proxyAdminAddress":"0x254852B9422e202c190B068dB0e4A3e4341d31De","tickLensAddress":"0x6093fca2f7b10C1e4052a0Ac3B418A0219B1AA2C","nftDescriptorLibraryAddressV1_3_0":"0x93eF7d628281ab3692d95f4514cbA307f7aF7AFE","nonfungibleTokenPositionDescriptorAddressV1_3_0":"0x68C3FDdd5911B55231b488a856876911CB64a878","descriptorProxyAddress":"0x2392B2d3Cd2Fe9c6120967168B23E815c964FE0e","nonfungibleTokenPositionManagerAddress":"0xe6861c155cb8630372e76762deA595682c89DB32","v3MigratorAddress":"0x5385F4807a5E65C9ec799ab0eE0dccD02CedF6ac","v3StakerAddress":"0x5FB5cC94AEB0DE107C0c4958D38BAd958827D70d","quoterV2Address":"0xd26EcB8DD244f9589c005e55FCacc8f0ebE2BF1b","swapRouter02":"0xadEDAbdb44A883C0103C582e4e66F9879E98AB3e"}
# ✨  Done in 79.09s.