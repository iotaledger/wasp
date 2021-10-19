import config from '../config.dev';
import App from './App.svelte';
import { googleAnalytics } from './lib/utils';

const app = new App({
	target: document.body,
	props: {

	}
});

if (config?.googleAnalytics) googleAnalytics(config?.googleAnalytics);

export default app;
