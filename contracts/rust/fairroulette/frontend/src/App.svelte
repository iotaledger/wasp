<script lang="ts">
  export const name = 'app';

  import { onMount } from 'svelte';
  import config, { chainId } from '../config.dev';
  import { BettingSystem, Panel, Roulette } from './components';
  import State from './components/state.svelte';
  import type { Bet } from './fairroulette_client';
  import { FairRouletteService } from './fairroulette_client';
  import type { IBalancePanel } from './models/IBalancePanel';
  import { BALANCE_PANEL_TYPE } from './models/IBalancePanel';
  import type { IEntriesPanel } from './models/IEntriesPanel';
  import { ENTRIES_PANEL_TYPE } from './models/IEntriesPanel';
  import { LOG_ENTRIES_TYPE } from './models/ILogEntries';
  import { PLAYER_ENTRIES_TYPE } from './models/IPlayerEntries';
  import type { IState } from './models/IState';
  import type { IWalletPanel } from './models/IWalletPanel';
  import { WALLET_PANEL_TYPE } from './models/IWalletPanel';
  import {
    address,
    addressIndex,
    balance,
    keyPair,
    seed,
    seedString,
    requestingFunds,
    logs,
  } from './store';
  import {
    BasicClient,
    Colors,
    PoWWorkerManager,
    WalletService,
  } from './wasp_client';
  import { Base58 } from './wasp_client/crypto/base58';
  import { Seed } from './wasp_client/crypto/seed';

  let fundsUpdaterHandle;

  let client: BasicClient;
  let walletService: WalletService;
  let fairRouletteService: FairRouletteService;

  const powManager: PoWWorkerManager = new PoWWorkerManager();

  const view = {
    timestamp: 0,
    isWorking: false,
    showController: true,

    round: {
      active: false,
      number: 0n,
      eventList: [] as string[],
      betSelection: 1,
      winningNumber: 0n,
      startedAt: 0,
    },
  };

  // Panels located around roulette
  let walletPanel: IWalletPanel = {
    type: WALLET_PANEL_TYPE,
    data: [
      {
        eyebrow: 'Your seed',
        label: '-',
      },
      {
        eyebrow: 'Your address',
        label: '-',
      },
    ],
  };

  $: $seedString,
    $address,
    (walletPanel.data = [
      {
        eyebrow: 'Your seed',
        label: $seedString,
      },
      {
        eyebrow: 'Your address',
        label: $address,
      },
    ]);

  let balancePanel: IBalancePanel = {
    type: BALANCE_PANEL_TYPE,
    data: {
      eyebrow: 'Your balance',
      label: '-',
    },
    buttons: [
      {
        label: 'Request funds',
        onClick: sendFaucetRequest,
      },
    ],
  };

  $: $balance,
    (balancePanel.data = {
      eyebrow: 'Your balance',
      label: `${$balance.toString()}i`,
    });

  const PLAYERS_PANEL: IEntriesPanel = {
    type: ENTRIES_PANEL_TYPE,
    title: 'Players',
    ordered: true,
    entries: {
      type: PLAYER_ENTRIES_TYPE,
      data: [
        {
          address: 'address1',
          fields: [
            {
              label: 'Bet:',
              value: '1000i',
            },
            {
              label: 'W/L:',
              value: '600i',
            },
          ],
        },
        {
          address: 'address2',
          fields: [
            {
              label: 'Bet:',
              value: '1050i',
            },
            {
              label: 'W/L:',
              value: '200i',
            },
          ],
        },
      ],
    },
  };

  let logsPanel: IEntriesPanel = {
    type: ENTRIES_PANEL_TYPE,
    title: 'Logs',
    ordered: true,
    entries: {
      type: LOG_ENTRIES_TYPE,
      data: [
        {
          tag: 'Round',
          timestamp: '11:24:11',
          // label: '1022 Mi',
          description: 'Page loading...',
        },
        {
          tag: 'Site',
          timestamp: '11:24:11',
          description: 'Page loading...',
        },
      ],
    },
  };

  $: $logs, (logsPanel.entries.data = $logs);

  const INFORMATION_STATE: IState = {
    title: 'Start game',
    subtitle: 'This is a subtitle',
    description: 'The round starts in 50 seconds.',
  };

  // Entrypoint
  async function initialize() {
    log('Page', 'Loading');

    if (config.seed) {
      $seed = Base58.decode(config.seed);
    } else {
      $seed = Seed.generate();
    }

    client = new BasicClient({
      GoShimmerAPIUrl: config.goshimmerApiUrl,
      WaspAPIUrl: config.waspApiUrl,
      SeedUnsafe: $seed,
    });

    fairRouletteService = new FairRouletteService(client, config.chainId);
    walletService = new WalletService(client);

    powManager.load('/build/pow.worker.js');

    subscribeToRouletteEvents();
    setAddress($addressIndex);
    updateFunds();

    startFundsUpdater();

    // The best solution would be to call all of them in parallel. This is currently not possible.
    // As those requests can fail in certain cases, we need to wrap them in exception handlers,
    // to make sure that the other requests are being sent and that the page properly loads.
    const requests = [
      () =>
        fairRouletteService
          .getRoundStatus()
          .then((x) => (view.round.active = x == 1)),
      () =>
        fairRouletteService
          .getRoundNumber()
          .then((x) => (view.round.number = x)),
      () =>
        fairRouletteService
          .getLastWinningNumber()
          .then((x) => (view.round.winningNumber = x)),
      () =>
        fairRouletteService
          .getRoundStartedAt()
          .then((x) => (view.round.startedAt = x)),
    ];

    for (let request of requests) {
      await request().catch((e) => log('Error', e.message));
    }

    log('Page', 'Loaded');

    /**
     * ChainID => address
     * metadata: IOnLedgerRequest
     * transfer: {color:values}
     */

    //await walletService.sendOnLedgerRequest($address, chainId);

    //   const basicWallet = new BasicWallet(client);

    await walletService.sendOnLedgerRequest($address, chainId);
  }

  onMount(initialize);
  // /Entrypoint

  function log(tag: string, description: string) {
    // view.round.eventList.push(`${new Date().toLocaleTimeString()} | ${text}`);
    logs.set([
      ...$logs,
      {
        tag: tag,
        description: description,
        timestamp: new Date().toLocaleTimeString(),
      },
    ]);
  }

  function setAddress(index: number) {
    $addressIndex = index;

    $address = Seed.generateAddress($seed, $addressIndex);
    $keyPair = Seed.generateKeyPair($seed, $addressIndex);
  }

  function createNewAddress() {
    $addressIndex++;
    setAddress($addressIndex);
  }

  async function updateFunds() {
    let _balance = 0n;
    try {
      view.timestamp = Date.now() / 1000;
      _balance = await walletService.getFunds(
        $address,
        Colors.IOTA_COLOR_STRING
      );
    } catch (ex) {}
    balance.set(_balance);
  }

  function startFundsUpdater() {
    if (fundsUpdaterHandle) {
      fundsUpdaterHandle = clearInterval(fundsUpdaterHandle);
    }

    fundsUpdaterHandle = setInterval(updateFunds, 1000);
  }

  async function placeBet() {
    view.isWorking = true;
    try {
      await fairRouletteService.placeBet(
        $keyPair,
        view.round.betSelection,
        1234
      );
    } catch (ex) {
      log('Round', ex.message);
    }
    view.isWorking = false;
  }

  async function sendFaucetRequest() {
    requestingFunds.set(true);

    const faucetRequestResult = await walletService.getFaucetRequest($address);

    // In this example a difficulty of 20 is enough, might need a retune for prod to 21 or 22
    faucetRequestResult.faucetRequest.nonce =
      await powManager.requestProofOfWork(20, faucetRequestResult.poWBuffer);

    try {
      await client.sendFaucetRequest(faucetRequestResult.faucetRequest);
    } catch (ex) {
      log('Round', ex.message);
    }
    requestingFunds.set(false);
  }

  // To make sure the function gets called every second, we require that date.Now() is put in as a parameter to rely on sveltes change listener.
  function calculateRoundLengthLeft(timestamp: number) {
    const roundStarted = view.round.startedAt;

    if (roundStarted == 0) {
      return 0;
    }

    const diff = timestamp - roundStarted;

    // TODO: Explain.
    const executionCompensation = 5;
    const roundTimeLeft = Math.round(
      fairRouletteService.roundLength + executionCompensation - diff
    );

    if (roundTimeLeft <= 0) {
      return 0;
    }
    return roundTimeLeft;
  }

  function subscribeToRouletteEvents() {
    fairRouletteService.on('roundStarted', (timestamp) => {
      view.round.active = true;
      view.round.startedAt = timestamp;
      log('Round', 'Started');
    });

    fairRouletteService.on('roundStopped', () => {
      view.round.active = false;
      log('Round', 'Ended');
    });

    fairRouletteService.on('roundNumber', (roundNumber: bigint) => {
      view.round.number = roundNumber;
      log('Round', `Current round number: ${roundNumber}`);
    });

    fairRouletteService.on('winningNumber', (winningNumber: bigint) => {
      view.round.winningNumber = winningNumber;
      log('Round', `The winning number was: ${winningNumber}`);
    });

    fairRouletteService.on('betPlaced', (bet: Bet) => {
      log(
        'Bet',
        `Bet placed from ${bet.better} on ${bet.betNumber} with ${bet.amount}`
      );
    });

    fairRouletteService.on('payout', (bet: Bet) => {
      log('Win', `Payout for ${bet.better} with ${bet.amount}`);
    });
  }

  function isBroke(balance: bigint) {
    return balance < 200;
  }

  function isWealthy(balance: bigint) {
    return balance > 200;
  }
</script>

<main>
  <div class="header">
    <div class="header-logo">
      <img src="iota-roulette.svg" alt="iota-logo-roulette" />
    </div>
  </div>
  <div class="layout_state">
    <div class="balance">
      <Panel {...balancePanel} />
    </div>
    <div class="wallet">
      <Panel {...walletPanel} />
    </div>
    <div class="roulette_state">
      <State {...INFORMATION_STATE} />
    </div>
  </div>
  <div class="layout_roulette">
    <div class="roulette">
      <Roulette mode="GAME_STARTED" />
      <BettingSystem />
    </div>
    <div class="players">
      <Panel {...PLAYERS_PANEL} />
    </div>
    <div class="logs">
      <Panel {...logsPanel} />
    </div>
  </div>
  <!-- <div class="roulette">
    <img src="roulette_background.svg" alt="roulette" />
    <img src="2.svg" alt="" />
  </div> -->

  <!-- {#if view.isWorking}GENERAL
    <div class="loading_dim">
      <div class="loading_wrapper">
        <div class="loading_logo" />
        <div class="loading_text">Loading and working</div>
      </div>
    </div>
  {/if}

  {#if view.showController}
    <div class="welcome_screen">
      <div class="request_funds" class:disabled={isWealthy(view.balance)}>
        <div class="request_funds_text">1 | Request funds</div>
        <div class="request_funds_image" />
        <button
          class="request_funds_button"
          disabled={isWealthy(view.balance)}
          on:click={() => sendFaucetRequest()}>Request funds</button
        >
      </div>

      <Panel type="general" data={[{ eyebrow: "hola", label: "adios" }]} />
      <div class="place_bet" class:disabled={isBroke(view.balance)}>
        <div class="place_bet_text">2 | Place a bet</div>
        <div class="place_bet_image" />

        <div class="bet_selection">
          <input
            type="radio"
            bind:group={view.round.betSelection}
            name="bet_number"
            value="1"
            checked
          />
          <label for="1">1</label>

          <input
            type="radio"
            bind:group={view.round.betSelection}
            name="bet_number"
            value="2"
          />
          <label for="2">2</label>

          <input
            type="radio"
            bind:group={view.round.betSelection}
            name="bet_number"
            value="3"
          />
          <label for="3">3</label>

          <input
            type="radio"
            bind:group={view.round.betSelection}
            name="bet_number"
            value="4"
          />
          <label for="4">4</label>

          <input
            type="radio"
            bind:group={view.round.betSelection}
            name="bet_number"
            value="5"
          />
          <label for="5">5</label>
        </div>

        <button
          class="submit_bet_button"
          on:click={() => placeBet()}
          disabled={isBroke(view.balance)}>Submit bet</button
        >
      </div>
    </div>

    <div class="welcome_screen" />
  {/if}

  <div class="content">
    <div class="wheel_container">
      <Roulette width="300" height="300" spin={view.round.active} />

      <div class="wheel_status">
        <div>Round {view.round.active ? "started" : "stopped"}</div>
        <div>Round Nr. {view.round.number}</div>
        <div>
          Winning Nr. {view.round.active
            ? "Undecided"
            : view.round.winningNumber}
        </div>
        <div>
          Round TS: {view.round.startedAt}
        </div>
        <div>
          Remaining seconds: {calculateRoundLengthLeft(view.timestamp)}
        </div>
      </div>
    </div>

    <div class="log">
      <textarea rows="4" cols="50">{view.round.eventList.join("\n")}</textarea>
    </div>
  </div> -->
</main>

<style>
  main {
    width: 100%;
    height: 100%;
  }

  div.header {
    width: 100%;
    background-color: rgba(72, 87, 118, 0.2);
    height: 50px;
    padding: 16px;
  }
  .layout_state {
    display: grid;
    grid-template-rows: repeat(1fr);
    grid-column-gap: 0px;
    grid-row-gap: 0px;
    margin-top: 48px;
  }
  @media (min-width: 1024px) {
    .layout_state {
      display: grid;
      grid-template-columns: 1fr 2fr 1fr;
      grid-template-rows: auto auto;
      gap: 20px 20px;
      grid-template-areas:
        'aside-1 first aside-2'
        'aside-1 last aside-2';
    }
  }
  .roulette_state {
    margin-top: 40px;
  }
  @media (min-width: 1024px) {
    .roulette_state {
      margin-top: 0;
    }
  }
  @media (min-width: 1024px) {
    .wallet {
      grid-area: aside-1;
      margin-left: 60px;
    }
  }
  @media (min-width: 1300px) {
    .wallet {
      margin-left: 120px;
    }
  }
  @media (min-width: 1024px) {
    .balance {
      grid-area: aside-2;
      margin-right: 60px;
    }
  }
  @media (min-width: 1300px) {
    .balance {
      margin-right: 120px;
    }
  }
  .layout_roulette {
    display: grid;
    grid-template-rows: repeat(1fr);
    grid-column-gap: 0px;
    grid-row-gap: 0px;
    margin-top: 48px;
  }
  @media (min-width: 1024px) {
    .layout_roulette {
      display: grid;
      grid-template-columns: 1fr 2fr 1fr;
      grid-template-rows: auto auto;
      gap: 20px 20px;
      grid-template-areas:
        'aside-1 first aside-2'
        'aside-1 last aside-2';
    }
  }
  .roulette {
    margin-bottom: 100px;
  }
  @media (min-width: 1024px) {
    .roulette {
      margin-bottom: 0;
    }
  }
  @media (min-width: 1024px) {
    .players {
      grid-area: aside-1;
      margin-left: 60px;
    }
  }
  @media (min-width: 1300px) {
    .players {
      margin-left: 120px;
    }
  }

  @media (min-width: 1024px) {
    .logs {
      grid-area: aside-2;
      margin-right: 60px;
    }
  }
  @media (min-width: 1300px) {
    .logs {
      margin-right: 120px;
    }
  }
</style>
