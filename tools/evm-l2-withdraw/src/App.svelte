<script lang="ts">
import { connected, web3, selectedAccount, chainId, chainData, defaultEvmStores } from 'svelte-web3';

import iscAbiAsText from '../../../packages/vm/core/evm/isccontract/ISC.abi?raw';

const iscAbi = JSON.parse(iscAbiAsText);
const iscContractAddress: string = '0x0000000000000000000000000000000000001074';

async function connectToWallet() {
  await defaultEvmStores.setProvider()
}

async function start() {
  await defaultEvmStores.setProvider();

  const chainId = await $web3.eth.getChainId()
  const contract = new $web3.eth.Contract(iscAbi, iscContractAddress, {
    from: '0x8B65DD08C7784017fe6B8Af20904e61916506fD4'
  });

  var amount = $web3.utils.toBN(1074);

  var parameters = [
    { // Receiver
        data: new Uint8Array(32) 
      }, 
      { // Fungible Tokens
        baseTokens: 1074,
        tokens: []
      }, 
      true,
      { // Metadata
        targetContract: $web3.utils.hexToNumber('0x3c4b5e02'),
        entrypoint: $web3.utils.hexToNumber('0x23f4e3a1'), 
        gasBudget: 50000000,

        params: {
          items: [
            {
              key: "x".charCodeAt(0), 
              value: "0x" + amount //TODO: Fix
            }
          ]
        },

        allowance: {
          nfts:[],
          baseTokens: 1074,
          assets: [
            /*{
              ID: [],
              Amount: 1
            }*/
          ]
            
        }
      },
      { // Options
        timelock: 0,
        expiration: {
          time: 0,
          returnAddress: {
            data: []
          }
        }
      } 
  ];

  console.log(...parameters)

  const result = await contract.methods.send(...parameters).send();
  console.log(result)
}


start();

</script>

<main>
  {#if !$connected}
    <button on:click="{connectToWallet}">Connect to Wallet</button>
  {:else}
      Connected to Chain {$chainId}
  {/if}
</main>

<style>
  .logo {
    height: 6em;
    padding: 1.5em;
    will-change: filter;
  }
  .logo:hover {
    filter: drop-shadow(0 0 2em #646cffaa);
  }
  .logo.svelte:hover {
    filter: drop-shadow(0 0 2em #ff3e00aa);
  }
  .read-the-docs {
    color: #888;
  }
</style>