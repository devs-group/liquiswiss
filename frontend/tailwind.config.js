/** @type {import('tailwindcss').Config} */

export default {
  content: [],
  darkMode: 'media',
  theme: {
    extend: {
      colors: {
        'primary': '#10b981',
        'help': '#a855f7',
      },
      gridTemplateColumns: {
        'transactions': 'minmax(14rem, 1fr), 8rem, 8rem, 12rem, minmax(12rem, 1fr), 10rem, 10rem, 14rem',
        'employees': 'minmax(14rem, 1fr), 14rem, minmax(12rem, 1fr), 12rem, 12rem, 12rem',
      },
      maxWidth: {
        screen: '100vw'
      }
    },
  },
  plugins: [],
}
