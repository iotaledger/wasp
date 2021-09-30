<script lang="ts">
  import { navigateTo, routeIsActive } from 'svelte-router-spa';

  export let currentRoute;
  let isLanding: boolean = false;

  $: currentRoute, (isLanding = routeIsActive('/'));
</script>

<header class="header">
  <div class="container">
    <img src="/assets/iota-roulette-logo.svg" alt="iota-logo-roulette" />
  </div>
  {#if isLanding}
    <button on:click={() => navigateTo('/demo')}>Try demo</button>
  {/if}
</header>

<style lang="scss">
  .header {
    width: 100%;
    background-color: rgba(72, 87, 118, 0.2);
    height: 50px;
    position: relative;
    @media (min-width: 1024px) {
      height: 80px;
    }
    & img {
      max-width: 200px;
      padding: 10px 0px 10px 12px;
      @media (min-width: 1024px) {
        padding: 16px 0;
        max-width: 300px;
      }
    }
    button {
      position: absolute;
      border: 0;
      border-radius: 0;
      right: 0;
      top: 0;
      height: 50px;
      background: var(--mint-green-light);
      color: var(--white);
      display: flex;
      align-items: center;
      text-decoration: none;
      text-align: center;
      letter-spacing: 0.08em;
      text-transform: uppercase;
      font-weight: bold;
      font-size: 14px;
      line-height: 120%;
      padding: 10px;
      @media (min-width: 1024px) {
        padding: 30px;
        font-size: 16px;
        height: 80px;
      }
    }
  }
</style>
