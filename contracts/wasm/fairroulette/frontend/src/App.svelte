<script lang="ts">
  import { onMount } from 'svelte';
  import { Router } from 'svelte-router-spa';
  import config from '../config.dev';
  import { CookieDisclaimer } from './components';
  import { getCookie, loadGoogleAnalytics, setCookie } from './lib/utils';
  import { routes } from './routes';

  const SITE_COOKIES_ENABLED_NAME = 'iota_roulette_cookies_enabled';
  const googleAnalyticsId = config?.googleAnalyticsId;
  const cookiesEnabled = getCookie(SITE_COOKIES_ENABLED_NAME) === 'true';
  let showCookieDisclaimer = false;

  onMount(() => {
    if (googleAnalyticsId) {
      showCookieDisclaimer = !getCookie(SITE_COOKIES_ENABLED_NAME)?.length;
      if (cookiesEnabled) {
        loadGoogleAnalytics(googleAnalyticsId);
      }
    }
  });
  const allowCookies = (): void => {
    if (googleAnalyticsId) {
      loadGoogleAnalytics(googleAnalyticsId);
      setCookie(SITE_COOKIES_ENABLED_NAME, 'true', 30);
      showCookieDisclaimer = false;
    }
  };
  const declineCookies = (): void => {
    if (googleAnalyticsId) {
      setCookie(SITE_COOKIES_ENABLED_NAME, 'false', 30);
      showCookieDisclaimer = false;
    }
  };
</script>

<main>
  <Router {routes} options={{ gaPageviews: true }} />
  {#if showCookieDisclaimer}
    <CookieDisclaimer allow={allowCookies} decline={declineCookies} />
  {/if}
</main>
