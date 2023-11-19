export async function editor(extension: string, content?: string) {
    if (extension.startsWith(".")) {
        extension = extension.slice(1);
    }

    const command = new Deno.Command("sunbeam", {
        args: ["edit", "--extension", extension],
        stdin: "piped",
        stdout: "piped",
    })

    const process = await command.spawn();

    const writer = process.stdin.getWriter()
    writer.write(new TextEncoder().encode(content || ""));
    writer.releaseLock();

    await process.stdin.close();

    const { success, stdout } = await process.output();
    if (!success) {
        throw new Error("Editor failed");
    }

    return new TextDecoder().decode(stdout);
}
