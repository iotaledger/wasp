<script lang="ts">
  import { Input, Select } from '$components';
  import { selectedNetwork, networks } from '../../store';
</script>

<network-settings-component class="flex flex-col space-y-4">
  {#if $selectedNetwork}
    <Select
      options={$networks}
      displayValue={network => network.text}
      bind:value={$selectedNetwork}
      index={$networks?.findIndex(network => network.id === $selectedNetwork.id)}
    />
    <div class="flex flex-col space-y-2">
      <Input
        id="hornetEndpoint"
        label="Hornet API endpoint"
        bind:value={$selectedNetwork.apiEndpoint}
        disabled={$selectedNetwork.id !== 1}
        stretch
      />
      <Input
        id="faucetEndpoint"
        label="Faucet API endpoint"
        bind:value={$selectedNetwork.faucetEndpoint}
        disabled={$selectedNetwork.id !== 1}
        stretch
      />
      <Input
        id="chainAddress"
        label="Chain Address"
        bind:value={$selectedNetwork.chainAddress}
        disabled={$selectedNetwork.id !== 1}
        stretch
      />
    </div>
  {:else}
    <span>Loading Network Config File...</span>
  {/if}
</network-settings-component>

<style lang="scss">
</style>
