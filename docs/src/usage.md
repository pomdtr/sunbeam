# Usage

## Running Sunbeam

To run sunbeam, you need to provide a script as the first argument. Sunbeam will run the script and interpret the output as a JSON object.

```bash
sunbeam github.sh
```

Once the output is shown, you can use the arrow keys to navigate the UI, and press enter to select an item. Use the tab key to show all available actions for the selected item.

Depending on the action, sunbeam may:

- Open an URI in the default application/browser.
- Copy text to the system clipboard.
- Run a command, and display its output on stdout.
- Push a new page to the navigation stack
- Pop the current page from the navigation stack
- Reload the current page

The flow of data in sunbeam is as follows:

![Event Loop](./assets/event-loop.excalidraw.png)

Using `push` actions, you can build complex UIs by composing multiple pages.
