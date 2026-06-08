import type { Config } from 'tailwindcss';

const config: Config = {
  content: ['./src/**/*.{js,ts,jsx,tsx,mdx}'],
  theme: {
    extend: {
      colors: {
        gov: {
          navy: '#1e3a5f',
          'navy-dark': '#152a45',
          slate: '#475569',
          'slate-light': '#64748b',
          gold: '#c9a227',
          'gold-light': '#e8c547',
          success: '#166534',
          warning: '#b45309',
          danger: '#b91c1c',
          surface: '#f8fafc',
          border: '#cbd5e1',
        },
      },
      fontFamily: {
        sans: ['var(--font-inter)', 'system-ui', 'sans-serif'],
      },
    },
  },
  plugins: [],
};

export default config;
