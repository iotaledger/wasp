<script lang="ts">
  import { navigateTo, routeIsActive } from 'svelte-router-spa';

  export let currentRoute;
  let isLanding: boolean = false;

  $: currentRoute, (isLanding = routeIsActive('/'));

  let isRepositoriesExpanded: boolean = false;

  const REPOSITORIES: { name: string; link: string }[] = [
    {
      name: 'Dummy repo',
      link: 'http://www.iota.org',
    },
  ];
</script>

<header class="header">
  <div class="container">
    <div class="logo" on:click={() => navigateTo('/')}>
      <img src="/assets/iota-roulette-logo.svg" alt="iota-logo-roulette" />
    </div>

    <div class="nav-items">
      <a target="_blank" href="https://wiki.iota.org/">Visit the Wiki</a>

      <div
        class="repositories"
        on:click={() => {
          isRepositoriesExpanded = !isRepositoriesExpanded;
        }}
      >
        <img src="/assets/github.svg" alt="iota-logo-roulette" />
        <span>Repositories</span>
        <img
          class:expanded={isRepositoriesExpanded}
          src="/assets/dropdown.svg"
          alt="iota-logo-roulette"
        />
      </div>
    </div>
  </div>
  {#if isLanding}
    <button class="try-demo" on:click={() => navigateTo('/demo')}
      >Try demo</button
    >
  {/if}
  {#if isRepositoriesExpanded}
    <div class="repositories-expanded">
      {#each REPOSITORIES as { name, link }}
        <a class="repo" target="_blank" href={link}>{name}</a>
      {/each}
    </div>
  {/if}
</header>

<style lang="scss">
  .header {
    width: 100%;
    background-color: rgba(72, 87, 118, 0.2);
    height: 50px;
    position: relative;
    display: flex;

    .container {
      display: flex;
      align-items: center;
      width: 100%;

      .logo {
        cursor: pointer;
        img {
          max-width: 200px;
          padding: 10px 0px 10px 12px;
          @media (min-width: 1024px) {
            padding: 16px 0;
            max-width: 300px;
          }
        }
      }
      .nav-items {
        display: flex;
        justify-content: flex-end;
        align-items: center;

        width: 100%;
        gap: 32px;

        font-family: 'Inter', sans-serif;
        font-size: 14px;
        line-height: 21px;
        letter-spacing: 0.5px;
        color: white;
        .repositories {
          display: flex;
          align-items: center;
          gap: 8px;
          cursor: pointer;
          img {
            transition: transform 0.2s ease;
            &.expanded {
              transform: rotate(180deg);
            }
          }
        }
      }
    }
    .repositories-expanded {
      position: absolute;
      border-top: 1px solid #1e2439;
      top: 100%;
      width: 100%;
      background-color: #262e44;
      min-height: 250px;
      padding: 64px;
      z-index: 2;
    }
    @media (min-width: 1024px) {
      height: 80px;
    }

    a {
      color: white;
      text-decoration: none;
      font-family: 'Inter';
    }
    .try-demo {
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
