#!/usr/bin/env zx

const res = await $`gh api /users/pomdtr/repos --cache 3600s`;
const repos = JSON.parse(res);

const items = repos.map((repo) => ({
  title: repo.full_name,
  subtitle: `${repo.stargazers_count} ⭐`,
  actions: [
    {
      type: "open-url",
      title: "Open in Browser",
      url: repo.html_url,
    },
    {
      type: "push",
      target: "list-prs",
      keybind: "ctrl+p",
      params: {
        repository: repo.full_name,
      },
    },
  ],
}));

for (const item of items) {
  console.log(JSON.stringify(item));
}
