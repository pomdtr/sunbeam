async function trendingRepositories() {
  const repos = await fetch(
    "https://api.github.com/search/repositories?q=created:>2017-10-22&sort=stars&order=desc"
  );
}
