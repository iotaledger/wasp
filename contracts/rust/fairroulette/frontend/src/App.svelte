<script lang="ts">
  export const name = 'app';

  import { onMount } from 'svelte';
  import {
    BalancePanel,
    BettingSystem,
    LogsPanel,
    PlayersPanel,
    Roulette,
  } from './components';
  import { WalletPanel } from './components/';
  import Header from './components/header.svelte';
  import State from './components/state.svelte';
  import { initialize } from './lib';
  import type { IState } from './models/IState';

  const INFORMATION_STATE: IState = {
    title: 'Start game',
    subtitle: 'This is a subtitle',
    description: 'The round starts in 50 seconds.',
  };

  onMount(initialize);
</script>

<main>
  <Header />

  <div class="container">
    <div class="layout_state">
      <div class="balance">
        <BalancePanel />
      </div>
      <div class="wallet">
        <WalletPanel />
      </div>
      <div class="roulette_state">
        <State phase="GAME_STARTED" />
      </div>
    </div>
    <div class="layout_roulette">
      <div class="roulette_game">
        <Roulette mode="GAME_STARTED" />
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
</main>

<style lang="scss">
  .container {
    max-width: 1200px;
    margin: 0 auto;
    @media (min-width: 1024px) {
      padding: 0 24px;
    }
    @media (min-width: 2000px) {
      max-width: 1600px;
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
          left: 50%;
          transform: translateX(-50%);
        }
        .bet_system {
          margin-top: 40px;
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
  }
</style>
