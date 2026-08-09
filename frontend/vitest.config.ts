import react from "@vitejs/plugin-react";
import { readFileSync } from "node:fs";
import { defineConfig } from "vitest/config";

const version = readFileSync(new URL("../VERSION", import.meta.url), "utf8").trim();

export default defineConfig({
  plugins: [react()],
  define: {
    __APP_VERSION__: JSON.stringify(version),
  },
  test: {
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
    testTimeout: 20_000,
    coverage: {
      provider: "v8",
      reporter: ["text", "json-summary"],
      include: ["src/**/*.{ts,tsx}"],
      exclude: [
        "src/**/*.test.{ts,tsx}",
        "src/shared/api/generated.ts",
        "src/main.tsx",
        "src/test/**",
        "src/vite-env.d.ts",
      ],
      thresholds: {
        statements: 80,
        branches: 73,
        functions: 80,
        lines: 82,
      },
    },
  },
});
