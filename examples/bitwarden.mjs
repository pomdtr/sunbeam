#!/usr/bin/env zx

const { stdout } = await $`bw list items`;
const items = JSON.parse(stdout).map((item) => ({
  title: item.name,
  subtitle: item.login?.username || "",
  actions: [
    {
      type: "copy",
      title: "Copy Password",
      content: item.login?.password || "",
    },
  ],
}));

for (const item of items) {
  console.log(JSON.stringify(item));
}
