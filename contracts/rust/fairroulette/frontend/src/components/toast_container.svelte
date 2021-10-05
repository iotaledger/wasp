<script lang="typescript">
  import { Toast } from './../components';
  import {
    displayNotifications,
    Notification,
    NOTIFICATION_TIMEOUT_NEVER,
    removeDisplayNotification,
  } from './../lib/notifications';
  import { fade } from 'svelte/transition';

  let dummyNotifications = [
    {
      type: Notification.Info,
      message: 'Dummy info message',
      id: '1',
    },
    {
      type: Notification.Error,
      message: 'Dummy error message',
      id: '1',
    },
    {
      type: Notification.Win,
      message: 'Dummy win message',
      id: '1',
    },
    {
      type: Notification.Info,
      message: 'Dummy lose message',
      id: '1',
    },
  ];
</script>

<div class="toast-container">
  <ul>
    {#each $displayNotifications as notification}
      <li in:fade={{ duration: 100 }} out:fade={{ duration: 100 }}>
        <Toast
          {...notification}
          timeout={NOTIFICATION_TIMEOUT_NEVER}
          onClose={() => removeDisplayNotification(notification.id)}
        />
      </li>
    {/each}
  </ul>
</div>

<style type="text/scss">
  .toast-container {
    position: absolute;
    right: 20px;
    bottom: 20px;
    width: 400px;
    ul {
      display: flex;
      flex-direction: column;
      z-index: 3;
      gap: 8px;
    }
  }
</style>
