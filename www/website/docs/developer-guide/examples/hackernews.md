# Hacker News (Deno)

## Writing the manifest

When the script is called without arguments, it must return a json manifest describing the extension and its commands.

```ts
import * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";

const manifest = {
  title: "Hacker News",
  description: "Browse Hacker News",
  commands: [
    {
      name: "browse",
      title: "Show a feed",
      mode: "filter",
      params: [{ name: "topic", title: "Topic", type: "string" }],
    },
  ],
} as const satisfies sunbeam.Manifest;

if (Deno.args.length == 0) {
  console.log(JSON.stringify(manifest));
  Deno.exit(0);
}
```

Notice that we use the as const keyword here (it will come in handy later).
We also use the satisfies keyword to make sure that the manifest is correctly typed.

Let's make sure that our manifest is valid: `./hackernews.ts | sunbeam validate manifest`
Then install the extension: `sunbeam extension install ./hackernews.ts`

Now if we run `sunbeam hackernews --help` we should see the generated help.

```txt
sunbeam hackernews --help
Browse Hacker News

Usage:
  sunbeam hackernews [flags]
  sunbeam hackernews [command]

Available Commands:
  browse      Show a feed

Flags:
  -h, --help   help for hackernews

Use "sunbeam hackernews [command] --help" for more information about a command.
```

## Writing the command

We will use the <https://hnrss.org/> api to get the hackernews feeds as rss.
Deno does not have a built-in xml parser, so we will use the [rss-parser](https://www.npmjs.com/package/rss-parser) npm package.

```ts
import Parser from "npm:rss-parser";
import * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";
```

Now we can implement the command.

```ts
#!/usr/bin/env -S deno run -A

import Parser from "npm:rss-parser";
import { formatDistance } from "npm:date-fns";
import * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";

const manifest = {
  title: "Hacker News",
  description: "Browse Hacker News",
  commands: [
    {
      name: "browse",
      title: "Show a feed",
      mode: "filter",
      params: [{ name: "topic", title: "Topic", type: "string" }],
    },
  ],
} as const satisfies sunbeam.Manifest;

if (Deno.args.length == 0) {
  console.log(JSON.stringify(manifest));
  Deno.exit(0);
}

const payload: sunbeam.Payload<typeof manifest> = JSON.parse(Deno.args[0]);
if (payload.command == "browse") {
  const { topic } = payload.params;
  const feed = await new Parser().parseURL(
    `https://hnrss.org/${topic}?description=0&count=25`
  );
  const page: sunbeam.List = {
    items: feed.items.map((item) => ({
      title: item.title || "",
      subtitle: item.categories?.join(", ") || "",
      accessories: item.isoDate
        ? [
            formatDistance(new Date(item.isoDate), new Date(), {
              addSuffix: true,
            }),
          ]
        : [],
      actions: [
        {
          title: "Open in browser",
          type: "open",
          url: item.link || "",
        },
        {
          title: "Open Comments in Browser",
          type: "open",
          url: item.guid || "",
        },
        {
          title: "Copy Link",
          type: "copy",
          key: "c",
          text: item.link || "",
          exit: true,
        },
      ],
    })),
  };

  console.log(JSON.stringify(page));
} else {
  console.error("Unknown command");
  Deno.exit(1);
}
```

Our command takes a `topic` parameter, which is used to build the url to fetch the feed.

If you trigger the command from the root list, a form will be shown to prompt the user for the topic.
If you run the command from the cli, you will need to pass the topic as a parameter: `sunbeam hackernews browse --topic=showhn`

## Additional Root Commands

As a user of the extension, you can add new items to the root list by editing the sunbeam config file.

```json
{
  "extensions": {
    "hackernews": {
      "origin": "...",
      "root": [
        {
          "title": "Show HN",
          "command": "browse",
          "params": {
            "topic": "show"
          }
        }
      ]
    }
  }
}
```
