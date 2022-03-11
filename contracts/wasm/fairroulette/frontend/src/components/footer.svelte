<script lang="ts">
  import type { IFoundationData } from '../lib/models/IWebassets';

  export let foundationData: IFoundationData;

  const { registeredAddress, visitingAddress, information } = foundationData;
  const SOCIAL_LINKS = [
    {
      name: 'Youtube',
      icon: 'youtube.svg',
      url: 'https://www.youtube.com/c/iotafoundation',
      color: '#131F37',
    },
    {
      name: 'GitHub',
      icon: 'github.svg',
      url: 'https://github.com/iotaledger/',
      color: '#2C3850',
    },
    {
      name: 'Discord',
      icon: 'discord.svg',
      url: 'https://discord.iota.org/',
      color: '#4B576F',
    },
    {
      name: 'Twitter',
      icon: 'twitter.svg',
      url: 'https://twitter.com/iota',
      color: '#6A768E',
    },
    {
      name: 'Reddit',
      icon: 'reddit.svg',
      url: 'https://www.reddit.com/r/Iota/',
      color: '#7D89A1',
    },
    {
      name: 'LinkedIn',
      icon: 'linkedin.svg',
      url: 'https://www.linkedin.com/company/iotafoundation/',
      color: '#8995AD',
    },
    {
      name: 'Instagram',
      icon: 'instagram.svg',
      url: 'https://www.instagram.com/iotafoundation/',
      color: '#99A5BD',
    },
  ];
</script>

<footer>
  <div class="container">
    <div class="logo">
      <img src="/assets/iota-logo.svg" alt="IOTA roulette" />
    </div>
    {#if registeredAddress || visitingAddress || information}
      <div class="address">
        {#if registeredAddress || visitingAddress}
          <div class="registered-address">
            <div>
              <p>{registeredAddress?.value?.join('\n') ?? ''}</p>
            </div>
            <div>
              <p>{visitingAddress?.label ?? ''}</p>
              <p>{visitingAddress?.value?.join('\n') ?? ''}</p>
            </div>
          </div>
        {/if}
      </div>
      {#if information}
        <div class="information">
          {#each information as informationLine}
            <div>
              <span
                >{informationLine?.label?.replace(/&copy;/g, '©') ?? ''}</span
              >
              {#if informationLine?.urls}
                {#each informationLine?.urls as url, i}
                  <a class="text-green inline-block-block" href={url?.url}
                    >{url?.label}</a
                  >
                  {#if informationLine?.urls?.length > 1 && i < informationLine?.urls?.length - 1}
                    <span>{`, `}</span>
                  {/if}
                {/each}
              {/if}
              {#if informationLine?.value}
                <span>{informationLine?.value}</span>
              {/if}
            </div>
          {/each}
        </div>
      {/if}
    {/if}
  </div>
  <div class="icons">
    {#each SOCIAL_LINKS as link}
      <a
        href={link.url}
        target="_blank"
        color={link.color}
        class="icon"
        style="background-color:{link.color}"
      >
        <img src="/assets/{link.icon}" alt={link.name} />
        <span>{link.name}</span>
      </a>
    {/each}
  </div>
</footer>

<style lang="scss">
  footer {
    background: #141e31;
    box-shadow: 0px 2px 6px rgba(6, 16, 35, 0.12);
    .container {
      flex-direction: column;
      color: #8493ad;
      padding: 24px 24px 50px 24px;
      @media (min-width: 1024px) {
        display: flex;
        flex-direction: row;
        padding: 24px 0 64px 0;
        justify-content: space-between;
      }
      .logo {
        display: flex;
        justify-content: center;
      }
      .address {
        display: flex;
        flex-direction: column;
        justify-content: flex-start;
        font-size: 14px;
        color: var(--gray-6);
        @media (min-width: 1024px) {
          flex-direction: row;
        }
        .registered-address {
          display: flex;
          flex-wrap: wrap;
          white-space: pre;
          margin-right: 28px;
          font-weight: 500;
          font-size: 12px;
          line-height: 150%;
          letter-spacing: 0.5px;
        }
      }
      .information {
        font-size: 12px;
        line-height: 165%;
        letter-spacing: 0.02rem;
        color: var(--grey-6);
        font-weight: 500;
        font-size: 12px;
        line-height: 150%;
        letter-spacing: 0.5px;
        .text-green {
          color: var(--mint-green);
          display: inline-block;
          text-decoration: none;
        }
      }
    }
    .icons {
      width: 100%;
      display: flex;
      padding: 0px;
      margin: 0px;
      align-items: center;
      .icon {
        background: rgb(19, 31, 54);
        width: 16.6667%;
        text-decoration: none;
        font-family: 'Metropolis Bold';
        font-size: 16px;
        line-height: 24px;
        padding: 8px;
        text-align: center;
        letter-spacing: 0.02em;
        text-transform: capitalize;
        color: rgb(246, 248, 252);
        display: flex;
        -webkit-box-pack: center;
        justify-content: center;
        @media (min-width: 1024px) {
          padding: 1rem 0.5rem;
        }
        img {
          height: 24px;
        }
        span {
          display: none;
          @media (min-width: 1024px) {
            display: flex;
            margin-left: 12px;
          }
        }
      }
    }
  }
</style>
