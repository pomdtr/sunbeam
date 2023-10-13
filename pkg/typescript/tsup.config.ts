import { defineConfig } from "tsup";

export default defineConfig({
  entry: ["src/*.ts"],
  dts: {
    only: true,
  },
  clean: true,
});
