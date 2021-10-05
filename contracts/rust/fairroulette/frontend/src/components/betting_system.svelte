<script lang="ts">
  import { BarSelector, MultipleSelector } from '../components';
  import { placeBet } from '../lib/app';
  import { placingBet, round, balance, requestBet } from '../lib/store';
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
  <div class="betting-panel">
    <div class="step-title">
      {`Step ${step} of 2`}
    </div>
    <div class="selector">
      {#if step === Step.NumberChoice}
        <MultipleSelector
          disabled={$placingBet || $round.betPlaced}
          onClick={onNumberClick}
        />
      {/if}
      {#if step === Step.AmountChoice}
        <BarSelector bind:value={betAmount} />
      {/if}
    </div>
  </div>

  <div class="betting-actions">
    <Button
      label="Back"
      disabled={$placingBet}
      onClick={() => {
        step === Step.NumberChoice
          ? ($requestBet = false)
          : (step = Step.NumberChoice);
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
        : $balance < 1n ||
          $round.betAmount === 0n ||
          $placingBet ||
          $round.betPlaced}
      onClick={() => {
        if (step === Step.AmountChoice) {
          placeBet();
          resetBar();
        } else if (step === Step.NumberChoice) {
          step = Step.AmountChoice;
        }
      }}
      loading={$placingBet}
    />
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
    .betting-actions {
      display: flex;
    }
    .betting-panel {
      position: absolute;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      .step-title {
        color: white;
      }
    }
  }
</style>
