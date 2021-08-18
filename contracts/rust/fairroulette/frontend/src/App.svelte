<script lang="ts">
  export const name = 'app';

  import type { Buffer } from './client/buffer';
  import { Seed } from './client/crypto/seed';
  import { BasicClient, Colors } from './client/basic_client';
  import { onMount } from 'svelte';
  import { Base58 } from './client/crypto/base58';
  import { PoWWorkerManager } from './web_worker/pow_worker_manager';
  import type { IKeyPair } from './client/crypto/models/IKeyPair';
  import { FairRoulette } from './fair_roulette';

  import config from '../config.dev';

  const chainId = config.chainId; // TODO: Make configurable to adapt to other Wasp instances
  const testSeed = Base58.decode('H2ddSikVwqSfwo2KAzhDf4wRnr4oqqy8SDHofwewKaJ1');

  const seed: Buffer = Seed.generate(); //testSeed;
  const seedString: string = Base58.encode(seed);

  const client: BasicClient = new BasicClient({
    GoShimmerAPIUrl: config.goshimmerApiUrl,
    WaspAPIUrl: config.waspApiUrl,
    SeedUnsafe: seed,
  });

  const fairRoulette: FairRoulette = new FairRoulette(client, chainId);

  const powManager = new PoWWorkerManager();

  let addressIndex: number = 0;
  let address: string;
  let keyPair: IKeyPair;
  let iotaBalance: bigint = BigInt(0);

  let isWorking = false;
  let fundsUpdaterHandle;

  let betSelection = 1;

  async function updateFunds() {
    iotaBalance = await client.getFunds(address, Colors.IOTA_COLOR_STRING);
  }

  function setAddress(index: number) {
    addressIndex = index;

    address = Seed.generateAddress(seed, addressIndex);
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

  function initialize() {
    powManager.Load('/build/pow.worker.js');

    createNewAddress();
    updateFunds();

    startFundsUpdater();
  }

  onMount(initialize);

  function isBroke(balance: bigint) {
    return balance < 200;
  }

  function isWealthy(balance: bigint) {
    return balance > 200;
  }

  async function placeBet() {
    isWorking = true;
    console.log('Setting bet to: ' + betSelection);
    await fairRoulette.placeBet(keyPair, betSelection, 1234);
    isWorking = false;
  }

  async function sendFaucetRequest() {
    isWorking = true;

    const faucetRequestResult = await client.getFaucetRequest(address);

    faucetRequestResult.faucetRequest.nonce = await powManager.RequestProofOfWork(
      20,
      faucetRequestResult.poWBuffer
    );

    await client.sendFaucetRequest(faucetRequestResult.faucetRequest);

    isWorking = false;
  }
</script>

<main>
  <div class="header">
    <ul>
      <li>
        <span class="header_balance_title">Balance</span>
        <span class="header_balance_text">{iotaBalance}i</span>
      </li>
      <li>
        <span class="header_seed_title">Seed</span>
        <span class="header_seed_text">{seedString}</span>
      </li>
      <li>
        <span class="header_address_title">Address</span>
        <span class="header_address_text">{address}</span>
      </li>
    </ul>
  </div>

  {#if isWorking}
    <div class="loading_dim">
      <div class="loading_wrapper">
        <div class="loading_logo" />
        <div class="loading_text">Loading and working</div>
      </div>
    </div>
  {/if}

  <div class="welcome_screen">
    <div class="request_funds" class:disabled={isWealthy(iotaBalance)}>
      <div class="request_funds_text">1 | Request funds</div>
      <div class="request_funds_image" />
      <button
        class="request_funds_button"
        disabled={isWealthy(iotaBalance)}
        on:click={() => sendFaucetRequest()}>Request funds</button
      >
    </div>

    <div class="place_bet" class:disabled={isBroke(iotaBalance)}>
      <div class="place_bet_text">2 | Place a bet</div>
      <div class="place_bet_image" />

      <div class="bet_selection">
        <input type="radio" group={betSelection} name="bet_number" value="1" checked />
        <label for="1">1</label>

        <input type="radio" group={betSelection} name="bet_number" value="2" />
        <label for="2">2</label>

        <input type="radio" group={betSelection} name="bet_number" value="3" />
        <label for="3">3</label>

        <input type="radio" group={betSelection} name="bet_number" value="4" />
        <label for="4">4</label>

        <input type="radio" group={betSelection} name="bet_number" value="5" />
        <label for="5">5</label>
      </div>

      <button class="submit_bet_button" on:click={() => placeBet()} disabled={isBroke(iotaBalance)}
        >Submit bet</button
      >
    </div>
  </div>
</main>

<style>
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
