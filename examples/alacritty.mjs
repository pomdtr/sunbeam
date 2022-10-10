#!/usr/bin/env npx zx

// @sunbeam.schemaVersion 1
// @sunbeam.title Change Alacritty Theme
// @sunbeam.packageName Alacritty
// @sunbeam.mode interactive

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

console.log(
  JSON.stringify({
    type: "list",
    list: {
      title: "Alacritty Themes",
      items: items,
    },
  })
);
