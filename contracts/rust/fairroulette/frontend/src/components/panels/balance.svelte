<script lang="ts">
  import type { IPanelDataItem } from './../../models/IPanelDataItem';
  import type { IButton } from './../../models/IButton';
  import Button from '../button.svelte';

  export let data: IPanelDataItem;
  export let buttons: IButton[] = [];
</script>

<div class="panel">
  <div class="balance">
    <div class="eyebrow">{data.eyebrow}</div>
    {#if data.label}
      <div class="label">{data.label}</div>
    {/if}
  </div>
  <div class="value-button">
    {#if buttons}
      {#each buttons as button}
        <div>
          <Button {...button} />
        </div>
      {/each}
    {/if}
  </div>
</div>

<style lang="scss">
  .panel {
    align-items: center;
    display: grid;
    grid-template-rows: 1fr;
    grid-template-columns: 1fr 2fr;
    @media (min-width: 1024px) {
      text-align: center;
      grid-template-rows: 1fr 1fr;
      grid-template-columns: 1fr;
    }
    .balance {
      .label {
        font-family: 'Metropolis Bold';
        font-weight: 600;
        font-size: 24px;
        line-height: 115%;
        letter-spacing: 0.5px;
        color: var(--white);
        @media (min-width: 1024px) {
          font-size: 32px;
          line-height: 120%;
        }
      }
    }
    .value-button div {
      margin-bottom: 8px;
      @media (min-width: 1024px) {
        margin: 0 auto;
        margin-bottom: 8px;
      }
    }
    .value-button div:not(:last-child) {
      margin-bottom: 8px;
    }
  }
</style>
