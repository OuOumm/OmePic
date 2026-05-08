import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

const config = {
  preprocess: vitePreprocess(),
  kit: {
    adapter: adapter({
      pages: 'out',
      assets: 'out',
      fallback: 'index.html',
      precompress: false,
      strict: true,
    }),
    alias: {
      '@/*': './src/lib/*',
    },
    prerender: {
      entries: ['*'],
    },
  },
};

export default config;
