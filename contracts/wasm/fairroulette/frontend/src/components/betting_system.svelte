<script lang="ts">
  import { BarSelector, MultipleSelector } from './selectors';
  import { placeBet } from '../lib/app';
  import { BettingStep } from '../lib/store';
  import {
    bettingStep,
    placingBet,
    resetBettingSystem,
    round,
  } from '../lib/store';
  import Button from './button.svelte';

  let betAmount: number = 1;

  $: betAmount, onBarChange();

  $: isFirstStep = $bettingStep === BettingStep.NumberChoice;
  $: isLastStep = $bettingStep === BettingStep.AmountChoice;

  function onNumberClick(number: number): void {
    let betSelection = $round?.betSelection !== number ? number : undefined;
    round.update(($round) => ({ ...$round, betSelection }));
  }

  function onBarChange() {
    $round.betAmount = BigInt(betAmount);
  }

  function onNextClick() {
    if (isLastStep) {
      placeBet();
      resetBettingSystem();
    } else {
      bettingStep.update((_step) => _step + 1);
    }
  }

  function onBackClick() {
    if (isFirstStep) {
      resetBettingSystem();
    } else {
      bettingStep.update((_step) => _step - 1);
    }
  }
</script>

<div class="betting-system">
  <div class="betting-panel">
    <div class="step-title">
      {`Step ${$bettingStep} of 2`}
    </div>
    <div class="selector">
      {#if $bettingStep === BettingStep.NumberChoice}
        <MultipleSelector
          disabled={$placingBet || $round.betPlaced}
          onClick={onNumberClick}
        />
      {/if}
      {#if $bettingStep === BettingStep.AmountChoice}
        <BarSelector bind:value={betAmount} />
      {/if}
    </div>
  </div>

  <div class="betting-actions">
    <div>
      <Button
        label="Back"
        secondary
        disabled={$placingBet}
        onClick={onBackClick}
      />
    </div>
    <div>
      <Button
        label={isFirstStep ? 'Next' : 'Place bet'}
        disabled={(isFirstStep && !$round.betSelection) ||
          (isLastStep && !$round.betAmount)}
        onClick={onNextClick}
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
    @media (min-width: 1024px) {
      flex-direction: row;
      justify-content: center;
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
      margin-top: 24px;
      justify-content: space-evenly;
      @media (min-width: 1024px) {
        margin-top: 32px;
      }

      div {
        margin: 0 16px;
        min-width: 160px;
      }
    }
    .betting-panel {
      position: absolute;
      top: 40%;
      left: 50%;
      transform: translate(-50%, -50%);
      height: 200px;
      @media (min-width: 1024px) {
        height: 230px;
      }
      .step-title {
        font-weight: 600;
        font-size: 20px;
        line-height: 140%;
        text-align: center;
        letter-spacing: 0.02em;
        color: #909fbe;
        margin-bottom: 40px;
        @media (min-width: 1024px) {
          margin-bottom: 60px;
        }
      }
    }
  }
</style>
