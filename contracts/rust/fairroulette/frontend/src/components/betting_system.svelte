<script lang="ts">
  import { BarSelector, MultipleSelector } from '../components';
  import { BettingStep, placeBet } from '../lib/app';
  import {
    placingBet,
    round,
    balance,
    bettingStep,
    requestBet,
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

  function resetNumber() {
    $round.betSelection = undefined;
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
          resetNumber();
          $requestBet = false;
          $bettingStep = BettingStep.NumberChoice;
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
      gap: 32px;
      width: 250px;
      margin-top: 24px;

      @media (min-width: 1024px) {
        width: 350px;
        margin-top: 32px;
      }
    }
    .betting-panel {
      position: absolute;
      top: 40%;
      left: 50%;
      transform: translate(-50%, -50%);
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
