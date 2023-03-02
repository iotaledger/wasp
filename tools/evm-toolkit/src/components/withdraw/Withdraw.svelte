<script lang="ts">
  import {
    connected,
    web3,
    selectedAccount,
    chainId,
    chainData,
    defaultEvmStores
  } from "svelte-web3";
  import type {Contract} from 'web3-eth-contract';
  import { hornetAPI } from "../../store";
  import { waspAddrBinaryFromBech32 } from "../../lib/bech32";
  import { SingleNodeClient } from "@iota/iota.js";
  import { gasFee, iscAbi, iscContractAddress } from "./constants";
  import { hNameFromString } from "../../lib/hname";
  import { Converter } from "@iota/util.js";
  import { evmAddressToAgentID } from "../../lib/evm";

  let chainID: string;
  let contract: Contract;
  let balance = 0;
  let amountToSend = 0;

  let addrInput: string = "";

  $: formattedBalance = (balance / 1e6).toFixed(2);
  $: formattedAmountToSend = (amountToSend / 1e6).toFixed(2);
  $: canSendFunds = balance > 0 && amountToSend > 0;
  $: canSetAmountToSend = balance > gasFee + 1;

  console.log(hNameFromString("accounts"))

  async function pollBalance() {
    const addressBalance = await $web3.eth.getBalance(
      defaultEvmStores.$selectedAccount
    );
    balance = Number(BigInt(addressBalance) / BigInt(1e12));

    if (amountToSend > balance) {
      amountToSend = 0;
    }
  }

  async function getNativeTokens() {
    if (!defaultEvmStores.$selectedAccount) {
      console.log("no account selected");
      return;
    }

    const accountsCoreContract = hNameFromString("accounts");
    const getBalanceFunc = hNameFromString("balance");

    const agentID = evmAddressToAgentID(defaultEvmStores.$selectedAccount)
    
    let parameters = {
      items: [
        {
          key: Converter.utf8ToBytes("a"),
          value: agentID,
        }
      ],
    };

    const result = await contract.methods.callView(accountsCoreContract, getBalanceFunc, parameters).call();
    console.log(result);
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
    await getNativeTokens();


    subscribeBalance();
  }

  async function onWithdrawClick() {
    if (!defaultEvmStores.$selectedAccount) {
      console.log("no account selected");
      return;
    }

    const client = new SingleNodeClient($hornetAPI);

    let parameters = [
      {
        // Receiver
        data: await waspAddrBinaryFromBech32(client, addrInput),
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
      <input
        type="range"
        disabled={!canSetAmountToSend}
        min="0"
        max={balance}
        bind:value={amountToSend}
      />
    </div>

    <div class="input_container">
      <button disabled={!canSendFunds} on:click={onWithdrawClick}
        >Withdraw</button
      ><br />
    </div>
  {/if}
</component>

<style>
  component {
    color: rgba(255, 255, 255, 0.87);
    display: flex;
    flex-direction: column;
  }

  input[type="range"] {
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

  .balance,
  .chainid {
    padding-top: 5px;
    font-weight: 800;
    font-size: 32px;
  }
</style>
