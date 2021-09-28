<script lang="ts">
  import { BarSelector, MultipleSelector } from '../components';
  import { placeBet } from '../lib/app';
  import { balance, placingBet, round } from '../lib/store';
  import Button from './button.svelte';

  let betAmount: number = 0;

  $: betAmount, onBarChange();

  function onNumberClick(number): void {
    let betSelection = $round?.betSelection !== number ? number : undefined;
    round.update(($round) => ({ ...$round, betSelection }));
  }

  function onBarChange() {
    $round.betAmount = BigInt(betAmount);
  }

  function resetBar() {
    betAmount = 0;
  }
</script>

<div class="betting-system" class:disabled={$placingBet || $round.betPlaced}>
  <MultipleSelector
    disabled={$placingBet || $round.betPlaced}
    onClick={onNumberClick}
  />
  <div>
    <BarSelector
      disabled={$placingBet || $round.betPlaced}
      bind:value={betAmount}
    />
    <div class="bet-button">
      <Button
        label={$placingBet ? 'Placing bet...' : 'Place bet'}
        disabled={!$round.betSelection ||
          $round.betAmount === 0n ||
          $placingBet ||
          $round.betPlaced}
        onClick={() => {
          placeBet();
          resetBar();
        }}
        loading={$placingBet}
      />
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
    &.disabled {
      opacity: 0.5;
    }
    .bet-button {
      margin-top: 24px;
      display: flex;
      justify-content: center;
    }
  }
</style>
