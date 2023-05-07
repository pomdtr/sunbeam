import * as sunbeam from "https://deno.land/x/sunbeam@v0.9.23/index.d.ts";

const page: sunbeam.Page = {
  type: "detail",
  preview: {
    text: "Hello, world!",
  },
  actions: [
    {
      type: "copy",
      text: "Hello, world!",
    },
  ],
};

console.log(JSON.stringify(page));
