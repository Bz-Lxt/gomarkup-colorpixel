/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{vue,ts}"],
  theme: {
    extend: {
      colors: {
        ink: "#0B0C0E",
        panel: "#14161A",
        line: "#2A2E35",
        tungsten: "#E8A14A",
        film: "#E8DFD0",
        mute: "#8A8478",
        alert: "#D4533B",
        safe: "#6F9E7A",
      },
      fontFamily: {
        display: ["Fraunces", "serif"],
        body: ["Source Serif 4", "serif"],
        mono: ["IBM Plex Mono", "ui-monospace", "monospace"],
      },
    },
  },
  plugins: [],
};
