<script lang="ts">
  import { navigateTo, routeIsActive } from 'svelte-router-spa';
  import { fly, slide } from 'svelte/transition';

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
      <aside class="aside-expanded" transition:slide={{ duration: 700 }}>
        <div
          class="close-expanded"
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
  {#if isLanding}
    <button class="try-demo" on:click={() => navigateTo('/demo')}
      >Try demo</button
    >
  {/if}
</header>

<style lang="scss">
  .header {
    width: 100%;
    background-color: #091326;
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
        font-size: 16px;
        line-height: 150%;
        letter-spacing: 0.75px;
        color: var(--gray-3);
        width: 100%;
        gap: 50px;
        font-size: 14px;
        line-height: 21px;
        letter-spacing: 0.5px;
        color: var(--white);
        a {
          display: none;
          @media (min-width: 1024px) {
            display: flex;
          }
        }
        .repositories {
          align-items: center;
          gap: 8px;
          cursor: pointer;
          display: none;
          @media (min-width: 1024px) {
            display: flex;
          }
        }
      }
      .open-menu {
        display: flex;
        margin-right: 15px;
        @media (min-width: 1024px) {
          display: none;
        }
      }
      .dropdown {
        position: relative;
        font-size: 16px;
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
        transform: rotate(-90deg);
        @media (min-width: 1024px) {
          transform: rotate(180deg);
        }
        &.expanded {
          transform: rotate(0);
          @media (min-width: 1024px) {
            transform: rotate(359deg);
          }
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
        background-color: #262e44;
        color: white;
        z-index: 2;
        padding: 24px;
        .repositories-expanded {
          padding: 16px;
        }
        .close-expanded {
          display: flex;
          justify-content: flex-end;
          margin-bottom: 30px;
        }
        .dropdown {
          margin-top: 20px;
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
      background-color: #262e44;
      min-height: 50px;
      z-index: 2;
      @media (min-width: 1024px) {
        left: -50px;
        padding: 0 30px 30px 30px;
        margin: 28px;
        width: 170%;
      }
    }
    @media (min-width: 1024px) {
      height: 80px;
    }
    a {
      font-size: 16px;
      line-height: 150%;
      letter-spacing: 0.75px;
      color: var(--gray-3);
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
