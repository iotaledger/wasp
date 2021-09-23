<script lang="ts">
  import { Router } from 'svelte-router-spa';
  import Header from './components/header.svelte';
  import { routes } from './routes';
  import Footer from './components/footer.svelte';

  const FOUNDATION_DATA_URL = 'https://webassets.iota.org/data/foundation.json';

  async function getFoundationData() {
    let response = await fetch(FOUNDATION_DATA_URL);
    let users = await response.json();
    return users;
  }
  const promise = getFoundationData();
</script>

<main>
  <Header />
  <Router {routes} />
  {#await promise then foundationData}
    <Footer {foundationData} />
  {/await}
</main>
