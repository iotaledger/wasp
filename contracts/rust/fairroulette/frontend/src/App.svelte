<script lang="ts">
  export const name = 'app';

  import { Router } from 'svelte-router-spa';
  import Header from './components/header.svelte';
  import { routes } from './routes';
</script>

<main>
  <Header />
  <Router {routes} />
</main>
