<script lang="ts">
  export let label: string;
  export let disabled: boolean = false;
  export let onClick: () => void;
  export let loading: boolean = false;
</script>

<button {disabled} class:disabled on:click={onClick}>
  <div class="label">{label}</div>
  {#if loading}
    <div class="spinner" />
  {/if}
</button>

<style lang="scss">
  button {
    background: var(--mint-green-light);
    border: 1px solid var(--mint-green-light);
    border-radius: 6px;
    font-weight: bold;
    font-size: 14px;
    line-height: 120%;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--white);
    padding: 12px;
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    .label {
      margin: 0 6px;
    }
    @media (min-width: 1024px) {
      padding: 12px 30px;
    }
    &.disabled {
      background: transparent;
      color: var(--gray-5);
    }
    .spinner {
      position: relative;
      top: 10px;
      left: 10px;
      width: 20px;
      height: 20px;
      animation: pulse 1s ease-in-out infinite;
      border-radius: 100%;
      background-color: #c3d0e4;
    }

    @keyframes pulse {
      0% {
        transform: translate(-50%, -50%) scale(0);
        opacity: 1;
      }

      100% {
        transform: translate(-50%, -50%) scale(1);
        opacity: 0;
      }
    }

    .spinner + * {
      margin-left: 20px;
    }
  }
</style>
