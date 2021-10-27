<script lang="ts">
  import { onMount } from 'svelte';
  import {
    BalancePanel,
    Button,
    LogsPanel,
    PlayersPanel,
    Roulette,
    WalletPanel,
  } from '../components';
  import Animation from '../components/animation.svelte';
  import ToastContainer from '../components/toast_container.svelte';
  import { initialize, StateMessage } from '../lib/app';
  import {
    balance,
    bettingStep,
    firstTimeRequestingFunds,
    isAWinnerPlayer,
    placingBet,
    round,
    showBettingSystem,
    timeToFinished,
    requestingFunds,
    BettingStep,
  } from '../lib/store';

  export let currentRoute;
  export let params;

  let message: StateMessage;

  $: $round,
    $showBettingSystem,
    $bettingStep,
    $firstTimeRequestingFunds,
    $requestingFunds,
    updateMessage();

  $: MESSAGES = {
    [StateMessage.Running]: {
      title: 'Game Running!',
      description: `The round ends in ${$timeToFinished ?? '...'} seconds.`,
    },
    [StateMessage.Start]: {
      title: 'Start game',
      description:
        'Press the “Choose bet” button below and follow on-screen instructions.',
    },
    [StateMessage.AddFunds]: {
      title: 'Add funds',
      description:
        'To play, first request funds for your wallet. These are dev-net tokens and hold no value.',
    },
    [StateMessage.ChoosingNumber]: {
      title: 'Choose a number',
      description:
        'Select a number of the roulette that you want to bet on randomly winning',
    },
    [StateMessage.ChoosingAmount]: {
      title: 'Set your amount',
      description: 'Feeling lucky? How much will you risk?',
    },
    [StateMessage.PlacingBet]: {
      title: 'Placing Bet',
      description: 'Your bet is currently getting placed.',
      extra: 'The next round will begin in a few seconds.',
    },
  };

  onMount(initialize);

  const updateMessage = (): void => {
    if ($showBettingSystem && $bettingStep === BettingStep.NumberChoice) {
      message = StateMessage.ChoosingNumber;
    } else {
      if ($showBettingSystem && $bettingStep === BettingStep.AmountChoice) {
        message = StateMessage.ChoosingAmount;
      } else {
        if ($placingBet) {
          message = StateMessage.PlacingBet;
        } else {
          if ($round.active) {
            message = StateMessage.Running;
          } else {
            if ($firstTimeRequestingFunds) {
              message = StateMessage.AddFunds;
            } else if (!$requestingFunds) {
              message = StateMessage.Start;
            }
          }
        }
      }
    }
  };
</script>

<svelte:head>
  <title>Demo | IOTA Roulette</title>
</svelte:head>

<div class="container">
  {#if $isAWinnerPlayer}
    <div class="animation">
      <Animation animation="win" loop={false} destroyWhenFinished />
    </div>
  {/if}
  <div class="top_wrapper">
    <div class="balance">
      <BalancePanel />
    </div>
    <div class="wallet-desktop">
      <WalletPanel />
    </div>
    <div class="message">
      {#if MESSAGES[message].title}
        <h2 class="title">
          {MESSAGES[message].title}
        </h2>
      {/if}
      {#if !(message === StateMessage.Running && !$timeToFinished)}
        <div class="description">
          {MESSAGES[message].description}
          {#if message === StateMessage.PlacingBet && !$round?.active && MESSAGES[message].extra}
            {MESSAGES[message].extra}
          {/if}
        </div>
      {/if}
    </div>
  </div>
  <div class="layout_roulette">
    <div class="roulette_game">
      <Roulette />
      <div class="start_button">
        {#if !$showBettingSystem && $balance > 0n}
          <Button
            disabled={$round.betPlaced || $placingBet}
            onClick={() => showBettingSystem.set(true)}
            loading={$placingBet ||
              (!$placingBet && $round.active && $round.betPlaced)}
            label={$placingBet
              ? 'Placing bet'
              : $round.active && $round.betPlaced
              ? 'In progress'
              : 'Choose bet'}
          />
        {/if}
      </div>
      {#if $round.active && $balance === 0n && !$round.betPlaced}
        <div class="description">
          To play, first request funds for your wallet. These are dev-net tokens
          and hold no value.
        </div>
      {/if}
    </div>

    <div class="players">
      <PlayersPanel />
    </div>
    <div class="logs">
      <LogsPanel />
    </div>
  </div>
  <div class="wallet-mobile">
    <WalletPanel />
  </div>
  <ToastContainer />
</div>

<style lang="scss">
  .container {
    position: relative;
    .animation {
      position: absolute;
      z-index: 1;
      overflow: hidden;
      max-width: 100%;
      max-height: 100%;
    }
    .description {
      text-align: center;
      font-family: 'Metropolis Semi Bold';
      font-size: 16px;
      line-height: 150%;
      letter-spacing: 0.75px;
      color: var(--gray-5);
      padding: 10px 20px;
    }
    .top_wrapper {
      display: flex;
      flex-direction: column;
      position: relative;
      @media (min-width: 1024px) {
        flex-direction: row-reverse;
        justify-content: space-between;
        margin-top: 50px;
      }
      .wallet-desktop {
        width: 25%;
        display: none;
        position: relative;
        @media (min-width: 1024px) {
          display: inherit;
        }
      }
      .message {
        margin-top: 40px;
        font-family: 'Metropolis Semi Bold';
        text-align: center;
        @media (min-width: 1024px) {
          margin-top: 0;
          position: absolute;
          left: 50%;
          transform: translateX(-50%);
          width: 50%;
        }
        .title {
          text-align: center;
          color: var(--white);
        }
        .subtitle {
          font-size: 24px;
          line-height: 120%;
          letter-spacing: 0.02em;
          color: var(--mint-green);
          margin-bottom: 8px;
        }
      }
      .balance {
        @media (min-width: 1024px) {
          width: 25%;
        }
      }
    }
    .layout_roulette {
      display: flex;
      flex-direction: column;
      position: relative;
      @media (min-width: 1024px) {
        flex-direction: row;
        justify-content: space-between;
        margin-bottom: 130px;
      }
      .players,
      .logs {
        height: min-content;
        max-height: 500px;
        position: relative;
        min-height: 100px;
        display: flex;
        @media (min-width: 1024px) {
          width: 25%;
          height: calc(100vh - 400px);
          margin-top: 32px;
          max-height: 920px;
          min-height: 500px;
        }
        @media (min-width: 1135px) {
          height: calc(100vh - 350px);
        }
      }
      .roulette_game {
        width: 100%;
        max-width: 500px;
        margin: 0 auto;
        margin-bottom: 50px;
        @media (min-width: 1024px) {
          position: absolute;
          top: -50px;
          left: 50%;
          transform: translateX(-50%);
        }
        .start_button {
          width: 50%;
          margin: 0 auto;
          margin-top: 30px;
        }
        .bet_system,
        .request_button {
          margin-top: 20px;
          margin-bottom: 100px;
        }
        .request_button {
          margin-top: 32px;
          @media (min-width: 1024px) {
            padding: 0 120px;
          }
        }
      }
    }
    .wallet-mobile {
      display: inherit;
      @media (min-width: 1024px) {
        display: none;
      }
    }
  }
</style>
