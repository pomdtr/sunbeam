import { toJson } from "https://deno.land/std@0.203.0/streams/mod.ts";
import { CommandInput, Extension } from "./extension.ts";

export async function exec(extension: Extension) {
  if (Deno.args.length == 0) {
    console.log(JSON.stringify(extension, null, 2));
  }

  const command = Deno.args[0];
  const input = await toJson(Deno.stdin.readable) as CommandInput;
  const res = extension.run(command, input);
  console.log(JSON.stringify(res, null, 2));
}
