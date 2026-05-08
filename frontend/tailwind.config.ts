import type { Config } from 'tailwindcss';

const config: Config = {
  darkMode: ['class'],
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      colors: {
        paper: 'hsl(var(--paper))',
        ink: 'hsl(var(--ink))',
      },
    },
  },
  plugins: [],
};

export default config;
