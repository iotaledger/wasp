<script lang="ts">
  import {
    connected,
    web3,
    selectedAccount,
    chainId,
    chainData,
    defaultEvmStores,
  } from 'svelte-web3';

  import { Bech32Helper, type IEd25519Address } from '@iota/iota.js';

  import iscAbiAsText from '../../assets/ISC.abi?raw';

  const waspAddrBinaryFromBech32 = (bech32String: string) => {
    // Depending on the network, the human readable part can change (tst, rms, ..).
    // - We need some kind of API that is not the direct Hornet node to fetch it..
    // - Maybe over some EVM info route?
    // For this PoC it should be enough to substr the first three chars.
    let humanReadablePart = bech32String.substring(0, 3);

    let receiverAddr = Bech32Helper.addressFromBech32(
      bech32String,
      humanReadablePart
    );
    
    const address: IEd25519Address = receiverAddr as IEd25519Address;

    let receiverAddrBinary = $web3.utils.hexToBytes(address.pubKeyHash);
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

  const gasFee = 300;
  const iscAbi = JSON.parse(iscAbiAsText);
  const iscContractAddress: string =
    '0x0000000000000000000000000000000000001074';

  let chainID;
  let contract;
  let balance = 0;
  let amountToSend = 0;

  $: formattedBalance = (balance / 1e6).toFixed(2);
  $: formattedAmountToSend = (amountToSend / 1e6).toFixed(2);
  $: canSendFunds = balance > 0 && amountToSend > 0;
  $: canSetAmountToSend = balance > gasFee+1;

  let addrInput = '';

  async function pollBalance() {
    const addressBalance = await $web3.eth.getBalance(
      defaultEvmStores.$selectedAccount
    );
    balance = Number(BigInt(addressBalance) / BigInt(1e12));

    if (amountToSend > balance) {
      amountToSend = 0;
    }
  }

  function subscribeBalance() {
    setTimeout(async () => {
      pollBalance();
      subscribeBalance();
    }, 2500);
  }

  async function connectToWallet() {
    await defaultEvmStores.setProvider();
    chainID = await $web3.eth.getChainId();
    contract = new $web3.eth.Contract(iscAbi, iscContractAddress, {
      from: defaultEvmStores.$selectedAccount,
    });

    await pollBalance();
    subscribeBalance();
  }

  async function onWithdrawClick() {
    if (!defaultEvmStores.$selectedAccount) {
      console.log('no account selected');
      return;
    }

    let parameters = [
      {
        // Receiver
        data: waspAddrBinaryFromBech32(addrInput),
      },
      {
        // Fungible Tokens
        baseTokens: amountToSend - gasFee,
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

    const result = await contract.methods.send(...parameters).send();
    console.log(result);
  }

</script>
<component>
  {#if !$connected}
  <div class="input_container">
    <button on:click={connectToWallet}>Connect to Wallet</button>
  </div>
  {:else}
    <div class="account_container">
      <div class="chain_container">
        <div>Chain ID</div>
        <div class="chainid">{$chainId}</div>
      </div>
      <div class="balance_container">
        <div>Balance</div>
        <div class="balance">{formattedBalance}Mi</div>
      </div>

    </div>

    <div class="input_container">
      <span class="header">Receiver address </span>
      <input
        type="text"
        placeholder="L1 address starting with (rms/tst/...)"
        bind:value={addrInput}
      />
    </div>

    <div class="input_container">
      <div class="header">
        Amount to send: {formattedAmountToSend}Mi  
      </div>
      <input type="range" disabled="{!canSetAmountToSend}" min="0" max={balance} bind:value={amountToSend} />
    </div>

    <div class="input_container">
      <button disabled="{!canSendFunds}" on:click={onWithdrawClick}>Withdraw</button><br />
    </div>
  {/if}
</component>

<style>
  component {
    color: rgba(255, 255, 255, 0.87);
    display: flex;
    flex-direction: column;
  }

  input[type=range] {
    width: 100%;
    padding: 10px 0 0 0;
    margin: 0;
  }

  .account_container {
    height: 64px;
    margin: 15px;
    display: flex;
    justify-content: space-between;
  }

  .balance_container {
    text-align: right;
  }

  .balance, .chainid {
    padding-top: 5px;
    font-weight: 800;
    font-size: 32px;
  }
</style>
