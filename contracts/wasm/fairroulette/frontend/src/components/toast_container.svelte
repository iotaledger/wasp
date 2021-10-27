<script lang="typescript">
  import Toast from './toast.svelte';
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
    bottom: 50px;
    width: 100%;
    padding: 0 20px;
    @media (min-width: 600px) {
      max-width: 400px;
      padding: 0;
      right: 20px;
    }
    ul {
      list-style-type: none;
      display: flex;
      flex-direction: column;
      z-index: 3;
      padding-inline-start: 0;
      margin-block-end: 0;
      margin-block-start: 0;
      li:not(first-of-type) {
        margin-top: 8px;
      }
    }
  }
</style>
