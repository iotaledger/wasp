<script lang="ts">
  import Button from '../components/button.svelte';
  import { Navigate } from 'svelte-router-spa';

  const HEADINGS = {
    eyebrow: 'PoC for Smart Contracts',
    title: 'IOTA Roulette',
    description:
      'IOTA Roulette is a PoC for IOTA Smart Contracts. In the demo, you can bet on a number from 1-8 and spin the roulette wheel. After starting the demo, you automatically have coins from the devnet transferred to your demo balance. Depending on if you win or lose, the underlying Smart Contract sends you your gained coins if won or keeps them if lost. ',
  };
  const CARDS = [
    {
      title: 'How it works',
      description:
        'IOTA Roulette is a PoC for IOTA Smart Contracts. In the demo, you can bet on a number from 1-8 and spin the roulette wheel. After starting the demo, you automatically have coins from the devnet transferred to your demo balance. Depending on if you win or lose, the underlying Smart Contract sends you your gained coins if won or keeps them if lost. ',
    },
    {
      title: 'What it does',
      description:
        'IOTA Roulette is a PoC for IOTA Smart Contracts. In the demo, you can bet on a number from 1-8 and spin the roulette wheel. After starting the demo, you automatically have coins from the devnet transferred to your demo balance. Depending on if you win or lose, the underlying Smart Contract sends you your gained coins if won or keeps them if lost. ',
    },
  ];
</script>

<section class="container">
  <div class="hero">
    <div class="hero-description">
      <span class="eyebrow">{HEADINGS.eyebrow}</span>
      <h1 class="title">{HEADINGS.title}</h1>
      <p class="description">{HEADINGS.description}</p>
      <div class="demo-button">
        <Navigate to="/demo">
          <Button label="Try demo" /></Navigate
        >
      </div>
    </div>
    <div class="wheel">
      <img src="wheel.svg" alt="wheel" />
    </div>
  </div>
  <div class="cards">
    {#each CARDS as card}
      <div class="card">
        <span class="title">{card.title}</span>
        <p class="description">{card.description}</p>
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
