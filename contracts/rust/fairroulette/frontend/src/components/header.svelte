<script lang="ts">
  import { navigateTo, routeIsActive } from 'svelte-router-spa';

  export let currentRoute;
  let isLanding: boolean = false;

  $: currentRoute, (isLanding = routeIsActive('/'));

  let isRepositoriesExpanded: boolean = false;
  let isMenuExpanded: boolean = false;

  const REPOSITORIES: { label: string; link: string }[] = [
    {
      label: 'Dummy repo',
      link: 'http://www.iota.org',
    },
  ];

  const NAV_LINKS: {
    label: string;
    href: string;
    target: '_blank' | 'self';
  }[] = [
    {
      label: 'Visit the Wiki',
      href: 'https://wiki.iota.org/',
      target: '_blank',
    },
  ];
</script>

<header class="header">
  <div class="container">
    <div class="logo" on:click={() => navigateTo('/')}>
      <img src="/assets/iota-roulette-logo.svg" alt="iota-logo-roulette" />
    </div>

    <!-- Desktop menu -->
    <div class="nav-items">
      {#each NAV_LINKS as { label, href, target }}
        <a {target} {href}>{label}</a>
      {/each}

      <div
        class="repositories"
        on:click={() => {
          isRepositoriesExpanded = !isRepositoriesExpanded;
        }}
      >
        <img src="/assets/github.svg" alt="github" />
        <div class="dropdown">
          <span>Repositories</span>
          <img
            class:expanded={isRepositoriesExpanded}
            src="/assets/dropdown.svg"
            alt="iota-logo-roulette"
          />
          {#if isRepositoriesExpanded}
            <div class="repositories-expanded">
              {#each REPOSITORIES as { label, link }}
                <a class="repo" target="_blank" href={link}>{label}</a>
              {/each}
            </div>
          {/if}
        </div>
      </div>
    </div>

    <!-- Mobile menu -->
    <div
      class="open-menu"
      on:click={() => {
        isMenuExpanded = true;
      }}
    >
      <img src="/assets/burger.svg" alt="menu" />
    </div>

    {#if isMenuExpanded}
      <div class="menu-expanded">
        <div
          on:click={() => {
            isMenuExpanded = false;
          }}
        >
          <img src="/assets/close.svg" alt="close" />
        </div>
        <div>
          {#each NAV_LINKS as { label, href, target }}
            <a {target} {href}>{label}</a>
          {/each}
          <div
            class="dropdown"
            on:click={() => {
              isRepositoriesExpanded = !isRepositoriesExpanded;
            }}
          >
            <span>Repositories</span>
            <img
              class:expanded={isRepositoriesExpanded}
              src="/assets/dropdown.svg"
              alt="iota-logo-roulette"
            />
            {#if isRepositoriesExpanded}
              <div class="repositories-expanded">
                {#each REPOSITORIES as { label, link }}
                  <a class="repo" target="_blank" href={link}>{label}</a>
                {/each}
              </div>
            {/if}
          </div>
        </div>
      </div>
    {/if}
  </div>
  {#if isLanding}
    <button class="try-demo" on:click={() => navigateTo('/demo')}
      >Try demo</button
    >
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
        }
      }
      .dropdown {
        position: relative;
      }
      img {
        transition: transform 0.2s ease;
        &.expanded {
          transform: rotate(180deg);
        }
      }
      .burger-menu {
        z-index: 2;
      }
      .menu-expanded {
        position: fixed;
        left: 0;
        top: 0;
        width: 100%;
        height: 100vh;
        background-color: #262e44;
        color: white;
        z-index: 2;
        .repositories-expanded {
          position: static;
          padding: 8px 16px;
        }
      }
    }
    .repositories-expanded {
      position: absolute;
      padding: 24px 4px;
      top: 100%;
      width: 100%;
      background-color: #262e44;
      min-height: 50px;
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
