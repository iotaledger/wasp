<script lang="ts">
  import { BarSelector, MultipleSelector } from '../components';
  import { BettingStep, placeBet } from '../lib/app';
  import {
    placingBet,
    round,
    balance,
    requestBet,
    bettingStep,
  } from '../lib/store';
  import Button from './button.svelte';

  let betAmount: number = 1;

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
    <Button
      label="Back"
      disabled={$placingBet}
      onClick={() => {
        $bettingStep === BettingStep.NumberChoice
          ? ($requestBet = false)
          : ($bettingStep = BettingStep.NumberChoice);
      }}
    />
    <Button
      label={$bettingStep === BettingStep.NumberChoice
        ? 'Next'
        : $placingBet
        ? 'Placing bet...'
        : 'Place bet'}
      disabled={$bettingStep === BettingStep.NumberChoice
        ? !$round.betSelection
        : $balance < 1n ||
          $round.betAmount === 0n ||
          $placingBet ||
          $round.betPlaced}
      onClick={() => {
        if ($bettingStep === BettingStep.AmountChoice) {
          placeBet();
          resetBar();
        } else if ($bettingStep === BettingStep.NumberChoice) {
          $bettingStep = BettingStep.AmountChoice;
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
