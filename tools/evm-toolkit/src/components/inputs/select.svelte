<script lang="ts">
  export let options: [string, string][] = [] // [key, value][]
  export let index: number = 0
  export let value: string = ''
  export let title: string = ''
  export let stretch: boolean = false

  $: value = options[index][1]
</script>

<div class="flex flex-col" class:w-full={stretch}>
  <label class="block text-12" for={title}>{title}</label>
  <select bind:value={index}>
      {#each options as option, i}
          <option value={i}>{option[1]}</option>
      {/each}
  </select>
</div>

<style lang="scss">
  select {
      @apply border;
      @apply border-zinc-200;
      @apply rounded-md;
      @apply py-2;
      @apply px-3;
      @apply flex;
      @apply flex-col;
      @apply bg-white;
      @apply h-full;
  }
  label {
      @apply mb-1;
  }
</style>
