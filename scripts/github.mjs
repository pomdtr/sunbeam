#!/usr/bin/env zx

// @sunbeam.schemaVersion 1
// @sunbeam.title Search My Repositories
// @sunbeam.mode command
// @sunbeam.packageName Github

$.verbose = false;

async function viewReadme(repo) {
  const res = await fetch(
    `https://raw.githubusercontent.com/${repo}/master/README.md`
  );
  const readme = await res.text();
  output({
    type: "detail",
    detail: {
      text: readme,
      actions: [{ type: "open-url", url: `https://github.com/${repo}` }],
    },
  });
}

async function listRepos() {
  const res = await $`/usr/local/bin/gh api /users/pomdtr/repos --cache 3600s`;
  const repos = JSON.parse(res);
  return {
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
            title: "View Readme",
            keybind: "ctrl+r",
            params: {
              type: "readme",
              repo: repo.full_name,
            },
          },
          {
            type: "callback",
            title: "List PRs",
            keybind: "ctrl+p",
            params: {
              type: "pr",
              repo: repo.full_name,
            },
          },
        ],
      })),
    },
  };
}

async function listPRs(repo) {
  const res = await $`/usr/local/bin/gh api /repos/${repo}/pulls`;
  const prs = JSON.parse(res);
  return {
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
}

function output(data) {
  console.log(JSON.stringify(data));
  process.exit(0);
}

// main

const { params } = JSON.parse(await stdin());

if (!params) {
  output(await listRepos());
}

switch (params.type) {
  case "readme":
    const readme = await viewReadme(params.repo);
    output(readme);
  case "pr":
    const prs = await listPRs(params.repo);
    output(prs);
}
