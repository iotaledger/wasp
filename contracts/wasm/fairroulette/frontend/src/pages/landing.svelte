<script lang="ts">
  import Button from '../components/button.svelte';
  import { navigateTo } from 'svelte-router-spa';

  export let currentRoute;
  export let params;

  const HEADINGS = {
    eyebrow: 'Smart Contracts Demo',
    title: 'IOTA Roulette',
    description:
      'The new IOTA Smart Contracts framework is here and can be used to power decentralized applications  – just like this Fair Roulette demo. The Proof of Concept demo gives a real-time impression of the performance of IOTA Smart Contracts by showcasing how the execution of a contract works in combination with a user-friendly front end. As the roulette shows, smart contracts can be about exchanging value, generating outcomes and trusting a decentralized software with your funds.',
  };
  const CARDS = [
    {
      title: 'How it works',
      description:
        "Because funds are necessary to make a smart contract transaction, our demo takes the form of a roulette. To start the demo, hit “request funds”. Coins from the IOTA Developer Network (DevNet) will be transferred to your demo balance. To start a round, place a bet on a number from one to eight, and let the roulette wheel spin. If your bet is successful, the underlying smart contract sends you your gained coins; if your bet isn’t successful, the smart contract keeps the coins. You can learn more about IOTA Smart Contracts in our documentation, accessible through the <a target='_blank' href='https://wiki.iota.org/wasp/welcome/'> IOTA Wiki</a>, and review the smart contract executing this roulette demo.",
    },
    {
      title: 'What it does',
      description:
        'In the demo, you request tokens from the IOTA GoShimmer nodes, which send the tokens to your wallet. By placing a bet, you wager tokens from your wallet by sending them to the smart contract; these staked tokens are held by the smart contract in a pool along with tokens gambled by other players. The smart contract generates a random number and automatically rewards the player(s) who bet on the winning number. If more than one player bets on the winning number, the earned tokens are divided based on the amount of tokens bet in total. And it’s all immutable and automated – no one can change the outcome and there is no one to trust.',
    },
  ];
</script>

<svelte:head>
  <title>IOTA Roulette</title>
</svelte:head>

<section class="container">
  <div class="hero">
    <div class="hero-description">
      <span class="eyebrow">{HEADINGS.eyebrow}</span>
      <h1 class="title">{HEADINGS.title}</h1>
      <p class="description">{HEADINGS.description}</p>
      <div class="demo-button">
        <Button label="Try demo" onClick={() => navigateTo('demo')} />
      </div>
    </div>
    <div class="wheel">
      <img src="/assets/wheel.svg" alt="wheel" />
    </div>
  </div>
  <div class="cards">
    {#each CARDS as card}
      <div class="card">
        <span class="title">{card.title}</span>
        <p class="description">{@html card.description}</p>
      </div>
    {/each}
  </div>
</section>

<style lang="scss">
  .hero {
    display: flex;
    flex-direction: column-reverse;
    @media (min-width: 1024px) {
      flex-direction: row;
      justify-content: space-between;
      margin-top: 112px;
    }
    .hero-description {
      padding: 0 24px;
      margin-top: -40px;
      @media (min-width: 1000px) {
        margin-top: 0;
      }
      @media (min-width: 1024px) {
        width: 50%;
        padding: 0;
      }
      .eyebrow {
        text-transform: uppercase;
        font-size: 16px;
        letter-spacing: 0.15em;
        font-family: 'Metropolis bold';
        @media (min-width: 1024px) {
          color: var(--mint-green);
          font-size: 24px;
        }
      }
      .title {
        font-family: 'Metropolis bold';
        margin-top: 16px;
      }
      .description {
        font-size: 16px;
        line-height: 150%;
        letter-spacing: 0.75px;
        color: var(--gray-3);
        @media (min-width: 1024px) {
          margin-top: 24px;
        }
      }
      .demo-button {
        margin-top: 30px;
        width: 50%;
      }
    }
    .wheel {
      margin: 0 auto;
      overflow: hidden;
      margin-top: 32px;
      @media (min-width: 1024px) {
        margin-top: 0;
      }
      img {
        max-height: 600px;
        width: 220%;
        margin-left: -210px;
        @media (min-width: 420px) {
          width: 100%;
          max-height: 400px;
          margin-left: 0;
        }

        @media (min-width: 1024px) {
          max-height: 600px;
        }
      }
    }
  }
  .cards {
    display: flex;
    flex-direction: column;
    margin-top: 64px;
    padding: 0 24px 100px 24px;
    @media (min-width: 1024px) {
      display: flex;
      flex-direction: row;
      justify-content: space-between;
      padding: 0 0 116px 0;
    }
    .card {
      background: #141e31;
      border: 1px solid rgba(255, 255, 255, 0.12);
      border-radius: 12px;
      padding: 40px 24px;
      margin-bottom: 16px;
      @media (min-width: 1024px) {
        width: 48%;
        flex-grow: 0;
        flex-shrink: 1;
        padding: 64px 40px;
        margin-bottom: 0;
      }
      .title {
        font-family: 'Metropolis bold';
        font-size: 28px;
        line-height: 115%;
        letter-spacing: 0.75px;
        color: var(--white);
        @media (min-width: 1024px) {
          font-size: 32px;
          line-height: 120%;
          letter-spacing: 0.02em;
        }
      }
      .description {
        font-size: 16px;
        line-height: 150%;
        letter-spacing: 0.75px;
        color: var(--gray-3);
        margin-bottom: 0;
      }
    }
  }
</style>
