<script>
  import { Route } from 'svelte-router-spa';
  import { Header, Footer } from './../components';
  export let currentRoute;

  const FOUNDATION_DATA_URL = 'https://webassets.iota.org/data/foundation.json';

  async function getFoundationData() {
    let response = await fetch(FOUNDATION_DATA_URL);
    let users = await response.json();
    return users;
  }
  const promise = getFoundationData();
</script>

<div>
  <Header {currentRoute} />

  <Route {currentRoute} />

  {#await promise then foundationData}
    <Footer {foundationData} />
  {/await}
</div>
