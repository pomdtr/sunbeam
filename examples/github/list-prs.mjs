#!/usr/bin/env zx

const repo = argv._[0];

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
