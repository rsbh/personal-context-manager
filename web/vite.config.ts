import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";

const api = "http://127.0.0.1:8080";

export default defineConfig({
  plugins: [react()],
  build: {
    // dist/app is embedded by web/embed.go. dist/.gitkeep stays.
    outDir: "dist/app",
    emptyOutDir: true,
  },
  server: {
    // npm run dev: API calls go to a running pcm-server.
    proxy: { "/api": api, "/health": api },
  },
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./src/test/setup.ts"],
    include: ["src/**/*.test.{ts,tsx}"],
  },
});
