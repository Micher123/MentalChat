/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'sage': '#8FBC8F',
        'dusty-rose': '#CAC0B5',
        'light-lilac': '#C8A2C6',
        'storm': '#708090',
        'french-gray': '#BEBEBE',
        'almond': '#EEDC82',
        'oat': '#F5F5DC',
      },
    },
  },
  plugins: [],
}
