import config from '../config.dev';
import App from './App.svelte';
import { googleAnalyticsInitialized } from './lib/store';
import { googleAnalytics } from './lib/utils';

const app = new App({
	target: document.body,
	props: {

	}
});

if (config?.googleAnalytics) {
	window[`ga-disable-${config?.googleAnalytics}`] = true; // disable GA before loading
	googleAnalytics(config?.googleAnalytics);
	googleAnalyticsInitialized.set(true);
}

export default app;
