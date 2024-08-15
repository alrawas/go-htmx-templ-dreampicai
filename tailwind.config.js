/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./**/*.html', './**/*.templ', './**/*.go'],
  safelist: [],
  plugins: [require('@tailwindcss/aspect-ratio'), require('daisyui')],
  daisyui: {
    themes: ['dark'],
  },
};
