import type { Config } from "tailwindcss";

const withOpacity = (variableName: string) => `rgb(var(${variableName}) / <alpha-value>)`;

const config: Config = {
  darkMode: "class",
  content: ["./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      fontFamily: {
        sans: ["Inter", "ui-sans-serif", "system-ui", "-apple-system", "BlinkMacSystemFont", "Segoe UI", "sans-serif"]
      },
      colors: {
        background: withOpacity("--background"),
        foreground: withOpacity("--foreground"),
        card: {
          DEFAULT: withOpacity("--card"),
          foreground: withOpacity("--card-foreground")
        },
        popover: {
          DEFAULT: withOpacity("--popover"),
          foreground: withOpacity("--popover-foreground")
        },
        primary: {
          DEFAULT: withOpacity("--primary"),
          foreground: withOpacity("--primary-foreground")
        },
        secondary: {
          DEFAULT: withOpacity("--secondary"),
          foreground: withOpacity("--secondary-foreground")
        },
        destructive: {
          DEFAULT: withOpacity("--destructive"),
          foreground: withOpacity("--destructive-foreground")
        },
        ring: withOpacity("--ring"),
        input: withOpacity("--input"),
        surface: withOpacity("--color-surface"),
        ink: withOpacity("--color-ink"),
        accent: {
          DEFAULT: withOpacity("--accent"),
          foreground: withOpacity("--accent-foreground")
        },
        brand: {
          DEFAULT: withOpacity("--color-accent"),
          foreground: withOpacity("--color-accent-foreground")
        },
        border: withOpacity("--color-border"),
        muted: {
          DEFAULT: withOpacity("--muted"),
          foreground: withOpacity("--muted-foreground")
        },
        panel: withOpacity("--color-panel"),
        "accent-soft": withOpacity("--color-accent-soft"),
        code: withOpacity("--color-code"),
        "code-foreground": withOpacity("--color-code-foreground"),
        danger: withOpacity("--color-danger"),
        "danger-foreground": withOpacity("--color-danger-foreground")
      },
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)"
      },
      boxShadow: {
        panel: "0 1px 2px rgba(15, 23, 42, 0.06)",
        focus: "0 0 0 3px rgb(var(--ring) / 0.18)",
        glow: "0 8px 24px rgba(15, 23, 42, 0.08)"
      },
      backgroundImage: {
        "app-mesh": "linear-gradient(180deg, var(--app-base-from), var(--app-base-to))",
        "glass-sheen": "linear-gradient(180deg, rgb(var(--card)), rgb(var(--card)))",
        "primary-gradient": "linear-gradient(180deg, rgb(var(--primary)), rgb(var(--primary)))"
      },
      animation: {
        "fade-in": "fadeIn 220ms ease-out both",
        "slide-up": "slideUp 320ms cubic-bezier(0.16, 1, 0.3, 1) both",
        "scale-in": "scaleIn 220ms ease-out both",
        "pulse-glow": "pulseGlow 1.8s ease-in-out infinite",
        shimmer: "shimmer 2s linear infinite"
      },
      keyframes: {
        fadeIn: {
          "0%": { opacity: "0" },
          "100%": { opacity: "1" }
        },
        slideUp: {
          "0%": { opacity: "0", transform: "translateY(16px)" },
          "100%": { opacity: "1", transform: "translateY(0)" }
        },
        scaleIn: {
          "0%": { opacity: "0", transform: "scale(0.96)" },
          "100%": { opacity: "1", transform: "scale(1)" }
        },
        pulseGlow: {
          "0%, 100%": { boxShadow: "0 0 28px rgba(139, 92, 246, 0.18)" },
          "50%": { boxShadow: "0 0 54px rgba(139, 92, 246, 0.34), 0 0 76px rgba(6, 182, 212, 0.16)" }
        },
        shimmer: {
          "0%": { backgroundPosition: "-200% 0" },
          "100%": { backgroundPosition: "200% 0" }
        }
      }
    }
  },
  plugins: []
};

export default config;
