<script lang="ts">
  export const name = 'app';

  import { Base58 } from './client/crypto/base58';
  import { BasicClient, Colors } from './client/basic_client';
  import { FairRoulette } from './client/fair_roulette_service';
  import { onMount } from 'svelte';
  import { PoWWorkerManager } from './web_worker/pow_worker_manager';
  import { Seed } from './client/crypto/seed';
  import config from '../config.dev';
  import Roulette from './Roulette.svelte';
  import type { Bet } from './client/fair_roulette_service';
  import type { Buffer } from './client/buffer';
  import type { IKeyPair } from './client/crypto/models/IKeyPair';

  const chainId = config.chainId;
  const testSeed = Base58.decode('GLHjvWjQ4N4iKzPVKc1iAMLRqRhWLBhysdxNV2JsBXs6');
  const seed: Buffer = testSeed;

  let addressIndex: number = 0;
  let keyPair: IKeyPair;
  let fundsUpdaterHandle;

  const powManager = new PoWWorkerManager();

  const client: BasicClient = new BasicClient({
    GoShimmerAPIUrl: config.goshimmerApiUrl,
    WaspAPIUrl: config.waspApiUrl,
    SeedUnsafe: seed,
  });

  const fairRoulette: FairRoulette = new FairRoulette(client, chainId);

  let view = {
    seedString: Base58.encode(seed),
    address: '',
    balance: 0n,

    showController: false,
    isWorking: false,

    round: {
      active: false,
      number: 0n,
      eventList: [] as string[],
      betSelection: 0,
      winningNumber: 0n,
    },
  };

  function log(text: string) {
    view.round.eventList.push(`${new Date().toLocaleTimeString()} | ${text}`);
  }

  async function updateFunds() {
    try {
      view.balance = await client.getFunds(view.address, Colors.IOTA_COLOR_STRING);
    } catch (ex) {
      console.log(ex);
      view.balance = 0n;
    }
  }

  function setAddress(index: number) {
    addressIndex = index;

    view.address = Seed.generateAddress(seed, addressIndex);
    keyPair = Seed.generateKeyPair(seed, addressIndex);
  }

  function createNewAddress() {
    setAddress(addressIndex++);
  }

  function startFundsUpdater() {
    if (fundsUpdaterHandle) {
      clearInterval(fundsUpdaterHandle);
    }

    fundsUpdaterHandle = setInterval(updateFunds, 1000);
  }

  function isBroke(balance: bigint) {
    return balance < 200;
  }

  function isWealthy(balance: bigint) {
    return balance > 200;
  }

  async function placeBet() {
    view.isWorking = true;
    try {
      await fairRoulette.placeBet(keyPair, view.round.betSelection, 1234);
    } catch (ex) {
      alert(ex);
    }
    view.isWorking = false;
  }

  async function sendFaucetRequest() {
    view.isWorking = true;

    const faucetRequestResult = await client.getFaucetRequest(view.address);

    faucetRequestResult.faucetRequest.nonce = await powManager.RequestProofOfWork(
      20, // In this example a difficulty of 20 is enough, might need a retune for prod to 21 or 22
      faucetRequestResult.poWBuffer
    );

    try {
      await client.sendFaucetRequest(faucetRequestResult.faucetRequest);
    } catch (ex) {
      alert(ex);
    }

    view.isWorking = false;
  }

  function subscribeToRouletteEvents() {
    fairRoulette.on('roundStarted', () => {
      view.round.active = true;
      log('[ROUND] started');
    });

    fairRoulette.on('roundStopped', () => {
      view.round.active = false;
      log('[ROUND] ended');
    });

    fairRoulette.on('roundNumber', (roundNumber: bigint) => {
      view.round.number = roundNumber;
      log(`[ROUND] Current round number: ${roundNumber}`);
    });

    fairRoulette.on('winningNumber', (winningNumber: bigint) => {
      view.round.winningNumber = winningNumber;
      log(`[ROUND] The winning number was: ${winningNumber}`);
    });

    fairRoulette.on('betPlaced', (bet: Bet) => {
      log(`[BET] Bet placed from ${bet.better} on ${bet.betNumber} with ${bet.amount}`);
    });

    fairRoulette.on('payout', (bet: Bet) => {
      log(`[WIN] Payout for ${bet.better} with ${bet.amount}`);
    });
  }

  async function initialize() {
    log('Page loading');
    powManager.Load('/build/pow.worker.js');

    subscribeToRouletteEvents();
    createNewAddress();
    updateFunds();

    startFundsUpdater();

    view.round.active = (await fairRoulette.getRoundStatus()) == 1;
    view.round.number = await fairRoulette.getRoundNumber();
    view.round.winningNumber = await fairRoulette.getLastWinningNumber();
    log('Page loaded');
  }

  // Entrypoint
  onMount(initialize);
  // Entrypoint
</script>

<main>
  <div class="header">
    <ul>
      <li>
        <span class="header_balance_title">Balance</span>
        <span class="header_balance_text">{view.balance}i</span>
      </li>
      <li>
        <span class="header_seed_title">Seed</span>
        <span class="header_seed_text">{view.seedString}</span>
      </li>
      <li>
        <span class="header_address_title">Address</span>
        <span class="header_address_text">{view.address}</span>
      </li>
    </ul>
  </div>

  {#if view.isWorking}
    <div class="loading_dim">
      <div class="loading_wrapper">
        <div class="loading_logo" />
        <div class="loading_text">Loading and working</div>
      </div>
    </div>
  {/if}

  {#if view.showController}
    {#if isBroke(view.balance)}
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
      </div>
    {/if}

    {#if isWealthy(view.balance)}
      <div class="welcome_screen">
        <div class="place_bet" class:disabled={isBroke(view.balance)}>
          <div class="place_bet_text">2 | Place a bet</div>
          <div class="place_bet_image" />

          <div class="bet_selection">
            <input
              type="radio"
              group={view.round.betSelection}
              name="bet_number"
              value="1"
              checked
            />
            <label for="1">1</label>

            <input type="radio" group={view.round.betSelection} name="bet_number" value="2" />
            <label for="2">2</label>

            <input type="radio" group={view.round.betSelection} name="bet_number" value="3" />
            <label for="3">3</label>

            <input type="radio" group={view.round.betSelection} name="bet_number" value="4" />
            <label for="4">4</label>

            <input type="radio" group={view.round.betSelection} name="bet_number" value="5" />
            <label for="5">5</label>
          </div>

          <button
            class="submit_bet_button"
            on:click={() => placeBet()}
            disabled={isBroke(view.balance)}>Submit bet</button
          >
        </div>
      </div>
    {/if}
  {/if}

  <div class="content">
    <div class="wheel_container">
      <Roulette width="300" height="300" spin={view.round.active} />

      <div class="wheel_status">
        <div>Round {view.round.active ? 'started' : 'stopped'}</div>
        <div>Round Nr. {view.round.number}</div>
        <div>
          Winning Nr. {view.round.active ? 'Undecided' : view.round.winningNumber}
        </div>
      </div>
    </div>

    <div class="log">
      <textarea rows="4" cols="50">{view.round.eventList.join('\n')}</textarea>
    </div>
  </div>
</main>

<style>
  textarea {
    width: 100%;
    height: 100%;
  }

  .log {
    width: 75%;
    margin-left: 35px;
  }
  .wheel_container {
    border: 1px solid white;
    border-radius: 4px;
    padding: 5px;
  }

  .wheel_status {
    margin-top: 15px;
    width: 100%;
    height: 100px;
    font-size: 26px;
    color: gray;
  }

  .content {
    display: flex;
    width: 75%;
    margin: 15px auto;
  }

  main {
    width: 100%;
    height: 100%;
  }

  div.header {
    width: 100%;
    background-color: rgba(72, 87, 118, 0.2);
    height: 45px;
  }

  div.header ul {
    margin: 0;
    padding: 0;
    display: flex;
    list-style: none;
    height: 100%;
  }

  div.header li {
    height: 100%;
    margin: auto;
    display: flex;
    padding: 0;
    align-items: center;
    justify-content: center;
    color: rgba(255, 255, 255, 0.9);

    text-shadow: 0px 2px 2px rgba(13, 15, 4, 0.42);
  }

  .header_balance_title {
    font-size: 24px;
  }

  .header_balance_text {
    padding-left: 5px;
    font-size: 26px;
  }

  .header_seed_title,
  .header_address_title {
    font-size: 20px;
  }

  .header_seed_text,
  .header_address_text {
    font-size: 16px;
    padding-left: 5px;
  }

  .loading_dim {
    background: rgba(0, 0, 0, 0.5);
    width: 100%;
    height: 100%;
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 999;
  }

  .loading_wrapper {
    position: absolute;

    top: 0;
    left: 0;
    right: 0;
    bottom: 0;

    width: 260px;
    height: 260px;

    margin: auto;
  }

  .loading_logo {
    width: 160px;
    height: 160px;
    margin: auto;
  }

  /**
    * Source images for this gif are from: https://www.reddit.com/r/Iota/comments/8kp9uv/created_an_iota_loading_animation_free_for_anyone/
    * They were scaled down and the transparency was replaced with the dim background of the page.
  */
  .loading_logo {
    background-image: image-set(
      url('/iota.webp') type('image/webp'),
      url('/iota.gif') type('image/gif')
    );

    background-repeat: no-repeat;
  }

  .loading_text {
    text-shadow: 0px 2px 2px rgba(13, 15, 4, 0.42);
    font-size: 24px;
    color: white;
    text-align: center;
  }

  .welcome_screen {
    display: flex;
    color: white;
    margin: 25px auto;
    width: 25%;
  }

  .welcome_screen > .disabled {
    color: dimgrey !important;
  }

  .request_funds {
    width: 200px;
  }

  .place_bet {
    width: 200px;
  }

  .request_funds_text,
  .place_bet_text {
    text-shadow: 0px 2px 2px rgba(13, 15, 4, 0.42);
    font-size: 24px;
    text-align: center;
  }

  .request_funds_image {
    border: none;
    background-color: transparent;
    background-image: url(/money_w_k.png);
    width: 96px;
    height: 96px;
    margin: 15px auto;
  }

  .place_bet_image {
    border: none;
    background-color: transparent;
    background-image: url(/dice_w.png);
    width: 96px;
    height: 108px;
    margin: 15px auto;
  }

  .request_funds_button,
  .submit_bet_button {
    margin: 15px auto;
    display: block;
  }

  .bet_selection {
    margin: 15px auto;
    color: white;
    display: flex;
  }

  .bet_selection input {
    flex: 1;
  }
</style>
