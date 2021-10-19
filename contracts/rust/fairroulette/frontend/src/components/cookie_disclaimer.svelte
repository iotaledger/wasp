<script lang="ts">
  export let allow: () => void;
  export let decline: () => void;
</script>

<div class="cookie-disclaimer">
  <div class="disclaimer-description">
    <span>By using the IOTA Roulette, you agree with our use of cookies. </span>
    <a href="https://www.iota.org/privacy-policy">Read our Privacy Policy.</a>
  </div>
  <div class="disclaimer-buttons">
    <button class="green-button" on:click={allow}>Allow cookies</button>
    <button on:click={decline}>Decline cookies</button>
  </div>
</div>

<style lang="scss">
  .cookie-disclaimer {
    background: var(--blue-dark);
    display: flex;
    flex-direction: column;
    width: 100%;
    border-radius: 0.5rem;
    padding: 20px;
    color: var(--white);
    position: fixed;
    bottom: 0;
    align-items: center;
    font-size: 14px;
    border: 1px solid rgba(255, 255, 255, 0.12);
    @media (min-width: 600px) {
      width: 462px;
      flex-direction: row;
      bottom: 20px;
      right: 20px;
    }
    .disclaimer-description {
      display: flex;
      flex-direction: column;
      margin-right: 15px;
      span {
        letter-spacing: 0.02em;
        line-height: 22px;
      }
      a {
        color: var(--mint-green-light);
        font-weight: 700;
        line-height: 22px;
        text-decoration: none;
      }
    }
    .disclaimer-buttons {
      display: flex;
      flex-direction: row;
      width: 100%;
      @media (min-width: 600px) {
        flex-direction: column;
      }
      button {
        width: 100%;
        margin: 10px 5px;
        border-radius: 0.5rem;
        font-weight: 700;
        padding: 0.75rem 1.5rem;
        line-height: 14px;
        text-align: center;
        text-transform: uppercase;
        background-color: transparent;
        border: 2px solid rgba(255, 255, 255, 0.12);
        color: var(--gray-5);
        @media (min-width: 600px) {
          margin: 10px 0;
        }
        &.green-button {
          background: var(--mint-green-light);
          color: var(--white);
          border-color: var(--mint-green-light);
        }
      }
    }
  }
</style>
