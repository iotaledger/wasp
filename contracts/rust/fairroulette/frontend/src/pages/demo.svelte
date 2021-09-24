<script lang="ts">
  import { onMount } from 'svelte';
  import { initialize } from '../lib/app';
  import {
    balance,
    round,
    updateGameState,
    address,
    placingBet,
    newAddressNeeded,
    fundsRequested,
  } from '../lib/store';
  import {
    BalancePanel,
    BettingSystem,
    Button,
    LogsPanel,
    PlayersPanel,
    Roulette,
    State,
    WalletPanel,
  } from './../components';

  onMount(initialize);

  $: updateGameState(), $balance, $round;
</script>

<div class="container">
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
      <div class="bet_system">
        <BettingSystem />
      </div>
    </div>

    <div class="players">
      <PlayersPanel />
    </div>
    <div class="logs">
      <LogsPanel />
    </div>
  </div>
</div>

<style lang="scss">
  .simulator {
    color: white;
    margin: 48px 0;

    .sim-buttons {
      display: flex;
      justify-content: space-around;
      .sim-button {
        width: 350px;
      }
    }
  }
  .layout_state {
    display: flex;
    flex-direction: column;
    position: relative;
    @media (min-width: 1024px) {
      flex-direction: row-reverse;
      justify-content: space-between;
      margin-top: 48px;
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
      margin-top: 32px;
      margin-bottom: 300px;
    }
    .players {
      height: calc(100vh - 650px);
      position: relative;
      min-height: 400px;

      @media (min-width: 1024px) {
        width: 25%;
        height: calc(100vh - 450px);
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
      .bet_system {
        margin-top: 20px;
        margin-bottom: 100px;
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
      }
    }
  }
</style>
