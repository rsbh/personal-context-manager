import { defineConfig } from "@hey-api/openapi-ts";

// Generates the typed API client from the swagger spec (make swagger).
// Run: npm run gen:api. The output is in git; CI checks it is up to date.
export default defineConfig({
  input: "../docs/swagger/swagger.json",
  output: "src/client",
  plugins: ["@hey-api/client-fetch", "@hey-api/typescript", "@hey-api/sdk", "@tanstack/react-query"],
});
