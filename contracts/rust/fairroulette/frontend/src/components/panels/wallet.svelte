<script lang="ts">
  import { address, seedString } from '../../store';
</script>

<div class="panel wallet">
  <div class="wallet-data">
    <div class="eyebrow">Your seed</div>
    <div class="seed">{$seedString}</div>
  </div>
  <div class="wallet-data">
    <div class="eyebrow">Your address</div>
    <div class="address">{$address}</div>
  </div>
</div>

<style lang="scss">
  .wallet {
    @media (min-width: 1024px) {
      padding-bottom: 24px;
    }
    .wallet-data {
      display: grid;
      grid-template-rows: 1fr;
      grid-template-columns: 1fr 2fr;
      align-items: end;
      @media (min-width: 1024px) {
        grid-template-rows: 1fr 1fr;
        grid-template-columns: 1fr;
      }
      .seed,
      .address {
        width: 100%;
        font-size: 14px;
        letter-spacing: 0.5px;
        color: var(--gray-3);
        line-height: 150%;
        @media (min-width: 1024px) {
          font-weight: bold;
        }
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
