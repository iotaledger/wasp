<script lang="ts">
  export let allowCookies: () => void;
  export let declineCookies: () => void;
</script>

<div class="cookie-disclaimer">
  <button on:click={declineCookies}>Decline cookies</button>
  <button on:click={allowCookies}>Allow cookies</button>
</div>

<style lang="scss">
  .cookie-disclaimer {
    padding: 20px;
    background: turquoise;
    position: fixed;
    bottom: 20px;
    right: 20px;
  }
</style>
