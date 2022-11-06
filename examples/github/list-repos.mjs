#!/usr/bin/env zx

const owner = argv.owner;

if (!owner) {
  console.error("Please specify the owner");
  process.exit(1);
}

const res = await $`gh api /users/${owner}/repos --cache 3600s`;
const repos = JSON.parse(res);

const items = repos.map((repo) => ({
  title: repo.name,
  subtitle: repo.owner.login,
  accessories: [`${repo.language}`, `${repo.stargazers_count} *`],
  actions: [
    {
      type: "open-url",
      title: "Open in Browser",
      shortcut: "enter",
      url: repo.html_url,
    },
    {
      title: "List Pull Requests",
      type: "push",
      page: "list-pull-requests",
      shortcut: "ctrl+p",
      with: {
        repository: repo.full_name,
      },
    },
  ],
}));

for (const item of items) {
  console.log(JSON.stringify(item));
}
