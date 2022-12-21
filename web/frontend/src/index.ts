import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import { CanvasAddon } from "xterm-addon-canvas";
import { AttachAddon } from "xterm-addon-attach";
import { WebLinksAddon } from "xterm-addon-web-links";
import defaultTheme from "./theme.json";

declare global {
  interface Window {
    electron?: {
      windowHide: () => Promise<void>;
      openInBrowser: (url: string) => Promise<void>;
    };
  }
}
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
  const webLinksAddon = new WebLinksAddon((_, url) => {
    if (window.electron) {
      window.electron.openInBrowser(url);
    } else {
      window.open(url);
    }
  });

  terminal.open(document.getElementById("terminal")!);

  terminal.loadAddon(fitAddon);
  terminal.loadAddon(canvasAddon);
  terminal.loadAddon(attachAddon);
  terminal.loadAddon(webLinksAddon);

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

  ws.onmessage = async (event) => {
    const data = event.data;
    if (typeof data === "string") {
      // Copy to clipboard
      // await navigator.clipboard.writeText(data);
      // Open URL
      window.open(data);
    }
  };

  ws.onclose = async () => {
    if (window.electron) {
      await window.electron.windowHide();
      // @ts-ignore
      window.location = window.location.pathname;
    } else {
      location.reload();
    }
  };

  window.onresize = () => {
    fitAddon.fit();
  };
}

main();
