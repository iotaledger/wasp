<script lang="ts">
  import {
    connected,
    web3,
    selectedAccount,
    chainId,
    defaultEvmStores,
  } from 'svelte-web3';
  import { setIntervalAsync, clearIntervalAsync } from 'set-interval-async';
  import { nodeClient, indexerClient, selectedNetwork } from '../../store';
  import {
    gasFee,
    iscAbi,
    iscContractAddress,
    multiCallAbi,
  } from './constants';
  import { Bech32AddressLength } from '../../lib/constants';
  import { onDestroy, onMount } from 'svelte';
  import { toast } from '@zerodevx/svelte-toast';
  import type { WithdrawFormInput, WithdrawState } from './component_types';
  import { ISCMagic } from './iscmagic/iscmagic';
  import type { INativeToken } from '../../lib/native_token';
  import type { INFT } from '../../lib/nft';

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
    nftIDToSend: undefined,
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
      const isNativeTokenAvailable =
        state.availableNativeTokens.findIndex(x => x.id == nativeTokenID) >= 0;

      if (!isNativeTokenAvailable) {
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

    state.availableNFTs = await state.iscMagic.getNFTs(
      $nodeClient,
      $indexerClient,
      $selectedAccount,
    );
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
      state.iscMagic = new ISCMagic(state.contract, null);
      await pollAccount();
      await subscribeBalance();
    } catch (ex) {
      toast.push(`Failed to connect to wallet: ${ex}`);
      console.log('connectToWallet', ex);
    }
    state.isLoading = false;
  }
  async function withdraw(
    baseTokens: number,
    nativeTokens: INativeToken[],
    nftID?: string,
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
        nftID,
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

    await withdraw(
      formInput.baseTokensToSend,
      nativeTokensToSend,
      formInput.nftIDToSend,
    );
  }

  function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
  async function onWithdrawEverythingClick() {
    try {
      for (let nft of state.availableNFTs.reverse()) {
        await pollBalance();
        await withdraw(gasFee * 1000, [], nft.id);
        await sleep(5 * 1000);
      }

      await pollBalance();
      await withdraw(
        state.availableBaseTokens,
        state.availableNativeTokens,
        null,
      );
    } catch {}
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
        <div class="balance">{formattedBalance}</div>
      </div>
    </div>

    <div class="input_container">
      <span class="header">Receiver address</span>
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
            disabled={!canSetAmountToWithdraw}
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
      <div class="header">NFT to send</div>

      <div class="token_list">
        <select bind:value={formInput.nftIDToSend}>
          <option value={undefined} />
          {#each state.availableNFTs as nft}
            <option value={nft.id}>
              {nft.metadata.name}
              {nft.id}
            </option>
          {/each}
        </select>
      </div>
    </div>

    <div class="input_container">
      <button disabled={!canWithdraw} on:click={onWithdrawClick}>
        Withdraw
      </button>
    </div>
    <div class="input_container">
      <button
        class="warning"
        disabled={!canWithdrawEverything}
        on:click={onWithdrawEverythingClick}
      >
        Withdraw everything at once
      </button>
    </div>
  {/if}
</component>

<style>
  .warning:disabled {
    background-color: #6a1b1e;
  }
  .warning {
    background-color: #b92e34;
    border-color: red;
    color: white;
  }
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
