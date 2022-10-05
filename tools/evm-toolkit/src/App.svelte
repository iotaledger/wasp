<script>
  import Deposit from './components/deposit/Deposit.svelte';
  import Withdraw from './components/withdraw/Withdraw.svelte';
  import { SvelteToast } from '@zerodevx/svelte-toast'
</script>

<main>  
  <div class="item">
    <h2>Deposit</h2>
    <Deposit />
  </div>
  <div class="item">
    <h2>Withdraw</h2>
    <Withdraw />
  </div>
</main>
<SvelteToast/>

<style>
  h2 {
    margin: 15px;
  }
  main {
    display: flex; 
    flex: 1 1 auto;
    flex-direction: row;
    justify-content: center;
  }

  .item {
    flex-grow: 1;
    max-width: 900px;
  }

  @media (max-width: 900px) {
    main {
      flex-direction: column;
    }
    .item {
      max-width: 450x;
    }
  }

</style>