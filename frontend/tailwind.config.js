/** @type {import('tailwindcss').Config} */
export default {
  content: [],
  darkMode: ["selector", ".lq-dark"],
  theme: {
    extend: {
      colors: {
        'primary': '#10b981',
        'help': '#a855f7'
      },
      gridTemplateColumns: {
        'transactions': 'minmax(16rem, 1fr), 8rem, 8rem, 12rem, minmax(12rem, 1fr), 12rem, 12rem, 16rem',
        'employees': 'minmax(16rem, 1fr), minmax(12rem, 1fr), minmax(12rem, 1fr), minmax(12rem, 1fr), 12rem, 12rem',
      }
    },
  },
  plugins: [],
}
