import { spawn } from 'node:child_process'
import { Buffer } from 'node:buffer'

export async function editor(options = {
    extension: 'txt',
    content: '',
}) {
    let extension = options.extension || "txt";
    if (extension.startsWith(".")) {
        extension = extension.slice(1);
    }

    const command = spawn("sunbeam", ["edit", "--extension", extension], {
        stdio: ['pipe', 'pipe', 'pipe'],
    });

    const writer = command.stdin;
    writer.write(Buffer.from(options.content || ""));
    writer.end();

    const stdoutPromise = new Promise((resolve) => {
        let output = '';
        command.stdout.on('data', (data) => {
            output += data.toString();
        });

        command.stdout.on('end', () => {
            resolve(output);
        });
    });

    const output = await stdoutPromise;


    return output;
}
