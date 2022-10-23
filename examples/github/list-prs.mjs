#!/usr/bin/env zx

let repo = argv.repo;

if (!repo) {
  const { stdout: remote } = await $`git config --get remote.origin.url`;
  const match = remote.match(/github.com[:/](.*)\.git/);
  repo = match[1];
}

const res = await $`gh api /repos/${repo}/pulls`;
const prs = JSON.parse(res);
const items = prs.map((pr) => ({
  title: pr.title,
  subtitle: pr.user.login,
  actions: [
    {
      type: "open-url",
      title: "Open in Browser",
      url: pr.html_url,
    },
  ],
}));

for (const item of items) {
  console.log(JSON.stringify(item));
}
