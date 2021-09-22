<script lang="ts">
  import { balance } from './../../store';
  import { Button } from './../../components';
  import { sendFaucetRequest } from './../../lib';
</script>

<div class="panel">
  <div class="balance">
    <div class="eyebrow">Your balance</div>
    <div class="balance-amount">{$balance}i</div>
  </div>
  <div class="request-funds-button">
    <Button label="Request funds" onClick={sendFaucetRequest} />
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
    .request-funds-button div {
      margin-bottom: 8px;
      width: 80%;

      @media (min-width: 1024px) {
        margin: 0 auto;
        margin-bottom: 3px;
      }
    }
    .request-funds-button div:not(:last-child) {
      margin-bottom: 8px;
    }
  }
</style>
