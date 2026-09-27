import { defineConfig } from "astro/config";

export default defineConfig({
  site: "https://golangskills.com",
  srcDir: "./src",
  publicDir: "./public",
  outDir: "../site",
  trailingSlash: "always",
  build: {
    format: "directory",
    emptyOutDir: true,
  },
});
