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
        surface: withOpacity("--color-surface"),
        ink: withOpacity("--color-ink"),
        accent: {
          DEFAULT: withOpacity("--color-accent"),
          foreground: withOpacity("--color-accent-foreground")
        },
        border: withOpacity("--color-border"),
        muted: withOpacity("--color-muted"),
        panel: withOpacity("--color-panel"),
        "accent-soft": withOpacity("--color-accent-soft"),
        code: withOpacity("--color-code"),
        "code-foreground": withOpacity("--color-code-foreground"),
        danger: withOpacity("--color-danger"),
        "danger-foreground": withOpacity("--color-danger-foreground")
      },
      boxShadow: {
        panel: "0 20px 60px rgba(15, 23, 42, 0.14)",
        focus: "0 0 0 4px rgba(139, 92, 246, 0.2)",
        glow: "0 18px 48px rgba(139, 92, 246, 0.22)"
      },
      backgroundImage: {
        "app-mesh": "radial-gradient(circle at 14% 10%, rgba(139, 92, 246, 0.18), transparent 28%), radial-gradient(circle at 86% 24%, rgba(6, 182, 212, 0.14), transparent 30%), linear-gradient(180deg, var(--app-base-from), var(--app-base-to))",
        "glass-sheen": "linear-gradient(135deg, rgba(255,255,255,0.42), rgba(255,255,255,0.08) 42%, rgba(6,182,212,0.08))",
        "primary-gradient": "linear-gradient(135deg, rgb(124, 58, 237), rgb(8, 145, 178))"
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
