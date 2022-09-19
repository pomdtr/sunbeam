#!/usr/bin/env zx

// @raycast.schemaVersion 1
// @raycast.title List My Repositories
// @raycast.mode command
// @raycast.packageName Github

$.verbose = false;

if (argv.pr) {
  const out = await $`gh pr list --repo ${argv.pr} --json title,url`;
  const prs = JSON.parse(out);
  console.log(
    JSON.stringify({
      type: "list",
      list: {
        items: prs.map((pr) => ({
          title: pr.title,
          icon: "https://github.githubassets.com/favicons/favicon.svg",
          actions: [
            {
              type: "open-url",
              title: "Open in Browser",
              url: pr.url,
            },
          ],
        })),
      },
    })
  );
  process.exit(0);
}

var res;
if (argv.owner) {
  res =
    await $`gh repo list ${argv.owner} --json nameWithOwner,description,url,stargazerCount,languages`;
} else {
  res = await $`gh api /users/pomdtr/repos --cache 3600s`;
}
const repos = JSON.parse(res);

console.log(
  JSON.stringify({
    type: "list",
    list: {
      title: "Repositories",
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
            type: "callback",
            title: "List PRs",
            args: ["--pr", repo.full_name],
          },
        ],
      })),
    },
  })
);
