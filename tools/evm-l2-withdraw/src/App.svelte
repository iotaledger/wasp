<script lang="ts">
  import {
    connected,
    web3,
    selectedAccount,
    chainId,
    chainData,
    defaultEvmStores,
  } from "svelte-web3";

  import { Bech32Helper } from "@iota/iota.js";

  import iscAbiAsText from "../../../packages/vm/core/evm/iscmagic/ISC.abi?raw";

  const waspAddrBinaryFromBech32 = (bech32String) => {
    let receiverAddr = Bech32Helper.addressFromBech32(bech32String, "rms");
    let receiverAddrBinary = $web3.utils.hexToBytes(receiverAddr.pubKeyHash);
    //  // AddressEd25519 denotes an Ed25519 address.
    // AddressEd25519 AddressType = 0
    // // AddressAlias denotes an Alias address.
    // AddressAlias AddressType = 8
    // // AddressNFT denotes an NFT address.
    // AddressNFT AddressType = 16
    //
    // 0 is the ed25519 prefix
    return new Uint8Array([0, ...receiverAddrBinary]);
  };

  const iscAbi = JSON.parse(iscAbiAsText);
  const iscContractAddress: string =
    "0x0000000000000000000000000000000000001074";

  let chainID;
  let contract;

  async function connectToWallet() {
    await defaultEvmStores.setProvider();
    chainID = await $web3.eth.getChainId();
    contract = new $web3.eth.Contract(iscAbi, iscContractAddress, {
      from: defaultEvmStores.$selectedAccount,
    });
  }

  let addrInput = "";

  async function onWithdrawClick() {
    if (!defaultEvmStores.$selectedAccount) {
      console.log("no account selected");
      return;
    }

    let amount = 1e6; //1 million
    let parameters = [
      {
        // Receiver
        data: waspAddrBinaryFromBech32(addrInput),
      },
      {
        // Fungible Tokens
        baseTokens: amount,
        tokens: [],
      },
      false,
      {
        // Metadata
        targetContract: 0,
        entrypoint: 0,
        gasBudget: 0,
        params: {
          items: [],
        },
        allowance: {
          nfts: [],
          baseTokens: 0,
          tokens: [],
        },
      },
      {
        // Options
        timelock: 0,
        expiration: {
          time: 0,
          returnAddress: {
            data: [],
          },
        },
      },
    ];

    console.log(...parameters);

    const result = await contract.methods.send(...parameters).send();
    console.log(result);
  }

  connectToWallet();
</script>

<main>
  {#if !$connected}
    <button on:click={connectToWallet}>Connect to Wallet</button>
  {:else}
    Connected to Chain {$chainId}<br /><br />
    <input placeholder="rms..." style="width: 500px;" value={addrInput} /><br
    /><br />
    <button on:click={onWithdrawClick}>Withdraw</button><br />
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
