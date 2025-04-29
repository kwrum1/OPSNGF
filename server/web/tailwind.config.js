/** @type {import('tailwindcss').Config} */
export default {
    darkMode: ["class"],
    content: ["./index.html", "./src/**/*.{ts,tsx,js,jsx}"],
    theme: {
        extend: {
            borderRadius: {
                lg: 'var(--radius)',
                md: 'calc(var(--radius) - 2px)',
                sm: 'calc(var(--radius) - 4px)'
            },
            colors: {
                background: 'hsl(var(--background))',
                foreground: 'hsl(var(--foreground))',
                card: {
                    DEFAULT: 'hsl(var(--card))',
                    foreground: 'hsl(var(--card-foreground))'
                },
                popover: {
                    DEFAULT: 'hsl(var(--popover))',
                    foreground: 'hsl(var(--popover-foreground))'
                },
                primary: {
                    DEFAULT: 'hsl(var(--primary))',
                    foreground: 'hsl(var(--primary-foreground))',
                },
                secondary: {
                    DEFAULT: 'hsl(var(--secondary))',
                    foreground: 'hsl(var(--secondary-foreground))'
                },
                muted: {
                    DEFAULT: 'hsl(var(--muted))',
                    foreground: 'hsl(var(--muted-foreground))'
                },
                accent: {
                    DEFAULT: 'hsl(var(--accent))',
                    foreground: 'hsl(var(--accent-foreground))',
                },
                destructive: {
                    DEFAULT: 'hsl(var(--destructive))',
                    foreground: 'hsl(var(--destructive-foreground))'
                },
                border: 'hsl(var(--border))',
                input: 'hsl(var(--input))',
                ring: 'hsl(var(--ring))',
                chart: {
                    '1': 'hsl(var(--chart-1))',
                    '2': 'hsl(var(--chart-2))',
                    '3': 'hsl(var(--chart-3))',
                    '4': 'hsl(var(--chart-4))',
                    '5': 'hsl(var(--chart-5))'
                },
            },
            keyframes: {
                'icon-shake': {
                    '0%': { transform: 'rotate(0deg)' },
                    '25%': { transform: 'rotate(-12deg)' },
                    '50%': { transform: 'rotate(10deg)' },
                    '75%': { transform: 'rotate(-6deg)' },
                    '85%': { transform: 'rotate(3deg)' },
                    '92%': { transform: 'rotate(-2deg)' },
                    '100%': { transform: 'rotate(0deg)' }
                },
                'float': {
                    '0%': { transform: 'translateY(0px) translateX(0px)' },
                    '50%': { transform: 'translateY(-20px) translateX(10px)' },
                    '100%': { transform: 'translateY(0px) translateX(0px)' }
                },
                'float-reverse': {
                    '0%': { transform: 'translateY(0px) translateX(0px)' },
                    '50%': { transform: 'translateY(20px) translateX(-10px)' },
                    '100%': { transform: 'translateY(0px) translateX(0px)' }
                },
                'pulse-glow': {
                    '0%, 100%': { opacity: 0.6, transform: 'scale(1)' },
                    '50%': { opacity: 1, transform: 'scale(1.05)' }
                },
                'fade-in-up': {
                    '0%': { opacity: 0, transform: 'translateY(20px)' },
                    '100%': { opacity: 1, transform: 'translateY(0)' }
                },
                'wiggle': {
                    '0%, 100%': { transform: 'rotate(-2deg)' },
                    '50%': { transform: 'rotate(2deg)' }
                },
                'gradient-shift': {
                    '0%': { backgroundPosition: '0% 50%' },
                    '50%': { backgroundPosition: '100% 50%' },
                    '100%': { backgroundPosition: '0% 50%' }
                },
                'aurora': {
                    '0%': { filter: 'hue-rotate(0deg) brightness(1) saturate(1.5)' },
                    '33%': { filter: 'hue-rotate(60deg) brightness(1.1) saturate(1.8)' },
                    '66%': { filter: 'hue-rotate(180deg) brightness(1.05) saturate(1.6)' },
                    '100%': { filter: 'hue-rotate(360deg) brightness(1) saturate(1.5)' }
                },
                'logo-pulse': {
                    '0%, 100%': { opacity: 0.9, transform: 'scale(0.95)' },
                    '50%': { opacity: 1, transform: 'scale(1.05)' }
                }
            },
            animation: {
                'icon-shake': 'icon-shake 0.7s ease-out',
                'float': 'float 8s ease-in-out infinite',
                'float-reverse': 'float-reverse 9s ease-in-out infinite',
                'pulse-glow': 'pulse-glow 4s ease-in-out infinite',
                'fade-in-up': 'fade-in-up 0.5s ease-out',
                'wiggle': 'wiggle 1s ease-in-out infinite',
                'gradient-shift': 'gradient-shift 8s ease infinite',
                'aurora': 'aurora 20s ease infinite',
                'logo-pulse': 'logo-pulse 1.5s infinite ease-in-out'
            }
        }
    },
    plugins: [require("tailwindcss-animate")],
}

