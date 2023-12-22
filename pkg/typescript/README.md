# Sunbeam SDK

## Type Validation for Sunbeam Scripts

## Node

```typescript
import * as sunbeam from "sunbeam-sdk";

if (process.args.length < 3) {
    const manifest: sunbeam.Manifest = {...}
    console.log(manifest);
    Deno.exit(0);
}

const payload: sunbeam.Payload = JSON.parse(process.args[2]);

if (payload.command = "show") {
    const list: sunbeam.List = {...}
    console.log(JSON.stringify(list));
}
```

## Deno

```typescript
import * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";

if (Deno.args.length === 0) {
    const manifest: sunbeam.Manifest = {...}
    console.log(manifest);
    Deno.exit(0);
}

const payload: sunbeam.Payload = JSON.parse(Deno.args[0]);

if (payload.command = "show") {
    const list: sunbeam.List = {...}
    console.log(JSON.stringify(list));
}

```

## Helper Functions (Only for Deno)

```typescript
import { editor } from "https://deno.land/x/sunbeam/editor.ts";

// ...

if (payload.command === "edit") {
    // open an editor and wait for the user to save and exit
    const edited = await editor(payload.text);
}
```
