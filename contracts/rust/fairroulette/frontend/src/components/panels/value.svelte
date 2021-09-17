<script lang="ts">
  import type { IPanelDataItem } from './../../models/IPanelDataItem';
  import type { IButton } from './../../models/IButton';
  import Button from '../button.svelte';

  export let data: IPanelDataItem;
  export let buttons: IButton[];
</script>

<div class="value-panel">
  <div class="balance">
    <div class="eyebrow">{data.eyebrow}</div>
    {#if data.label}
      <div class="label">{data.label}</div>
    {/if}
  </div>
  <div class="value-button">
    {#if buttons}
      {#each buttons as button}
        <Button {...button} />
      {/each}
    {/if}
  </div>
</div>

<style lang="scss">
  .value-panel {
    background: rgba(6, 16, 35, 0.2);
    backdrop-filter: blur(20px);
    border-top: 1px solid rgba(255, 255, 255, 0.08);
    padding-top: 8px;
    padding-bottom: 8px;
    width: 100%;
    display: flex;
    justify-content: space-between;
    align-items: center;
    @media (min-width: 1024px) {
      background: #141e31;
      border: 1px solid rgba(255, 255, 255, 0.12);
      border-radius: 12px;
      width: 275px;
      padding-top: 8px;
      padding-bottom: 8px;
      display: block;
      text-align: center;
    }
    .balance {
      margin-left: 24px;
      @media (min-width: 1024px) {
        margin-left: 0;
        margin-top: 8px;
        margin-bottom: 8px;
      }
      .eyebrow {
        text-transform: capitalize;
        font-weight: bold;
        font-size: 12px;
        line-height: 150%;
        letter-spacing: 0.5px;
        color: #36a1ac;
      }
      .label {
        font-family: 'Metropolis Bold';
        font-weight: 600;
        font-size: 24px;
        line-height: 115%;
        letter-spacing: 0.5px;
        color: #ffffff;
        @media (min-width: 1024px) {
          font-size: 32px;
          line-height: 120%;
        }
      }
    }
    .value-button {
      margin-right: 24px;
      @media (min-width: 1024px) {
        margin: 0 auto;
        margin-top: 8px;
        margin-bottom: 8px;
      }
    }
  }
</style>
