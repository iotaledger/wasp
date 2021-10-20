<script lang="ts">
  import Button from './button.svelte';

  export let allow: () => void;
  export let decline: () => void;
</script>

<div class="cookie-disclaimer">
  <div class="disclaimer-description">
    <span
      >By using the IOTA Roulette site, you agree with our use of cookies.
      <a
        href="https://www.iota.org/privacy-policy"
        target="_blank"
        rel="noopener noreferrer">Read our Privacy Policy.</a
      >
    </span>
  </div>
  <div class="disclaimer-buttons">
    <div><Button label="Accept" onClick={allow} /></div>
    <div><Button label="Decline" secondary onClick={decline} /></div>
  </div>
</div>

<style lang="scss">
  .cookie-disclaimer {
    background: var(--blue-dark);
    display: flex;
    flex-direction: column;
    width: 100%;
    padding: 20px;
    color: var(--white);
    position: fixed;
    bottom: 0;
    align-items: center;
    font-size: 14px;
    border-radius: 0.5rem;
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
      margin-bottom: 15px;
      @media (min-width: 600px) {
        margin-right: 15px;
        margin-bottom: 0;
      }
      span {
        letter-spacing: 0.02em;
        line-height: 24px;
        a {
          color: var(--mint-green-light);
          font-weight: 700;
          text-decoration: none;
        }
      }
    }
    .disclaimer-buttons {
      display: flex;
      flex-direction: row;
      width: 100%;
      div {
        width: 100%;
        &:not(:last-of-type) {
          margin-right: 15px;
          @media (min-width: 600px) {
            margin-bottom: 10px;
            margin-right: 0;
          }
        }
      }
      @media (min-width: 600px) {
        flex-direction: column;
      }
    }
  }
</style>
