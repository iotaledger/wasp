<script>
  import lottie from 'lottie-web';
  import { onDestroy } from 'svelte';

  export let animation = undefined;
  export let loop = true;
  export let autoplay = true;
  export let renderer = 'svg';
  export let destroyWhenFinished = false;

  const animations = {
    win: 'win',
    loading: 'loading',
  };

  let container;
  let lottieAnimation;

  $: selected = animations[animation];

  $: if (selected && container) {
    const options = {
      container,
      renderer,
      path: `assets/animations/${selected}.json`,
      loop,
      autoplay,
    };
    lottieAnimation = lottie.loadAnimation(options);
  }

  $: if (lottieAnimation && destroyWhenFinished) {
    lottieAnimation.onComplete = function () {
      destroyAnimation();
    };
  }

  function destroyAnimation() {
    if (lottieAnimation) {
      try {
        lottieAnimation.destroy();
      } catch (e) {
        console.error(e);
      }
    }
  }
  onDestroy(() => {
    if (lottieAnimation) {
      destroyAnimation();
    }
  });
</script>

<div bind:this={container} />
