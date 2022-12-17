import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import { CanvasAddon } from "xterm-addon-canvas";
import { AttachAddon } from "xterm-addon-attach";

async function main() {
  const res = await fetch("/theme.json");
  const theme = await res.json();

  const terminal = new Terminal({
    cursorBlink: true,
    allowTransparency: true,
    macOptionIsMeta: true,
    fontSize: 13,
    scrollback: 0,
    fontFamily: "Consolas,Liberation Mono,Menlo,Courier,monospace",
    theme: theme,
  });

  const ws = new WebSocket(`ws://${location.host}/ws`);

  const fitAddon = new FitAddon();
  const canvasAddon = new CanvasAddon();
  const attachAddon = new AttachAddon(ws);

  terminal.open(document.getElementById("terminal")!);

  terminal.loadAddon(fitAddon);
  terminal.loadAddon(canvasAddon);
  terminal.loadAddon(attachAddon);

  terminal.focus();

  ws.onopen = () => {
    const textEncoder = new TextEncoder();
    terminal.onResize(({ cols, rows }) => {
      const payload = JSON.stringify({ cols, rows });
      const encodedPayload = textEncoder.encode(payload);
      ws.send(encodedPayload);
    });
    fitAddon.fit();
  };

  window.onresize = () => {
    fitAddon.fit();
  };
}

main();
