<script lang="ts">
  import { BarSelector, MultipleSelector } from '../components';
  import { placeBet } from '../lib/app';
  import { placingBet, round } from '../lib/store';
  import Button from './button.svelte';

  let betAmount: number = 1;

  enum Step {
    NumberChoice = 1,
    AmountChoice = 2,
  }
  let step: Step = Step.NumberChoice;

  $: betAmount, onBarChange();

  function onNumberClick(number): void {
    let betSelection = $round?.betSelection !== number ? number : undefined;
    round.update(($round) => ({ ...$round, betSelection }));
  }

  function onBarChange() {
    $round.betAmount = BigInt(betAmount);
  }

  function resetBar() {
    betAmount = 1;
  }

</script>

<div class="betting-system">
  {#if step === Step.NumberChoice}
    <MultipleSelector
      disabled={$placingBet || $round.betPlaced}
      onClick={onNumberClick}
    />
  {/if}
  {#if step === Step.AmountChoice}
    <BarSelector bind:value={betAmount} />
  {/if}
  <Button
    label="Back"
    disabled={step === Step.NumberChoice}
    onClick={() => {
      step = Step.NumberChoice;
    }}
  />
  <Button
    label={step === Step.NumberChoice
      ? 'Next'
      : $placingBet
      ? 'Placing bet...'
      : 'Place bet'}
    disabled={step === Step.NumberChoice
      ? !$round.betSelection
      : $round.betAmount === 0n || $placingBet || $round.betPlaced}
    onClick={() => {
      if (step === Step.AmountChoice) {
        console.log('Place beeeet');
        placeBet();
        resetBar();
      }
      if (step === Step.NumberChoice) {
        step = Step.AmountChoice;
      }
    }}
    loading={$placingBet}
  />
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
