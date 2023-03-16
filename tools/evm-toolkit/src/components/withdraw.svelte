<script lang="ts">
  import {
    connected,
    web3,
    selectedAccount,
    chainId,
    defaultEvmStores,
  } from 'svelte-web3';
  import { setIntervalAsync, clearIntervalAsync } from 'set-interval-async';
  import { nodeClient, indexerClient, selectedNetwork } from '../store';
  import {
    gasFee,
    iscAbi,
    iscContractAddress,
    multiCallAbi,
  } from './withdraw/constants';
  import { Bech32AddressLength } from '../lib/constants';
  import { onDestroy, onMount } from 'svelte';
  import { toast } from '@zerodevx/svelte-toast';
  import type { WithdrawFormInput, WithdrawState } from './withdraw/component_types';
  import { ISCMagic } from './withdraw/iscmagic/iscmagic';
  import type { INativeToken } from '../../lib/native_token';
  import type { INFT } from '../../lib/nft';
  import { Input, Button } from '.';
  import { InputType } from '$lib/enums';

  const state: WithdrawState = {
    availableBaseTokens: 0,
    availableNativeTokens: [],
    availableNFTs: [],
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
  $: isValidAddress = formInput.receiverAddress.length == Bech32AddressLength;
  $: canWithdraw =
    state.availableBaseTokens > 0 &&
    formInput.baseTokensToSend > 0 &&
    isValidAddress;
  $: canWithdrawEverything = isValidAddress;
  $: canSetAmountToWithdraw = state.availableBaseTokens > gasFee + 1;
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
    state.availableBaseTokens = await state.iscMagic.getBaseTokens(
      $web3.eth,
      $selectedAccount,
    );

    if (formInput.baseTokensToSend > state.availableBaseTokens) {
      formInput.baseTokensToSend = 0;
    }
  }

  async function pollNativeTokens() {
    if (!$selectedAccount) {
      return;
    }

    state.availableNativeTokens = await state.iscMagic.getNativeTokens(
      $nodeClient,
      $indexerClient,
      $selectedAccount,
    );

    // Remove native tokens marked to be sent if the token does not exist anymore.
    for (const nativeTokenID of Object.keys(formInput.nativeTokensToSend)) {
      if (typeof state.availableBaseTokens[nativeTokenID] == 'undefined') {
        delete formInput.nativeTokensToSend[nativeTokenID];
      }
    }

    // Add all existing native tokens to the "to be sent" array but with an amount of 0
    // This makes it easier to connect the UI with the withdraw request.
    for (const nativeToken of state.availableNativeTokens) {
      if (typeof formInput.nativeTokensToSend[nativeToken.id] == 'undefined') {
        formInput.nativeTokensToSend[nativeToken.id] = 0;
      }
    }
  }

  async function pollNFTs() {
    if (!$selectedAccount) {
      return;
    }

    state.availableNFTs = await state.iscMagic.getNFTs($nodeClient, $indexerClient, $selectedAccount);
  }

  async function pollAccount() {
    await Promise.all([pollBalance(), pollNativeTokens(), pollNFTs()]);
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
      /*console.log(multiCallAbi);
      const multiCall = new $web3.eth.Contract(
        multiCallAbi,
        $selectedNetwork.multicallAddress,
        {
          from: $selectedAccount,
        },
      );*/

      state.iscMagic = new ISCMagic(state.contract, null);

      await pollAccount();
      await subscribeBalance();

      /*await state.iscMagic.withdrawMulticall(
        $web3,
        $selectedNetwork.multicallAddress,
        $nodeClient,
        'formInput.receiverAddress',
        123123123,
        [],
        null,
      );*/
    } catch (ex) {
      toast.push(`Failed to connect to wallet: ${ex}`);
      console.log('connectToWallet', ex);
    }

    state.isLoading = false;
  }

  async function withdraw(
    baseTokens: number,
    nativeTokens: INativeToken[],
    nft?: INFT,
  ) {
    if (!$selectedAccount) {
      return;
    }

    let result: any;

    try {
      result = await state.iscMagic.withdraw(
        $nodeClient,
        formInput.receiverAddress,
        baseTokens,
        nativeTokens,
        nft,
      );
    } catch (ex) {
      toast.push(
        `Failed to send withdraw request: ${JSON.stringify(ex, null, 4)}`,
        {
          duration: 8000,
        },
      );
      console.log(ex);
      return;
    }

    console.log(result);

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

  async function onWithdrawClick() {
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

    await withdraw(formInput.baseTokensToSend, nativeTokensToSend, undefined);
  }

  async function onWithdrawEverythingClick() {
    /*for (let nft of state.availableNFTs) {
      await pollBalance();
      await withdraw(900000, [], nft);
    }

    await pollBalance();
    await withdraw(
      state.availableBaseTokens,
      state.availableNativeTokens,
      null,
    );*/

    await pollAccount();

    await state.iscMagic.withdrawMulticall(
      $web3,
      $selectedNetwork.multicallAddress,
      $nodeClient,
      formInput.receiverAddress,
      state.availableBaseTokens,
      state.availableNativeTokens,
      state.availableNFTs,
    );
  }

</script>

<withdraw-component class="flex flex-col space-y-6 mt-6">
  {#if !$connected}
    <div class="input_container">
      <button on:click={connectToWallet}>Connect to Wallet</button>
    </div>
  {:else if !state.isLoading}
    <info-box>
      <div class="flex flex-col space-y-2">
        <info-item-title>Chain ID</info-item-title>
        <info-item-value>{$chainId}</info-item-value>
      </div>
      <div class="flex flex-col space-y-2">
        <info-item-title>Balance</info-item-title>
        <info-item-value>{formattedBalance}</info-item-value>
      </div>
    </info-box>
    <Input
      type={InputType.Text}
      label="Receiver address"
      bind:value={formInput.receiverAddress}
      placeholder="L1 address starting with (rms/tst/...)"
      stretch
    />
    <tokens-to-send-wrapper>
      <div class="mb-2">Tokens to send</div>
      <info-box class="flex flex-col space-y-2 max-h-96 overflow-auto">
        <div>
          <info-item-title>
            SMR Token: {formattedAmountToSend}
          </info-item-title>
          <input
            type="range"
            disabled={!canSetAmountToWithdraw}
            min="0"
            max={state.availableBaseTokens}
            bind:value={formInput.baseTokensToSend}
          />
        </div>

        {#each state.availableNativeTokens as nativeToken}
          <div>
            <info-item-title>
              {nativeToken.metadata.name} Token: {formInput.nativeTokensToSend[
                nativeToken.id
              ] || 0}
            </info-item-title>
            <input
              type="range"
              min="0"
              max={Number(nativeToken.amount)}
              bind:value={formInput.nativeTokensToSend[nativeToken.id]}
            />
          </div>
        {/each}
      </info-box>
    </tokens-to-send-wrapper>
    {#if state.availableNFTs.length > 0}
      <nfts-wrapper>
        <div class="mb-2">NFTs</div>
        <info-box>
          <div class="flex flex-col space-y-2">
            {#each state.availableNFTs as nft}
              <info-item-title>
                {nft.id}
              </info-item-title>
            {/each}
          </div>
        </info-box>
      </nfts-wrapper>
    {/if}
    <Button
      title="Withdraw"
      onClick={onWithdrawClick}
      disabled={!canWithdraw}
      stretch
    />
    <Button
      danger
      title="Withdraw everything at once"
      onClick={onWithdrawEverythingClick}
      disabled={!canWithdrawEverything}
      stretch
    />
  {/if}
</withdraw-component>

<style>
  info-box {
    @apply w-full;
    @apply flex;
    @apply justify-between;
    @apply bg-shimmer-background-tertiary;
    @apply rounded-xl;
    @apply p-4;
  }
  info-item-title {
    @apply text-xs;
    @apply text-shimmer-text-secondary;
  }

  info-item-value {
    @apply text-2xl;
  }
</style>
