#!/usr/bin/env zx

const { stdout } = await $`alacritty-themes --list`;
const themes = stdout
  .split("\n")
  .slice(0, -1)
  .map((row) => row.split(" ")[1]);
const items = themes.map((theme) => {
  return {
    title: theme,
    actions: [
      {
        type: "exec",
        title: "Use Theme",
        command: ["alacritty-themes", theme],
      },
    ],
  };
});

for (const item of items) {
  console.log(JSON.stringify(item));
}
