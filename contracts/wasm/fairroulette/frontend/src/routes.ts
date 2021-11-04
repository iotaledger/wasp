import { Demo, Landing, Layout } from './pages';

const routes = [
  {
    name: '/',
    component: Landing,
    layout: Layout,
  },
  {
    name: 'demo',
    component: Demo,
    layout: Layout,
  },
];

export { routes };
