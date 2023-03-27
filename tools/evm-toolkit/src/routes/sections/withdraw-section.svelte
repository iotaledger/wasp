<script lang="ts">
  import { chainId, connected, selectedAccount } from 'svelte-web3';
  import { Button, Input, RangeInput } from '$components';
  import { InputType } from '$lib/common/enums';
  import { Bech32AddressLength } from '$lib/constants';
  import { nodeClient } from '$lib/evm-toolkit';
  import type { INativeToken } from '$lib/native-token';
  import { NotificationType, showNotification } from '$lib/notification';
  import type { WithdrawFormInput } from '$lib/withdraw';
  import {
    connectToWallet,
    gasFee,
    pollBalance,
    withdrawStateStore,
  } from '$lib/withdraw';

  const formInput: WithdrawFormInput = {
    receiverAddress: '',
    baseTokensToSend: 0,
    nativeTokensToSend: {},
    nftIDToSend: undefined,
  };

  let isWithdrawing: boolean = false;

  $: formattedBalance = ($withdrawStateStore.availableBaseTokens / 1e6).toFixed(
    2,
  );
  $: formattedAmountToSend = (formInput.baseTokensToSend / 1e6).toFixed(2);
  $: isValidAddress = formInput.receiverAddress.length == Bech32AddressLength;
  $: canWithdraw =
    $withdrawStateStore.availableBaseTokens > 0 &&
    formInput.baseTokensToSend > 0 &&
    isValidAddress;
  $: canWithdrawEverything = isValidAddress;
  $: canSetAmountToWithdraw =
    $withdrawStateStore.availableBaseTokens > gasFee + 1;
  $: $withdrawStateStore.isMetamaskConnected = window.ethereum
    ? window.ethereum.isConnected()
    : false;

  $: $withdrawStateStore, updateFormInput();

  function updateFormInput() {
    if (formInput.baseTokensToSend > $withdrawStateStore.availableBaseTokens) {
      formInput.baseTokensToSend = 0;
    }
    // Remove native tokens marked to be sent if the token does not exist anymore.
    for (const nativeTokenID of Object.keys(formInput.nativeTokensToSend)) {
      const isNativeTokenAvailable =
        $withdrawStateStore.availableNativeTokens.findIndex(
          x => x.id == nativeTokenID,
        ) >= 0;

      if (!isNativeTokenAvailable) {
        delete formInput.nativeTokensToSend[nativeTokenID];
      }
    }
    // Add all existing native tokens to the "to be sent" array but with an amount of 0
    // This makes it easier to connect the UI with the withdraw request.
    for (const nativeToken of $withdrawStateStore.availableNativeTokens) {
      if (typeof formInput.nativeTokensToSend[nativeToken.id] == 'undefined') {
        formInput.nativeTokensToSend[nativeToken.id] = 0;
      }
    }
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
      isWithdrawing = true;
      result = await $withdrawStateStore.iscMagic.withdraw(
        $nodeClient,
        formInput.receiverAddress,
        baseTokens,
        nativeTokens,
        nftID,
      );
    } catch (ex) {
      isWithdrawing = false;
      showNotification({
        type: NotificationType.Error,
        message: `Failed to send withdraw request: ${ex.message}`,
        duration: 8000,
      });
      console.log(ex);
      return;
    }

    if (result.status) {
      showNotification({
        type: NotificationType.Success,
        message: `Withdraw request sent. BlockIndex: ${result.blockNumber}`,
        duration: 4000,
      });
    } else {
      showNotification({
        type: NotificationType.Error,
        message: `Failed to send withdraw request: ${JSON.stringify(
          result,
          null,
          4,
        )}`,
        duration: 8000,
      });
    }
    isWithdrawing = false;
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
      for (let nft of $withdrawStateStore.availableNFTs.reverse()) {
        await pollBalance();
        await withdraw(gasFee * 1000, [], nft.id);
        await sleep(5 * 1000);
      }

      await pollBalance();
      await withdraw(
        $withdrawStateStore.availableBaseTokens,
        $withdrawStateStore.availableNativeTokens,
        null,
      );
    } catch {}
  }
</script>

<withdraw-component class="flex flex-col space-y-6 mt-6">
  {#if !$connected && !$selectedAccount}
    <div class="input_container">
      <button on:click={connectToWallet}>Connect to Wallet</button>
    </div>
  {:else if !$withdrawStateStore.isLoading}
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
      <info-box class="flex flex-col space-y-4 max-h-96 overflow-auto">
        <RangeInput
          label="SMR Token: {formattedAmountToSend}"
          bind:value={formInput.baseTokensToSend}
          disabled={!canSetAmountToWithdraw}
          min="0"
          max={$withdrawStateStore.availableBaseTokens}
        />

        {#each $withdrawStateStore.availableNativeTokens as nativeToken}
          <RangeInput
            bind:value={formInput.nativeTokensToSend[nativeToken.id]}
            label="{nativeToken?.metadata?.name ?? ''} Token: {formInput
              .nativeTokensToSend[nativeToken.id] || 0}"
            min="0"
            max={Number(nativeToken.amount)}
          />
        {/each}
      </info-box>
    </tokens-to-send-wrapper>
    {#if $withdrawStateStore.availableNFTs.length > 0}
      <nfts-wrapper>
        <div class="mb-2">NFTs</div>
        <info-box>
          <div class="flex flex-col space-y-2">
            {#each $withdrawStateStore.availableNFTs as nft}
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
      busy={isWithdrawing}
      stretch
    />
    <Button
      danger
      title="Withdraw everything at once"
      onClick={onWithdrawEverythingClick}
      disabled={!canWithdrawEverything || isWithdrawing}
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
