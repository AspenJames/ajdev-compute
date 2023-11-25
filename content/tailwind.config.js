/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./templates/**/*.html", "./templates/*.html"],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        "ajdev-pink-dark": "#ffa4ff",
        "ajdev-blue-dark": "#0ff",
        "ajdev-text-dark": "#fff",
        "ajdev-pink": "#C000C0",
        "ajdev-blue": "#007979",
        "ajdev-text": "#333",
        "bg-navbar": "#0ff",
        "bg-light": "#fafafa",
        "bg-accent-light": "#fff",
        "bg-dark": "#303030",
        "bg-accent-dark": "#424242",
      },
    },
  },
  plugins: [],
};
