#!/usr/bin/env zx

// @sunbeam.schemaVersion 1
// @sunbeam.title Which Key
// @sunbeam.mode command
// @sunbeam.packageName WhichKey

const rootActions = [
  {
    keybind: "a",
    title: "Applications",
    type: "callback",
    params: [
      {
        keybind: "b",
        title: "Browser",
        type: "open",
        path: "/Applications/Arc.app",
      },
      {
        keybind: "c",
        title: "Calendar",
        type: "open",
        path: "/Applications/Calendar.app",
      },
      {
        keybind: "n",
        title: "Notes",
        type: "open",
        path: "/Applications/Notes.app",
      },
      {
        keybind: "r",
        title: "Reminders",
        type: "open",
        path: "/Applications/Reminders.app",
      },
      {
        keybind: "s",
        title: "Slack",
        type: "open",
        path: "/Applications/Slack.app",
      },
    ],
  },
  {
    keybind: "l",
    title: "Links",
    type: "callback",
    params: [
      {
        keybind: "g",
        title: "Github",
        type: "callback",
        params: [
          {
            keybind: "s",
            title: "Home",
            type: "open-url",
            url: "http://github.com",
          },
          {
            keybind: "s",
            title: "Notifications",
            type: "open-url",
            url: "https://github.com/notifications",
          },
          {
            keybind: "s",
            title: "StackOverflow",
            type: "open-url",
            url: "https://github.com/pulls",
          },
        ],
      },
      {
        keybind: "s",
        title: "StackOverflow",
        type: "open-url",
        url: "https://stackoverflow.com",
      },
      {
        keybind: "t",
        title: "Twitter",
        type: "open-url",
        url: "https://twitter.com",
      },
      {
        keybind: "y",
        title: "Youtube",
        type: "open-url",
        url: "https://youtube.com",
      },
    ],
  },
];

const menuToView = (actions) => {
  return {
    type: "detail",
    detail: {
      format: "markdown",
      text: actions
        .map((action) => `**${action.keybind}** => ${action.title}`)
        .join("\n\n"),
      actions,
    },
  };
};

const output = (data) => {
  console.log(JSON.stringify(data));
};

const { params: actions } = JSON.parse(await stdin());

if (!actions) {
  output(menuToView(rootActions));
  process.exit(0);
}

output(menuToView(actions));
process.exit(0);
