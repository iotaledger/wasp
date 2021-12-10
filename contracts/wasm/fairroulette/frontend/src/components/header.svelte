<script lang="ts">
  import { navigateTo, routeIsActive } from 'svelte-router-spa';
  import { fly } from 'svelte/transition';

  export let currentRoute: string;
  let isLanding: boolean = false;

  $: currentRoute, (isLanding = routeIsActive('/'));

  let isRepositoriesExpanded: boolean = false;
  let isMenuExpanded: boolean = false;

  const REPOSITORIES: { label: string; link: string }[] = [
    {
      label: 'Fair Roulette',
      link: 'https://github.com/iotaledger/wasp/tree/master/contracts/wasm/fairroulette',
    },
    {
      label: 'Wasp',
      link: 'https://github.com/iotaledger/wasp',
    },
  ];

  const NAV_LINKS: {
    label: string;
    href: string;
    target: '_blank' | 'self';
    landingPage?: boolean;
  }[] = [
    {
      label: 'About Demo',
      href: '/',
      target: 'self',
      landingPage: true,
    },
    {
      label: 'About ISCP',
      href: 'https://blog.iota.org/an-introduction-to-iota-smart-contracts-16ea6f247936/',
      target: '_blank',
    },
    {
      label: 'Wasp Wiki',
      href: 'https://wiki.iota.org/wasp/welcome/',
      target: '_blank',
    },
    {
      label: 'Demo Wiki',
      href: 'https://wiki.iota.org/wasp/guide/example_projects/fair_roulette',
      target: '_blank',
    },
    {
      label: 'White Paper',
      href: 'https://files.iota.org/papers/ISC_WP_Nov_10_2021.pdf',
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
      {#each NAV_LINKS as { label, href, target, landingPage }}
        <a {target} {href} class:active={landingPage && isLanding}>{label}</a>
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
            alt="dropdown"
            class="arrow"
          />
        </div>
        {#if isRepositoriesExpanded}
          <div class="repositories-expanded">
            {#each REPOSITORIES as { label, link }}
              <a class="repo" target="_blank" href={link}>{label}</a>
            {/each}
          </div>
        {/if}
      </div>
      {#if isLanding}
        <div class="empty" />
        <button class="try-demo" on:click={() => navigateTo('/demo')}
          >Try demo</button
        >
      {/if}

      <!-- Mobile menu -->
      <div
        class="open-menu"
        on:click={() => {
          isMenuExpanded = true;
        }}
      >
        <img src="/assets/burger.svg" alt="menu" />
      </div>
    </div>
    {#if isMenuExpanded}
      <aside class="aside-expanded" transition:fly={{ x: 800, duration: 500 }}>
        <div
          class="close-expanded"
          on:click={() => {
            isMenuExpanded = false;
          }}
        >
          <img src="/assets/close.svg" alt="close" />
        </div>
        <div class="aside-links">
          {#each NAV_LINKS as { label, href, target, landingPage }}
            <a {target} {href} class:active={landingPage && isLanding}
              >{label}</a
            >
          {/each}
          <div
            class="dropdown flex-shrink-0"
            on:click={() => {
              isRepositoriesExpanded = !isRepositoriesExpanded;
            }}
          >
            <span>Repositories</span>
            <img
              class:expanded={isRepositoriesExpanded}
              src="/assets/dropdown.svg"
              alt="dropdown"
              class="arrow"
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
      </aside>
    {/if}
  </div>
</header>

<style lang="scss">
  .header {
    width: 100%;
    background-color: #091326;
    height: 50px;
    position: relative;
    display: flex;
    border-bottom: 1px solid rgba(255, 255, 255, 0.12);
    .container {
      display: flex;
      align-items: center;
      width: 100%;
      justify-content: space-between;
      .logo {
        cursor: pointer;
        img {
          max-width: 200px;
          padding: 10px 0px 0px 12px;
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
        line-height: 150%;
        letter-spacing: 0.75px;
        color: var(--gray-3);
        width: 100%;
        font-size: 14px;
        line-height: 21px;
        letter-spacing: 0.5px;
        color: var(--white);
        & > a {
          display: none;
          @media (min-width: 1024px) {
            display: flex;
            padding: 16px 0;
            margin: 0 16px;
          }
        }
        .repositories {
          align-items: center;
          cursor: pointer;
          display: none;
          position: relative;
          flex-shrink: 0;
          margin-left: 16px;
          @media (min-width: 1024px) {
            display: flex;
          }
          img {
            margin-right: 8px;
          }
        }
        .empty {
          width: 0;
          @media (min-width: 1010px) {
            width: 140px;
          }
          @media (min-width: 1440px) {
            width: 0;
          }
        }
      }
      .open-menu {
        display: flex;
        margin-right: 10px;
        cursor: pointer;
        @media (min-width: 1024px) {
          display: none;
        }
      }
      .dropdown {
        position: relative;
        font-size: 14px;
        line-height: 150%;
        letter-spacing: 0.75px;
        color: var(--gray-3);
        font-family: 'Inter';
        display: flex;
        justify-content: space-between;
        @media (min-width: 1024px) {
          display: block;
        }
      }
      .arrow {
        transition: transform 0.4s ease;
        &.expanded {
          transform: rotate(-180deg);
        }
      }
      .burger-menu {
        z-index: 2;
      }
      .aside-expanded {
        position: fixed;
        left: 0;
        top: 0;
        width: 100%;
        height: 100%;
        background-color: #091326;
        color: white;
        z-index: 2;
        padding: 16px 20px;
        .aside-links {
          a {
            margin-top: 20px;
            display: block;
          }
        }
        .repositories-expanded {
          padding: 16px;
          a {
            margin: 0;
          }
        }
        .close-expanded {
          cursor: pointer;
          position: absolute;
          right: 10px;
          top: 10px;
        }
        .dropdown {
          margin-top: 20px;
          cursor: pointer;
          a {
            padding: 15px 0;
          }
        }
        @media (min-width: 1024px) {
          display: none;
        }
      }
    }
    .repositories-expanded {
      position: absolute;
      padding: 24px 4px;
      top: 100%;
      width: 100%;
      background-color: #091326;
      min-height: 50px;
      z-index: 2;
      display: flex;
      flex-direction: column;

      a {
        margin: 8px 0;
      }

      @media (min-width: 1024px) {
        padding: 0 20px 20px 20px;
        border: 1px solid rgba(255, 255, 255, 0.12);
        border-top: none;
        border-radius: 0 0 12px 12px;
        top: 51px;
        width: 100%;
      }
    }
    @media (min-width: 1024px) {
      height: 80px;
    }
    a {
      font-size: 14px;
      line-height: 150%;
      letter-spacing: 0.75px;
      color: var(--gray-3);
      text-decoration: none;
      font-family: 'Inter';
      &.active {
        color: var(--mint-green-light);
      }
    }
    .try-demo {
      border: 0;
      border-radius: 0;
      right: 50px;
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
      padding: 16px;
      position: absolute;
      @media (min-width: 1024px) {
        padding: 30px;
        font-size: 14px;
        height: 80px;
        right: 0;
      }
    }
  }
</style>
