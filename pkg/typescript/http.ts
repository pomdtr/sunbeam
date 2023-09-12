import { Extension } from "./extension.ts";

export function handler(extension: Extension) {
  return async (req: Request) => {
    if (req.method == "GET") {
      return Response.json(extension);
    }

    if (req.method != "POST") {
      return Response.json({ "message": "Method not allowed" }, {
        status: 405,
      });
    }

    const url = new URL(req.url);
    const [, command] = url.pathname.split("/");
    const input = await req.json();
    const output = await extension.run(command, input);
    return Response.json(output);
  };
}

export function handle(extension: Extension, req: Request) {
  return handler(extension)(req);
}
