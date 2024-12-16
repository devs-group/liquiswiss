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
        'transactions': 'minmax(14rem, 1fr), 7rem, 7rem, 7rem, minmax(10rem, 1fr), 10rem, 10rem, 10rem, 12rem',
        'employees': 'minmax(14rem, 1fr), 14rem, minmax(12rem, 1fr), 12rem, 12rem, 12rem',
        'bank-accounts': '18rem, minmax(14rem, 1fr)',
      },
      maxWidth: {
        screen: '100vw'
      }
    },
  },
  plugins: [],
}
