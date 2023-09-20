import { Extension } from "./extension.ts";

new Extension({
  title: "Example",
}).command("list", {
  mode: "page",
  title: "List",
  run: async ({ params }) => {
  },
});
