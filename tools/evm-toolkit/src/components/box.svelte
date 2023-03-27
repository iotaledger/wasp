<box-container
  class="w-full h-full flex flex-col bg-shimmer-background rounded-xl p-6"
>
  <slot />
</box-container>

<style lang="scss">
  box-container {
    min-height: calc(100vh - var(--navbar-height) - 50px);
    @screen md {
      min-height: 700px;
    }
  }
</style>
