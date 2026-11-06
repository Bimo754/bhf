/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        canvas: {
          DEFAULT: '#080c14',
          subtle: '#0e1420',
          elevated: '#141b2b',
          inset: '#05080e',
        },
        border: {
          DEFAULT: 'rgba(255, 255, 255, 0.07)',
          subtle: 'rgba(255, 255, 255, 0.04)',
          focus: 'rgba(99, 102, 241, 0.5)',
        },
        fg: {
          DEFAULT: '#f1f5f9',
          muted: '#94a3b8',
          subtle: '#64748b',
        },
        accent: {
          DEFAULT: '#4f46e5',
          hover: '#6366f1',
          blue: '#3b82f6',
          emerald: '#10b981',
          amber: '#f59e0b',
          rose: '#f43f5e',
          purple: '#a855f7',
        }
      },
      fontFamily: {
        mono: ['ui-monospace', 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', 'monospace'],
        sans: ['Inter', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Helvetica', 'Arial', 'sans-serif'],
      }
    },
  },
  plugins: [],
}
