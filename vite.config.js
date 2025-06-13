import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  publicDir: false,
  build: {
    outDir: "internal/server/public/js",
    sourcemap: false,
    minify: false,
    terserOptions: {
      mangle: false, // Disable variable renaming
      compress: false, // You might want to disable compression too
      format: {
        beautify: true, // Make the output readable
      },
      keep_classnames: true,
      keep_fnames: true,
    },
    rollupOptions: {
      input: "internal/server/islands/islands.tsx",
      output: {
        entryFileNames: "islands.js",
      },
    },
  },
});
