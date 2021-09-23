<script lang="ts">
  export let foundationData;

  const { registeredAddress, visitingAddress, information } = foundationData;
  const SOCIAL_LINKS = [
    {
      title: 'Twitter',
      url: 'https://twitter.com/iota',
      icon: '/icons/twitter-icon.svg',
    },
    {
      title: 'Discord',
      url: 'https://discord.iota.org/',
      icon: '/icons/discord-icon.svg',
    },
    {
      title: 'Github',
      url: 'https://github.com/iotaledger',
      icon: '/icons/github-icon.svg',
    },
  ];
</script>

<footer class="relative bg-grey-800  ">
  <div class="container z-0">
    <div class="flex relative">
      <div
        class="pattern-bg-wrapper relative overflow-hidden order-last lg:order-first col-2 mt-96 lg:mt-0 md:col-3 -mr-6"
      >
        <img
          src="assets/footer-pattern.svg"
          alt="shimmer footer"
          class="-left-36 -bottom-130 absolute"
        />
      </div>
      <div class="pt-28 pb-20 col-3 md:col-9">
        <div class="flex lg:flex-row flex-col pb-16">
          <div
            class=" col-6 mb-16 md:mb-24 text-16 leading-150 tracking-0.04 text-grey-10 z-10"
          >
            <img src="iota-roulette.svg" alt="iota-logo-roulette" />
          </div>
          <div
            class="social-pattern-bg h-40 flex justify-center md:justify-start space-x-14 md:space-x-20 py-16 px-12 col-6 w-1/2 lg:3/4"
          >
            {#each SOCIAL_LINKS as { title, url, icon }}
              <a target="_blank" href={url}
                ><img src={icon} alt={title} class="max-w-none w-10" /></a
              >
            {/each}
          </div>
        </div>

        {#if registeredAddress || visitingAddress || information}
          <div
            class=" flex lg:flex-row flex-col justify-start text-14 text-grey-600 col-9"
          >
            {#if registeredAddress || visitingAddress}
              <div
                class="flex flex-wrap space-y-2 md:space-y-0 justify-start md:col-6 md:col-start-1"
              >
                <div class="w-full md:col-2">
                  <p
                    class="mb-2 text-16 leading-150 tracking-0.04 text-grey-500"
                  >
                    {registeredAddress?.label ?? ''}
                  </p>
                  <p
                    class="whitespace-pre text-14 leading-150 tracking-0.04 text-grey-600 mb-4"
                  >
                    {registeredAddress?.value?.join('\n') ?? ''}
                  </p>
                </div>
                <div class="w-full md:col-2">
                  <p
                    class="mb-2 text-16 leading-150 tracking-0.04 text-grey-500"
                  >
                    {visitingAddress?.label ?? ''}
                  </p>
                  <p
                    class="whitespace-pre text-14 leading-150 tracking-0.04 text-grey-600"
                  >
                    {visitingAddress?.value?.join('\n') ?? ''}
                  </p>
                </div>
              </div>
            {/if}
            {#if information}
              <div
                class="md:col-6 mt-8 lg:mt-0 text-12 leading-165 tracking-0.02 text-grey-600"
              >
                {#each information as informationLine}
                  <div>
                    <span
                      >{informationLine?.label?.replace(/&copy;/g, '©') ??
                        ''}</span
                    >
                    {#if informationLine?.urls}
                      {#each informationLine?.urls as url, i}
                        <a
                          class="text-green-200 inline-block-block"
                          href={url?.url}>{url?.label}</a
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
          </div>
        {/if}
      </div>
    </div>
  </div>
</footer>

<style lang="scss">
  footer {
    .pattern-bg-wrapper {
      height: 577px;
      img {
        min-width: 475px;
      }
      overflow: hidden;
    }
    .social-pattern-bg {
      background-image: url('/pattern/pattern-footer.svg');
      background-position-x: 40%;
    }
  }
</style>
