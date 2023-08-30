import { Hono } from "https://deno.land/x/hono@v3.3.1/mod.ts";

const app = new Hono();

app.get("/", (c) =>
  c.json({
    title: "Example server",
  })
);

app.post("/", async (c) => {
  const { args } = await c.req.json();
  console.log(args);

  return c.json({
    type: "list",
    items: [{ title: "Item 1" }],
  });
});

Deno.serve({ port: 8000 }, app.fetch);
