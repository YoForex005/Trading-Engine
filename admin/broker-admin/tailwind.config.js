/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
        "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
        "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
    ],
    theme: {
        extend: {
            colors: {
                background: "var(--background)",
                foreground: "var(--foreground)",
                // RTX Broker Admin Theme
                charcoal: {
                    950: '#0E0F11', // Main Background
                    900: '#141414', // Secondary Background
                    850: '#16181C', // Panels
                    800: '#1C1F26', // Hover states
                    border: '#2A2A2A', // Sharp borders
                },
                rtx: {
                    yellow: '#F5C542', // Primary
                    hover: '#FFD766',  // Primary Hover
                    muted: '#666666',  // Muted text
                    active: '#2ECC71', // Green
                    suspended: '#E74C3C', // Red
                }
            },
            fontFamily: {
                sans: ['var(--font-inter)', 'sans-serif'],
            },
        },
    },
    plugins: [],
};
