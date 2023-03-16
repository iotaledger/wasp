<script lang="ts">
  import type { NotificationType } from '$lib/notification'

  export let type: NotificationType
  export let message: string
  export let onClose: () => void

</script>

<notification-toast class="rounded-xl flex flex-row justify-center max-md:mx-2 items-center bg-zinc-200 p-3 {type}">
  <notification-message class="ml-3">{message}</notification-message>
  <button on:click={onClose} class="ml-auto font-bold text-18 h-4 leading-none px-1"> &#x2715; </button>
</notification-toast>

<style lang="scss">
  notification-toast {
      &.success {
          @apply bg-shimmer-background-success;
          notification-message, button {
              @apply text-shimmer-action-primary;
          }
      }
      &.error {
          @apply bg-shimmer-background-error;
          notification-message,button {
              @apply text-shimmer-text-error;
          }
      }
      &.warning {
          @apply bg-shimmer-background-warning;
        notification-message,button {
              @apply text-shimmer-text-secondary;
          }
      }
  }
</style>
