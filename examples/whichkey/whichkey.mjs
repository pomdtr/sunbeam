#!/usr/bin/env zx

// @sunbeam.schemaVersion 1
// @sunbeam.title Which Key
// @sunbeam.mode interactive
// @sunbeam.packageName WhichKey

let root = {
  title: "Which Key",
  children: {
    a: {
      title: "Applications",
      children: {
        b: {
          title: "Browser",
          type: "open",
          path: "/Applications/Arc.app",
        },
        c: {
          title: "Calendar",
          type: "open",
          path: "/Applications/Calendar.app",
        },
        n: {
          title: "Notes",
          type: "open",
          path: "/Applications/Notes.app",
        },
        r: {
          title: "Reminders",
          type: "open",
          path: "/Applications/Reminders.app",
        },
        s: {
          title: "Slack",
          type: "open",
          path: "/Applications/Slack.app",
        },
      },
    },
    l: {
      title: "Links",
      children: {
        g: {
          title: "Github",
          children: {
            h: {
              title: "Home",
              type: "open-url",
              url: "http://github.com",
            },
            n: {
              title: "Notifications",
              type: "open-url",
              url: "https://github.com/notifications",
            },
            s: {
              title: "StackOverflow",
              type: "open-url",
              url: "https://github.com/pulls",
            },
          },
        },
        s: {
          title: "StackOverflow",
          type: "open-url",
          url: "https://stackoverflow.com",
        },
        t: {
          title: "Twitter",
          type: "open-url",
          url: "https://twitter.com",
        },
        y: {
          title: "Youtube",
          type: "open-url",
          url: "https://youtube.com",
        },
      },
    },
  },
};

const args = argv._;
for (const arg of args) {
  root = root.children[arg];
}

const view = {
  type: "detail",
  detail: {
    format: "markdown",
    text: Object.entries(root.children)
      .map(([keybind, child]) => `**${keybind}** => ${child.title}`)
      .join("\n\n"),
    actions: Object.entries(root.children).map(([keybind, item]) => {
      if (item.children) {
        return {
          title: item.title,
          type: "push",
          keybind,
          path: "./whichkey.mjs",
          args: [...args, keybind],
        };
      } else {
        return { ...item, keybind };
      }
    }),
  },
};

console.log(JSON.stringify(view));
