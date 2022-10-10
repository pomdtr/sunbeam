#!/usr/bin/env zx

// @sunbeam.schemaVersion 1
// @sunbeam.title Search My Repositories
// @sunbeam.mode interactive
// @sunbeam.packageName Github

$.verbose = false;

const res = await $`gh api /users/pomdtr/repos --cache 3600s`;
const repos = JSON.parse(res);
const view = {
  type: "list",
  list: {
    items: repos.map((repo) => ({
      title: repo.full_name,
      subtitle: `${repo.stargazers_count} ⭐`,
      icon: "🔍",
      actions: [
        {
          type: "open-url",
          title: "Open in Browser",
          url: repo.html_url,
        },
        {
          type: "push",
          title: "List Prs",
          keybind: "ctrl+a",
          path: "./list-prs.mjs",
          args: [repo.full_name],
        },
        {
          type: "push",
          title: "View README",
          keybind: "ctrl+r",
          path: "./view-readme.mjs",
          args: [repo.full_name],
        },
      ],
    })),
  },
};

console.log(JSON.stringify(view));
