<script lang="ts">
  import { notificationStore, removeNotification } from '$lib/notification'
  import { Notification } from '$components'
  import { fade } from 'svelte/transition'

  const TRANSITION_DURATION = 100
</script>

{#if $notificationStore.length}
  <div class="fixed right-0 bottom-4 sm:right-5 sm:bottom-6 w-full sm:max-w-sm z-50">
      <ul class="list-none flex flex-col space-y-2">
          {#each $notificationStore as { message, type, id } (id)}
              <li transition:fade={{ duration: TRANSITION_DURATION }}>
                  <Notification {message} {type} onClose={() => removeNotification(id)} />
              </li>
          {/each}
      </ul>
  </div>
{/if}
