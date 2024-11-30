# Tips

## Non Interactive Mode

When the stdin or stdout is not a tty, sunbeam will adapt it's behavior:

- If stdin is not a tty, sunbeam will try to read i's input from stdin. It allows you to pipe the sunbeam payload instead of using positional arguments.

  ```sh
  jq '{ command: "list-docsets" } | sunbeam devdocs
  ```

- If stdout is not a tty, sunbeam will output the result as a JSON string. It allows you to pipe the result to another program.

  ```sh
  # print the config
  sunbeam | jq
  # print the manifest of the devdocs extension
  sunbeam devdocs | jq
  # print the json representation of the go entries
  sunbeam devdocs list-entries --slug go | jq
  ```

You can combine both to create a pipeline:

```sh
jq '{ command: "list-docsets" }' | sunbeam devdocs | jq
```

## Extension Validation

The sunbeam validate command allows you to validate the config file, the manifest of an extension, or the output of a command.

```sh
sunbeam validate config
./devdocs.sh | sunbeam validate manifest
sunbeam devdocs list-docsets | sunbeam validate list
```

You can use those commands to validate an extension in a CI pipeline.

## Workspace Structure

You are free to store your local extensions anywhere you want. I personally store them directly in the sunbeam config directory.

```txt
~/.config/sunbeam/
├── sunbeam.json
└── extensions/
    ├── devdocs.sh
    └── github.ts
```

And then reference them in the config file using relative paths:

```json
{
  "extensions": {
    "devdocs": {
      "origin": "./extensions/devdocs.sh"
    },
    "github": {
      "origin": "./extensions/github.ts"
    }
  }
}
```

## Additional Tools

Sunbeam pages are described using JSON, so it pairs really well with other JSON tools:

- [jq](https://github.com/jqlang/jq) - a lightweight and flexible command-line JSON processor
- [jc](https://github.com/kellyjonbrazil/jc) - a command line utility to convert the output of popular command-line tools and file-types to JSON
- [pup](https://github.com/ericchiang/pup) - a command line tool for processing HTML, can be used to extract data from HTML pages and convert it to JSON
- [jo](https://github.com/jpmens/jo) - a command line utility to create JSON objects
- [bkt](https://github.com/dimo414/bkt) - a caching tool for for subprocess output
