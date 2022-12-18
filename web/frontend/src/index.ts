import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import { CanvasAddon } from "xterm-addon-canvas";
import { AttachAddon } from "xterm-addon-attach";
import defaultTheme from "./theme.json";

async function loadTheme() {
  const res = await fetch("/theme.json");
  if (res.ok) {
    return res.json();
  } else {
    const error = await res.text();
    console.error("Failed to load theme", error);
    return defaultTheme;
  }
}

async function main() {
  const theme = await loadTheme();

  const terminal = new Terminal({
    cursorBlink: true,
    allowTransparency: true,
    macOptionIsMeta: true,
    fontSize: 13,
    scrollback: 0,
    fontFamily: "Consolas,Liberation Mono,Menlo,Courier,monospace",
    theme,
  });

  const protocol = location.protocol === "https:" ? "wss:" : "ws:";
  const ws = new WebSocket(
    `${protocol}//${location.host}/ws${location.search}`
  );

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

  ws.onclose = () => {
    // @ts-ignore
    window.location = "/";
  };

  window.onresize = () => {
    fitAddon.fit();
  };
}

main();
