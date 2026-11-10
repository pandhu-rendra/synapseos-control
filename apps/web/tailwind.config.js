/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        bg: '#0A0F14',
        surface: '#111820',
        text: '#E6EDF3',
        muted: '#8B98A5',
        accent: '#22D3EE',
        'accent-2': '#14B8A6',
        danger: '#F87171',
        warning: '#FBBF24',
        success: '#34D399',
      },
    },
  },
  plugins: [],
}
