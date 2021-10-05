<script lang="ts">
  import { BettingStep, State } from '../lib/app';
  import {
    bettingStep,
    firstTimeRequestingFunds,
    requestBet,
    round,
    timeToFinished,
  } from '../lib/store';

  let state: State;

  export let forceState: State = undefined;

  $: MESSAGES = {
    [State.Running]: {
      title: 'Game Running!',
      description: `The round ends in ${$timeToFinished ?? '...'} seconds.`,
    },
    [State.Start]: {
      title: 'Start game',
      description:
        'Press the “Choose your bet” button below and follow on-screen instructions.',
    },
    [State.AddFunds]: {
      title: 'Add funds',
      description:
        'To play, first request funds for your wallet. Those are dev-net tokens and hold no value.',
    },
    [State.AddFundsRunning]: {
      title: undefined,
      description:
        'To play, first request funds for your wallet. Those are dev-net tokens and hold no value.',
    },
    [State.ChoosingNumber]: {
      title: 'Choose a number',
      description:
        'Select a number of the roulette that you want to bet on randomly winning',
    },
    [State.ChoosingAmount]: {
      title: 'Set your amount',
      description: 'Feeling lucky? How much will you risk?',
    },
  };

  $: forceState
    ? (state = forceState)
    : (state =
        $requestBet && $bettingStep === BettingStep.NumberChoice
          ? State.ChoosingNumber
          : $requestBet && $bettingStep === BettingStep.AmountChoice
          ? State.ChoosingAmount
          : $round.active
          ? State.Running
          : !$firstTimeRequestingFunds
          ? State.AddFunds
          : State.Start);
</script>

{#if state}
  <div class="message">
    {#if MESSAGES[state].title}
      <h2 class="title">
        {MESSAGES[state].title}
      </h2>
    {/if}
    <div class="description">
      {MESSAGES[state].description}
    </div>
    <div />
  </div>
{/if}

<style lang="scss">
  .message {
    font-family: 'Metropolis Bold';
    text-align: center;
    .title {
      text-align: center;
      color: var(--white);
    }
    .subtitle {
      font-family: 'Metropolis Bold';
      font-size: 24px;
      line-height: 120%;
      letter-spacing: 0.02em;
      color: var(--mint-green);
      margin-bottom: 8px;
    }
    .description {
      padding: 16px;
      font-size: 16px;
      line-height: 150%;
      letter-spacing: 0.75px;
      color: var(--gray-5);
    }
  }
</style>
