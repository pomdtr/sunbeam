const { stdout } = await $`tldr --list`;

const commands = stdout.split("\n").slice(0, -1);

const items = commands.map((item) => ({
  title: item,
  detail: {
    command: `tldr ${item} --raw`,
  },
}));

for (const item of items) {
  console.log(JSON.stringify(item));
}
