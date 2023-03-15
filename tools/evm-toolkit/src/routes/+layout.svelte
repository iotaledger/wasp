<script lang="ts">
  import { Navbar } from '$components';
import '../app.scss'
 
</script>

<Navbar />
<main class="w-full h-full">
  <slot />
</main>