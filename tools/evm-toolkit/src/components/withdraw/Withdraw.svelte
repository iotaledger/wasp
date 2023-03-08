<script lang="ts">
  import {
    connected,
    web3,
    selectedAccount,
    chainId,
    defaultEvmStores,
  } from 'svelte-web3';
  import { setIntervalAsync, clearIntervalAsync } from 'set-interval-async';
  import { nodeClient, indexerClient } from '../../store';
  import { Converter } from '@iota/util.js';
  import { gasFee, iscAbi, iscContractAddress } from './constants';
  import { hNameFromString } from '../../lib/hname';
  import { evmAddressToAgentID } from '../../lib/evm';
  import { getBalanceParameters, withdrawParameters } from './parameters';
  import {
    Bech32AddressLength,
    NativeTokenIDLength,
  } from '../../lib/constants';
  import {
    getNativeTokenMetaData,
    type INativeToken,
  } from '../../lib/native_token';
  import { onDestroy, onMount } from 'svelte';
  import { toast } from '@zerodevx/svelte-toast';
  import type { WithdrawFormInput, WithdrawState } from './component_types';

  const state: WithdrawState = {
    availableBaseTokens: 0,
    availableNativeTokens: [],
    contract: undefined,
    evmChainID: 0,

    balancePollingHandle: undefined,
    isMetamaskConnected: false,
    isLoading: true,
  };

  const formInput: WithdrawFormInput = {
    receiverAddress: '',
    baseTokensToSend: 0,
    nativeTokensToSend: {},
  };

  $: formattedBalance = (state.availableBaseTokens / 1e6).toFixed(2);
  $: formattedAmountToSend = (formInput.baseTokensToSend / 1e6).toFixed(2);
  $: canSendFunds =
    state.availableBaseTokens > 0 &&
    formInput.baseTokensToSend > 0 &&
    formInput.receiverAddress.length == Bech32AddressLength;
  $: canSetAmountToSend = state.availableBaseTokens > gasFee + 1;
  $: state.isMetamaskConnected = window.ethereum
    ? window.ethereum.isConnected()
    : false;

  onDestroy(async () => {
    await unsubscribeBalance();
  });

  onMount(async () => {
    // It's a bit confusing:
    // $connected does only return true if Metamask is connected to the page AND the defaultProvider is initialized.
    // This makes us unable to automatically initialize the store as it will open a Metamask authorization request without indicating why immediately on the first visit.
    // We can use window.ethereum.isConnected to first validate if the user already has set up a connection by clicking "Connect Wallet".
    // Then we can automatically initialize the store and not require manual user interaction each time. (User only has to click "Connect Wallet" once).
    if (state.isMetamaskConnected) {
      await connectToWallet();
    }
  });

  async function pollBalance() {
    const addressBalance = await $web3.eth.getBalance($selectedAccount);

    state.availableBaseTokens = Number(BigInt(addressBalance) / BigInt(1e12));

    if (formInput.baseTokensToSend > state.availableBaseTokens) {
      formInput.baseTokensToSend = 0;
    }
  }

  async function pollNativeTokens() {
    if (!$selectedAccount) {
      return;
    }

    const accountsCoreContract = hNameFromString('accounts');
    const getBalanceFunc = hNameFromString('balance');
    const agentID = evmAddressToAgentID($selectedAccount);

    let parameters = getBalanceParameters(agentID);

    const result = await state.contract.methods
      .callView(accountsCoreContract, getBalanceFunc, parameters)
      .call();

    const nativeTokens: INativeToken[] = [];

    for (let item of result.items) {
      const id = item.key;
      const idBytes = Converter.hexToBytes(id);

      if (idBytes.length != NativeTokenIDLength) {
        continue;
      }

      var nativeToken: INativeToken = {
        // TODO: BigInt is required for native tokens, but it causes problems with the range slider. This needs to be adressed before shipping.
        amount: BigInt(item.value),
        id: id,
        metadata: await getNativeTokenMetaData($nodeClient, $indexerClient, id),
      };

      nativeTokens.push(nativeToken);
    }

    state.availableNativeTokens = nativeTokens;

    for (const nativeToken of nativeTokens) {
      if (typeof formInput.nativeTokensToSend[nativeToken.id] == 'undefined') {
        formInput.nativeTokensToSend[nativeToken.id] = 0;
      }
    }
  }

  async function pollAccount() {
    await Promise.all([pollBalance(), pollNativeTokens()]);
  }

  async function subscribeBalance() {
    if (state.balancePollingHandle) {
      return;
    }

    state.balancePollingHandle = setIntervalAsync(pollAccount, 2500);
  }

  async function unsubscribeBalance() {
    if (!state.balancePollingHandle) {
      return;
    }

    await clearIntervalAsync(state.balancePollingHandle);
    state.balancePollingHandle = undefined;
  }

  async function connectToWallet() {
    state.isLoading = true;

    try {
      await defaultEvmStores.setProvider();

      state.evmChainID = await $web3.eth.getChainId();
      state.contract = new $web3.eth.Contract(iscAbi, iscContractAddress, {
        from: $selectedAccount,
      });

      await pollAccount();
      await subscribeBalance();
    } catch (ex) {
      toast.push(`Failed to connect to wallet: ${ex}`);
      console.log('connectToWallet', ex);
    }

    state.isLoading = false;
  }

  async function onWithdrawClick() {
    if (!$selectedAccount) {
      return;
    }

    const nativeTokensToSend: INativeToken[] = [];

    for (const tokenID of Object.keys(formInput.nativeTokensToSend)) {
      const amount = formInput.nativeTokensToSend[tokenID];

      if (amount > 0) {
        nativeTokensToSend.push({
          // TODO: BigInt is required for native tokens, but it causes problems with the range slider. This needs to be adressed before shipping.
          // In this function the amount is actually of type "number" not bigint, so we lose precision at 53bits which is a problem that needs to be solved.
          amount: BigInt(amount),
          id: tokenID,
        });
      }
    }

    let parameters = await withdrawParameters(
      $nodeClient,
      formInput.receiverAddress,
      gasFee,
      formInput.baseTokensToSend,
      nativeTokensToSend,
    );

    let result: any;

    try {
      result = await state.contract.methods.send(...parameters).send();
    } catch (ex) {
      toast.push(
        `Failed to send withdraw request: ${JSON.stringify(ex, null, 4)}`,
        {
          duration: 8000,
        },
      );

      return;
    }

    if (result.status) {
      toast.push(`Withdraw request sent. BlockIndex: ${result.blockNumber}`, {
        duration: 4000,
      });
    } else {
      toast.push(
        `Failed to send withdraw request: ${JSON.stringify(result, null, 4)}`,
        {
          duration: 8000,
        },
      );
    }
  }
</script>

<component>
  {#if !$connected}
    <div class="input_container">
      <button on:click={connectToWallet}>Connect to Wallet</button>
    </div>
  {:else if !state.isLoading}
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
        bind:value={formInput.receiverAddress}
      />
    </div>

    <div class="input_container">
      <div class="header">Tokens to send</div>

      <div class="token_list">
        <div class="token_list-item">
          <div class="header">
            SMR Token: {formattedAmountToSend}
          </div>

          <input
            type="range"
            disabled={!canSetAmountToSend}
            min="0"
            max={state.availableBaseTokens}
            bind:value={formInput.baseTokensToSend}
          />
        </div>

        {#each state.availableNativeTokens as nativeToken}
          <div class="token_list-item">
            <div class="header">
              {nativeToken.metadata.name} Token: {formInput.nativeTokensToSend[
                nativeToken.id
              ] || 0}
            </div>
            <input
              type="range"
              min="0"
              max={Number(nativeToken.amount)}
              bind:value={formInput.nativeTokensToSend[nativeToken.id]}
            />
          </div>
        {/each}
      </div>
    </div>

    <div class="input_container">
      <button disabled={!canSendFunds} on:click={onWithdrawClick}
        >Withdraw</button
      ><br />
    </div>
  {/if}
</component>

<style>
  .token_list {
    display: flex;
    flex-direction: column;
  }

  .token_list-item {
    border: 1px solid gray;
    border-radius: 4px;
    padding: 20px;
    margin: 10px;
    margin-left: 0;
  }

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
