<script>
  import lottie from 'lottie-web';
  import { onDestroy } from 'svelte';

  export let animation = undefined;
  export let loop = true;
  export let autoplay = true;
  export let segments = undefined;
  export let renderer = 'svg';
  export let timeout;

  const animations = {
    win: {
      path: 'win.json',
    },
    loading: {
      path: 'loading.json',
    },
  };

  let container;
  let lottieAnimation;

  $: selected = animations[animation].path;

  $: if (selected && container) {
    let options = {
      container,
      renderer,
      path: `assets/animations/${selected}`,
      loop,
      autoplay,
    };
    if (timeout) {
      setInterval(() => {
        destroyAnimation();
      }, 3000);
    }
    lottieAnimation = lottie.loadAnimation(options);
  }
  $: if (lottieAnimation && segments) {
    lottieAnimation.removeEventListener('DOMLoaded', handleSegments);
    lottieAnimation.addEventListener('DOMLoaded', handleSegments);
  }

  function handleSegments() {
    if (segments) {
      lottieAnimation.playSegments(segments, true);
    }
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
      lottieAnimation.removeEventListener('DOMLoaded', handleSegments);
      destroyAnimation();
    }
  });
</script>

<div bind:this={container} />
