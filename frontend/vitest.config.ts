import { defineConfig } from "vitest/config"
import vue from "@vitejs/plugin-vue"

export default defineConfig({
  plugins: [vue()],
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
    include: ["src/**/*.{test,spec}.{js,ts}"],
    coverage: {
      provider: "v8",
      reporter: ["text", "json", "html"],
      include: ["src/**/*.{ts,vue}"],
      exclude: ["src/main.ts", "src/bindings/**", "src/test/**", "**/*.d.ts"],
    },
  },
})
