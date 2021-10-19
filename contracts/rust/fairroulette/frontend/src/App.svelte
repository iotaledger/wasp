<script lang="ts">
  import { Router } from 'svelte-router-spa';
  import config from '../config.dev';
  import { CookieDisclaimer } from './components';
  import { googleAnalyticsInitialized } from './lib/store';
  import { getCookie, setCookie } from './lib/utils';
  import { routes } from './routes';

  const SITE_COOKIES_ENABLED_NAME = 'iota_roulette_cookies_enabled';
  const googleAnalyticsId = config?.googleAnalytics;
  const cookiesEnabled = getCookie(SITE_COOKIES_ENABLED_NAME) === 'true';
  let showCookieDisclaimer = false;

  $: showCookieDisclaimer =
    $googleAnalyticsInitialized &&
    !getCookie(SITE_COOKIES_ENABLED_NAME)?.length;

  $: if ($googleAnalyticsInitialized) {
    window[`ga-disable-${googleAnalyticsId}`] = !cookiesEnabled;
  }

  const allowCookies = (): void => {
    if (googleAnalyticsId) {
      window[`ga-disable-${googleAnalyticsId}`] = false;
      setCookie(SITE_COOKIES_ENABLED_NAME, true, 30);
      showCookieDisclaimer = false;
    }
  };
  const declineCookies = (): void => {
    if (googleAnalyticsId) {
      window[`ga-disable-${googleAnalyticsId}`] = true;
      setCookie(SITE_COOKIES_ENABLED_NAME, false, 30);
      showCookieDisclaimer = false;
    }
  };
</script>

<main>
  <Router {routes} options={{ gaPageviews: true }} />
  {#if showCookieDisclaimer}
    <CookieDisclaimer {allowCookies} {declineCookies} />
  {/if}
</main>
