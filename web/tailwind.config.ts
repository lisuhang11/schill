import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./app/**/*.{ts,tsx}",
    "./components/**/*.{ts,tsx}",
    "./lib/**/*.{ts,tsx}"
  ],
  theme: {
    extend: {
      colors: {
        marine: {
          bg: "#EEFBFF",
          blue: "#61C9F5",
          deep: "#2F5F9E",
          text: "#20303F",
          muted: "#687789",
          mint: "#67D8C6",
          warm: "#FFD678",
          pink: "#FF9BBC"
        }
      },
      boxShadow: {
        soft: "0 12px 32px rgba(45, 68, 92, 0.10)",
        float: "0 18px 48px rgba(45, 68, 92, 0.14)"
      }
    }
  },
  plugins: []
};

export default config;
