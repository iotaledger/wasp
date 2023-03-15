const pxToRem = (px, base = 16) => `${px / base}rem`

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      colors: {
        green: {
          500: '#007A72',
          600: '#14CABF',
          300: '#00AD9C',
          200: '#00E0CA',
          175: '#00E0C7',
          150: '#00F5DD',
          100: '#14FFE8',
          50: '#7AFFF2',
        },
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
