import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  publicDir: false,
  build: {
    outDir: "internal/server/public/js",
    sourcemap: false,
    rollupOptions: {
      input: "internal/server/islands/islands.tsx",
      output: {
        entryFileNames: "islands.js",
      },
    },
  },
});
