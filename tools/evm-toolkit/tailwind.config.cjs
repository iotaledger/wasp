/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      colors: {
        shimmer: {
          ['action-primary']: '#00F5DD',
          background: '#04111B',
          ['background-secondary']: '#061928',
          ['background-tertiary']: '#082235',
          ['action-hover']: '#00D6C1',
          ['action-disabled']: '#1e7169',
          ['text-primary']: '#00121F',
          ['text-secondary']: '#738795',
          ['text-error']: '#E01A4F',
          ['background-error']: '#e01a4f29',
          ['background-tertiary-hover']: '#122b3d',
        },
      },
    },
    plugins: [],
  }
}
