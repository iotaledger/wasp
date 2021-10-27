<script lang="ts">
  import Button from '../button.svelte';
  import { sendFaucetRequest } from '../../lib/app';
  import { balance, placingBet, requestingFunds } from '../../lib/store';
</script>

<div class="panel">
  <div class="balance">
    <div class="eyebrow">Your balance</div>
    <div class="balance-amount">{$balance ?? '-'}i</div>
  </div>
  <div class="request-funds-button">
    <Button
      label={$requestingFunds ? 'Requesting...' : 'Request funds'}
      onClick={sendFaucetRequest}
      disabled={$requestingFunds || $placingBet || $balance > 0n}
      loading={$requestingFunds}
    />
  </div>
</div>

<style lang="scss">
  .panel {
    align-items: center;
    display: grid;
    grid-template-rows: 1fr;
    grid-template-columns: 1fr 2fr;
    gap: 12px;
    @media (min-width: 1024px) {
      text-align: center;
      grid-template-rows: 1fr 1fr;
      grid-template-columns: 1fr;
    }
    .balance {
      .balance-amount {
        font-family: 'Metropolis Bold';
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
  }
</style>
