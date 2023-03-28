<script lang="ts">
  import { connected, selectedAccount } from 'svelte-web3';
  import { get } from 'svelte/store';

  import { Button, Input, Select } from '$components';

  import { InputType } from '$lib/common';
  import {
    networks,
    selectedNetwork,
    selectedNetworkId,
    updateNetwork,
    type INetwork,
  } from '$lib/evm-toolkit';
  import { NotificationType, showNotification } from '$lib/notification';
  import { closePopup, PopupId } from '$lib/popup';
  import { connectToWallet } from '$lib/withdraw';

  let busy: boolean = false;

  let _selectedNetwork: INetwork = get(selectedNetwork);
  let _selectedNetworkId: number = get(selectedNetworkId);

  $: disableNetworkEdit = _selectedNetwork?.id !== 1;

  $: networkSelectorOptions = $networks?.map(({ text, id }) => ({
    label: text,
    id,
  }));

  $: _selectedNetworkId,
    (_selectedNetwork = $networks.find(
      network => network.id === _selectedNetworkId,
    ));

  async function handleSave(): Promise<void> {
    const _closePopup = () => {
      closePopup(PopupId.Settings);
    };
    selectedNetworkId.set(_selectedNetworkId);
    updateNetwork(_selectedNetwork);
    if ($connected && $selectedAccount) {
      try {
        busy = true;
        await connectToWallet();
        _closePopup();
      } catch (e) {
        showNotification({
          type: NotificationType.Error,
          message: e,
        });
        console.error(e);
      } finally {
        busy = false;
      }
    } else {
      _closePopup();
    }
  }
</script>

<network-settings-component class="flex flex-col space-y-4">
  {#if _selectedNetwork}
    <Select options={networkSelectorOptions} bind:value={_selectedNetworkId} />
    <form class="flex flex-col space-y-2">
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
        bind:value={_selectedNetwork.chainID}
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
    </form>
    <Button title="Save" busyMessage="Saving..." onClick={handleSave} {busy} />
  {:else}
    <span>Loading Network Configuration...</span>
  {/if}
</network-settings-component>

<style lang="scss">
</style>
