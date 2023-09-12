import type { Extension } from "./extension.ts"
import { readAll } from "https://deno.land/std@0.201.0/streams/read_all.ts";

export default async function exec(extension: Extension) {
    if (Deno.args.length == 0) {
        console.log(JSON.stringify(extension.toJSON()))
        return
    }

    const command = Deno.args[0]
    const input = JSON.parse(new TextDecoder().decode(await readAll(Deno.stdin)))
    const output = await extension.run(command, input)
    console.log(JSON.stringify(output))
}
