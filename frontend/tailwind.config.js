/** @type {import('tailwindcss').Config} */
module.exports = {
  // Follow the OS setting; every dark: class in the app now actually applies.
  darkMode: 'media',
  content: ['./src/**/*.{html,js}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Vazirmatn', 'system-ui', '-apple-system', '"Segoe UI"', 'Roboto', 'Tahoma', 'sans-serif'],
      },
    },
  },
  plugins: [],
};
