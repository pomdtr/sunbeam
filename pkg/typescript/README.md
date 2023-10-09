# Sunbeam SDK

## Usage (Deno)

```typescript
import {Extension} from "npm:sunbeam-sdk"

const extension = new Extension({
    title: "My Extension"
})

extension.addCommand({
    title: "My Command",
    name: "my-command",
    mode: "view",
}, () => (
    return {
        type: "detail",
        markdown: "Hello World!"
    }
))

if (Deno.args == 0) {
    console.log(JSON.stringify(extension))
} else {
    const res = await extension.run(Deno.args)
    console.log(JSON.stringify(res))
}
```
