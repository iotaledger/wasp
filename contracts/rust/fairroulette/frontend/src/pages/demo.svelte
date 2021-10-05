<script lang="ts">
  import { onMount } from 'svelte';
  import {
    BalancePanel,
    Button,
    LogsPanel,
    PlayersPanel,
    Roulette,
    State,
    WalletPanel,
  } from '../components';
  import Animation from '../components/animation.svelte';
  import ToastContainer from '../components/toast_container.svelte';
  import { createNewAddress, initialize } from '../lib/app';
  import {
    balance,
    fundsRequested,
    isAWinnerPlayer,
    newAddressNeeded,
    requestBet,
    round,
    showAddFunds,
  } from '../lib/store';

  onMount(initialize);

  $: if ($balance > 0n) {
    fundsRequested.set(true);
    newAddressNeeded.set(true);
    showAddFunds.set(false);
  } else if ($balance === 0n && $newAddressNeeded && $round.betPlaced) {
    createNewAddress();
    newAddressNeeded.set(false);
  }
</script>

<div class="container">
  {#if $isAWinnerPlayer}
    <div class="animation">
      <Animation animation="win" loop={false} destroyWhenFinished />
    </div>
  {/if}
  <div class="layout_state">
    <div class="balance">
      <BalancePanel />
    </div>
    <div class="wallet">
      <WalletPanel />
    </div>
    <div class="roulette_state">
      <State />
    </div>
  </div>
  <div class="layout_roulette">
    <div class="roulette_game">
      <Roulette />
      {#if !$requestBet && $balance > 0n}
        <Button
          disabled={$round.betPlaced}
          onClick={() => ($requestBet = true)}
          label={$round.active ? 'Join the game' : 'Choose your bet'}
        />
      {/if}
    </div>

    <div class="players">
      <PlayersPanel />
    </div>
    <div class="logs">
      <LogsPanel />
    </div>
  </div>
  <ToastContainer />
</div>

<style lang="scss">
  .animation {
    position: absolute;
    z-index: 1;
  }
  .layout_state {
    display: flex;
    flex-direction: column;
    position: relative;
    @media (min-width: 1024px) {
      flex-direction: row-reverse;
      justify-content: space-between;
      margin-top: 24px;
    }
    .wallet {
      @media (min-width: 1024px) {
        width: 25%;
      }
    }
    .roulette_state {
      margin-top: 40px;
      @media (min-width: 1024px) {
        margin-top: 0;
        position: absolute;
        left: 50%;
        transform: translateX(-50%);
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
      margin-bottom: 300px;
    }
    .players {
      height: calc(100vh - 650px);
      position: relative;
      min-height: 400px;

      @media (min-width: 1024px) {
        width: 25%;
        height: calc(100vh - 450px);
        margin-top: 32px;
      }
    }
    .roulette_game {
      max-width: max-content;
      margin: 0 auto;
      @media (min-width: 1024px) {
        position: absolute;
        top: -50px;
        left: 50%;
        transform: translateX(-50%);
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
    .logs {
      height: calc(100vh - 650px);
      position: relative;
      min-height: 400px;
      margin-bottom: 132px;
      @media (min-width: 1024px) {
        margin-bottom: 0;
        width: 25%;
        height: calc(100vh - 450px);
        margin-top: 32px;
      }
    }
  }
</style>
