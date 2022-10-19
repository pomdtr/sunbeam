#!/usr/bin/env zx

// @sunbeam.schemaVersion 1
// @sunbeam.title List Pull Requests
// @sunbeam.subtitle Github
// @sunbeam.mode interactive

// @sunbeam.argument1 { "type": "text", "placeholder": "repository" }

const repo = argv._[0];

const res = await $`gh api /repos/${repo}/pulls`;
const prs = JSON.parse(res);
const view = {
  type: "list",
  list: {
    title: "Pull Requests",
    items: prs.map((pr) => ({
      title: pr.title,
      subtitle: pr.user.login,
      icon: "🔍",
      actions: [
        {
          type: "open-url",
          title: "Open in Browser",
          url: pr.html_url,
        },
      ],
    })),
  },
};

console.log(JSON.stringify(view));
