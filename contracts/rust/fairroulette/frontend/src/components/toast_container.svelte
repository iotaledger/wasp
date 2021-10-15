<script lang="typescript">
  import { Toast } from './../components';
  import {
    displayNotifications,
    removeDisplayNotification,
  } from './../lib/notifications';
  import { fade } from 'svelte/transition';
</script>

<div class="toast-container">
  <ul>
    {#each $displayNotifications as notification}
      <li in:fade={{ duration: 100 }} out:fade={{ duration: 100 }}>
        <Toast
          {...notification}
          onClose={() => removeDisplayNotification(notification.id)}
        />
      </li>
    {/each}
  </ul>
</div>

<style type="text/scss">
  .toast-container {
    position: fixed;
    right: 20px;
    bottom: 50px;
    width: 100%;
    @media (min-width: 600px) {
      width: 400px;
    }
    ul {
      list-style-type: none;
      display: flex;
      flex-direction: column;
      z-index: 3;
      gap: 8px;
    }
  }
</style>
