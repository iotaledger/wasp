<script lang="ts">
  import { GOSHIMMER_ADDRESS_EXPLORER_URL } from '../../lib/app';
  import { address, seedString } from '../../lib/store';
</script>

<div class="panel wallet">
  <div class="wallet-data">
    <div class="eyebrow">Your seed</div>
    <div class="data">{$seedString ?? '-'}</div>
  </div>
  <div class="wallet-data">
    <div class="eyebrow">Your address</div>
    {#if $address}
      <a
        class="data"
        href={`${GOSHIMMER_ADDRESS_EXPLORER_URL}/${$address}`}
        target="_blank"
        rel="noopener noreferrer">{$address}</a
      >
    {:else}
      <div class="data">-</div>
    {/if}
  </div>
</div>

<style lang="scss">
  .wallet {
    border-bottom: 1px solid transparent;
    padding: 16px 24px;
    @media (min-width: 1024px) {
      border-bottom: 1px solid rgba(255, 255, 255, 0.08);
      padding: 10px 15px;
      height: auto;
    }
    .wallet-data {
      display: grid;
      grid-template-rows: 1fr;
      grid-template-columns: 1fr 2fr;
      @media (min-width: 1024px) {
        grid-template-columns: 1fr;
      }
      .data {
        width: 100%;
        font-size: 14px;
        letter-spacing: 0.5px;
        color: var(--gray-3);
        line-height: 150%;
        word-break: break-word;
      }
    }
    .wallet-data:not(:last-child) {
      margin-bottom: 4px;

      @media (min-width: 1024px) {
        margin-bottom: 16px;
      }
    }
  }
</style>
