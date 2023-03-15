const pxToRem = (px, base = 16) => `${px / base}rem`

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      colors: {
        shimmer:{
          ['action-primary']: '#00F5DD',
          background: '#04111B',
          ['background-secondary']: '#061928',
          ['background-tertiary']: '#082235',
          ['action-hover']: '#00D6C1',
          ['action-disabled']: '#1e7169',
          ['text-primary']: '#00121F',
          ['text-secondary']: '#738795',
        },
        fontSize: {
          12: pxToRem(12),
          14: pxToRem(14),
          16: pxToRem(16),
          18: pxToRem(18),
          20: pxToRem(20),
          24: pxToRem(24),
          28: pxToRem(28),
          32: pxToRem(32),
          38: pxToRem(38),
          40: pxToRem(40),
          48: pxToRem(48),
          54: pxToRem(54),
          64: pxToRem(64),
          72: pxToRem(72),
          80: pxToRem(80),
          96: pxToRem(96),
          120: pxToRem(120),
        },
      },
    },
    plugins: [],
  }
}
