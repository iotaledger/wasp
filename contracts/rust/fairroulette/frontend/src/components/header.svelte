<script lang="ts">
  import { Navigate } from 'svelte-router-spa';
</script>

<header class="header">
  <div class="container">
    <img src="iota-roulette.svg" alt="iota-logo-roulette" />
  </div>
  <a href="/demo">Try demo</a>
</header>

<style lang="scss">
  .header {
    width: 100%;
    background-color: rgba(72, 87, 118, 0.2);
    height: 80px;
    position: relative;
    & img {
      padding: 16px;
      max-width: 70%;
    }
    a {
      position: absolute;
      right: 0;
      top: 0;
      height: 80px;
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
      }
    }
  }
</style>
