<script>
  import { Route } from 'svelte-router-spa';
  import { Header, Footer } from './../components';
  
  export let currentRoute;
  export let params;

  const FOUNDATION_DATA_URL = 'https://webassets.iota.org/data/foundation.json';

  async function getFoundationData() {
    const response = await fetch(FOUNDATION_DATA_URL);
    const users = await response.json();
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
