<script lang="ts">
  import { Input, Select } from '$components';

  import {
    networks,
    selectedNetwork,
    selectedNetworkId,
    updateNetwork,
    type INetwork,
  } from '$lib/evm-toolkit';

  let _selectedNetwork: INetwork;
  // local copy to manage updates afterwards
  $: $selectedNetworkId, (_selectedNetwork = $selectedNetwork);
  $: _selectedNetwork, handleNetworkChange();
  $: networkSelectorOptions = $networks?.map(({ text, id }) => ({
    label: text,
    id,
  }));
  $: disableNetworkEdit = $selectedNetwork?.id !== 1;

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
