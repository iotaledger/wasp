<script lang="ts">
  export const name = 'app';

  import { onMount } from 'svelte';
  import {
  BalancePanel,
  BettingSystem,
  LogsPanel,
  PlayersPanel,
  Roulette
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
      </div>
      <div class="bet_system">
        <BettingSystem />
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
      display: grid;
      grid-template-rows: repeat(1fr);
      grid-column-gap: 0px;
      grid-row-gap: 0px;
      @media (min-width: 1024px) {
        display: grid;
        grid-template-columns: 1fr 2fr 1fr;
        grid-template-rows: auto auto;
        gap: 20px 20px;
        grid-template-areas:
          'aside-1 first aside-2'
          'aside-1 last aside-2';
        margin-top: 48px;
      }
      .wallet {
        @media (min-width: 1024px) {
          grid-area: aside-1;
        }
      }
      .balance {
        @media (min-width: 1024px) {
          grid-area: aside-2;
        }
      }

      .roulette_state {
        margin-top: 40px;
        @media (min-width: 1024px) {
          margin-top: 0;
        }
      }
    }
    .layout_roulette {
      display: grid;
      grid-template-rows: repeat(1fr);
      grid-column-gap: 0px;
      grid-row-gap: 0px;
      margin-top: 48px;
      @media (min-width: 1024px) {
        display: grid;
        grid-template-columns: 1fr 2fr 1fr;
        grid-template-rows: auto auto;
        gap: 20px 20px;
        grid-template-areas:
          'aside-1 first aside-2'
          'aside-1 last aside-2';
      }

      .roulette_game {
        max-height: fit-content;
        max-width: fit-content;
        margin: 0 auto;
      }
      .bet_system {
        margin-top: 40px;
        margin-bottom: 100px;
      }
      .players {
        height: calc(100vh - 450px);
        @media (min-width: 1024px) {
          grid-area: aside-1;
        }
      }
      .logs {
        height: calc(100vh - 450px);
        @media (min-width: 1024px) {
          grid-area: aside-2;
        }
      }
    }
  }
</style>
