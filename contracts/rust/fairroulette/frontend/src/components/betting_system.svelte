<script lang="ts">
  import { BarSelector, MultipleSelector } from './../components';
  import { placeBet, sendFaucetRequest } from './../lib';
  import { round, balance, requestingFunds } from './../store';
  import Button from './button.svelte';
</script>

<div class="betting-system">
  <MultipleSelector />
  <div>
    <BarSelector />
    <div class="bet-button">
      {#if $balance > 1n}
        <Button
          label="Place bet"
          disabled={$round.betSelection === undefined}
          onClick={placeBet}
        />
      {:else}
        <Button
          label={$requestingFunds ? 'Requesting...' : 'Request funds'}
          onClick={sendFaucetRequest}
          disabled={$requestingFunds}
          loading={$requestingFunds}
        />
      {/if}
    </div>
  </div>
</div>

<style lang="scss">
  .betting-system {
    display: flex;
    flex-direction: column;
    align-content: center;
    flex-wrap: wrap;
    gap: 30px;
    @media (min-width: 1024px) {
      flex-direction: row;
      justify-content: center;
      gap: 60px;
      align-items: flex-end;
    }
    .bet-button {
      margin-top: 24px;
      display: flex;
      justify-content: center;
    }
  }
</style>
