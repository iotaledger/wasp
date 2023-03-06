<script lang="ts">
  import {
    connected,
    web3,
    selectedAccount,
    chainId,
    chainData,
    defaultEvmStores,
  } from 'svelte-web3';
  import type { Contract } from 'web3-eth-contract';
  import { network } from '../../store';
  import { SingleNodeClient, type INativeToken } from '@iota/iota.js';
  import { gasFee, iscAbi, iscContractAddress } from './constants';
  import { hNameFromString } from '../../lib/hname';
  import { evmAddressToAgentID } from '../../lib/evm';
  import { getBalanceParameters, withdrawParameters } from './parameters';
  import { Bech32AddressLength, EVMAddressLength }  from '../../lib/constants';
    import { INSPECT_MAX_BYTES } from 'buffer';

  interface WithdrawState {
    /**
     * The current balance of the user
     */
    balance: number;

    /**
     * The reference to the ISC magic contract used for contract invocations
     */
    contract: Contract;

    /**
     * The EVM chain ID
     */
    evmChainID: number;
  }

  const state: WithdrawState = {
    balance: 0,
    contract: undefined,
    evmChainID: 0,
  };

  interface WithdrawFormInput {
    /**
     * [Form] The address to send funds to
     */
    receiverAddress: string;

    /**
     * [Form] The amount of base tokens to send.
     */
    baseTokensToSend: number;
  }

  const formInput: WithdrawFormInput = {
    receiverAddress: '',
    baseTokensToSend: 0,
  };

  $: formattedBalance = (state.balance / 1e6).toFixed(2);
  $: formattedAmountToSend = (formInput.baseTokensToSend / 1e6).toFixed(2);
  $: canSendFunds = state.balance > 0 && 
    formInput.baseTokensToSend > 0 && 
    formInput.receiverAddress.length == Bech32AddressLength;
  $: canSetAmountToSend = state.balance > gasFee + 1;

  async function pollBalance() {
    const addressBalance = await $web3.eth.getBalance(
      defaultEvmStores.$selectedAccount,
    );

    state.balance = Number(BigInt(addressBalance) / BigInt(1e12));

    if (formInput.baseTokensToSend > state.balance) {
      formInput.baseTokensToSend = 0;
    }
  }

  async function getNativeTokenIDs(): Promise<INativeToken[]> {
    if (!defaultEvmStores.$selectedAccount) {
      console.log('no account selected');
      return;
    }

    const accountsCoreContract = hNameFromString('accounts');
    const getBalanceFunc = hNameFromString('balance');
    const agentID = evmAddressToAgentID(defaultEvmStores.$selectedAccount);

    let parameters = getBalanceParameters(agentID);

    const result = await state.contract.methods
      .callView(accountsCoreContract, getBalanceFunc, parameters)
      .call();

      console.log(result)

    const nativeTokens = [];
    for (let item of result.items) {
      const id = item.key;

      // Ignore base token
      if (id.length <= 2) {
        continue;
      }
      
      const amount = BigInt(item.value);

      var nativeToken: INativeToken = {
        amount: amount.toString(),
        id: id,
      };

      nativeTokens.push(nativeToken);      
    }

    console.log(nativeTokens);
    return nativeTokens;
  }

  function subscribeBalance() {
    setTimeout(async () => {
      pollBalance();
      subscribeBalance();
    }, 2500);
  }

  async function connectToWallet() {
    await defaultEvmStores.setProvider();

    state.evmChainID = await $web3.eth.getChainId();
    state.contract = new $web3.eth.Contract(iscAbi, iscContractAddress, {
      from: defaultEvmStores.$selectedAccount,
    });

    await pollBalance();
    await getNativeTokenIDs();

    subscribeBalance();
  }

  async function onWithdrawClick() {
    if (!defaultEvmStores.$selectedAccount) {
      console.log('no account selected');
      return;
    }

    const client = new SingleNodeClient($network.apiEndpoint);
    let parameters = await withdrawParameters(client, formInput.receiverAddress, formInput.baseTokensToSend, gasFee);

    const result = await state.contract.methods.send(...parameters).send();
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
        <div class="chainid">{state.evmChainID}</div>
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
        bind:value={formInput.receiverAddress}
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
        max={state.balance}
        bind:value={formInput.baseTokensToSend}
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

  input[type='range'] {
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
