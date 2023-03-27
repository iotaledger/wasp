<script lang="ts">
  import { onMount } from 'svelte';
  import { connected, selectedAccount } from 'svelte-web3';

  import { Input, Select } from '$components';

  import { InputType } from '$lib/common';
  import {
    networks,
    selectedNetwork,
    selectedNetworkId,
    updateNetwork,
    type INetwork,
  } from '$lib/evm-toolkit';
  import { addSelectedNetworkToMetamask } from '$lib/withdraw';

  let previousSelectedNetworkId: number = $selectedNetworkId;
  let _selectedNetwork: INetwork;
  // local copies to manage updates afterwards
  let chainIdString: string = $selectedNetwork?.chainID.toString();

  $: $selectedNetworkId, (_selectedNetwork = $selectedNetwork);
  $: _selectedNetwork, handleNetworkChange();
  $: networkSelectorOptions = $networks?.map(({ text, id }) => ({
    label: text,
    id,
  }));
  $: disableNetworkEdit = $selectedNetwork?.id !== 1;
  $: if (chainIdString && $selectedNetwork.chainID) {
    _selectedNetwork.chainID = parseInt(chainIdString);
  }

  onMount(() => {
    const unsubscribe = selectedNetworkId.subscribe(id => {
      if (id !== previousSelectedNetworkId) {
        previousSelectedNetworkId = id;
        if ($connected && $selectedAccount) {
          void addSelectedNetworkToMetamask();
        }
      }
    });
    return () => unsubscribe();
  });

  function handleNetworkChange() {
    updateNetwork(_selectedNetwork);
  }
</script>

<network-settings-component class="flex flex-col space-y-4">
  {#if _selectedNetwork}
    <Select options={networkSelectorOptions} bind:value={$selectedNetworkId} />
    <div class="flex flex-col space-y-2">
      <Input
        id="hornetEndpoint"
        label="Hornet API endpoint"
        bind:value={_selectedNetwork.apiEndpoint}
        disabled={disableNetworkEdit}
        stretch
      />
      <Input
        id="faucetEndpoint"
        label="Faucet API endpoint"
        bind:value={_selectedNetwork.faucetEndpoint}
        disabled={disableNetworkEdit}
        stretch
      />
      <Input
        id="chainId"
        label="Chain ID"
        type={InputType.Number}
        bind:value={chainIdString}
        disabled={disableNetworkEdit}
        stretch
      />
      <Input
        id="networkUrl"
        label="Network URL"
        bind:value={_selectedNetwork.networkUrl}
        disabled={disableNetworkEdit}
        stretch
      />
      <Input
        id="blockExplorer"
        label="Block Explorer"
        bind:value={_selectedNetwork.blockExplorer}
        disabled={disableNetworkEdit}
        stretch
      />
      <Input
        id="chainAddress"
        label="Chain Address"
        bind:value={_selectedNetwork.chainAddress}
        disabled={disableNetworkEdit}
        stretch
      />
    </div>
  {:else}
    <span>Loading Network Configuration...</span>
  {/if}
</network-settings-component>

<style lang="scss">
</style>
