import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import { CanvasAddon } from "xterm-addon-canvas";
import { AttachAddon } from "xterm-addon-attach";
import { WebLinksAddon } from "xterm-addon-web-links";
import defaultTheme from "./theme.json";

declare global {
  interface Window {
    electron?: {
      showWindow: () => Promise<void>;
      hideWindow: () => Promise<void>;
      openInBrowser: (url: string) => Promise<void>;
      open: (path: string) => Promise<void>;
      copyToClipboard: (text: string) => Promise<void>;
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

async function reloadPage() {
  if (window.electron) {
    // @ts-ignore
    window.location = window.location.pathname;
  } else {
    location.reload();
  }
}

async function copyText(text: string) {
  if (window.electron) {
    await window.electron.copyToClipboard(text);
  } else {
    await navigator.clipboard.writeText(text);
  }
}

async function openPath(path: string) {
  if (window.electron) {
    await window.electron.open(path);
  } else {
    console.error("Cannot open files on remote server");
  }
}

async function openUrl(url: string) {
  if (window.electron) {
    await window.electron.openInBrowser(url);
  } else {
    window.open(url, "_blank");
  }
}

async function main() {
  const theme = await loadTheme();

  const terminal = new Terminal({
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

  let ready = false;
  ws.onmessage = async (event) => {
    // Show window after run command
    if (!ready) {
      ready = true;
      const params = new URLSearchParams(location.search);
      if (window.electron && params.get("extension")) {
        window.electron.showWindow();
      }
    }

    const data = event.data;
    if (typeof data === "string") {
      if (window.electron) {
        await window.electron.hideWindow();
      }
      const msg = JSON.parse(data);
      switch (msg.action) {
        case "open-url":
          await openUrl(msg.url);
          break;
        case "copy-text":
          await copyText(msg.text);
          break;
        case "open-path":
          await openPath(msg.path);
          break;
        case "exit":
          break;
        default:
          await copyText(`Unknown action: ${msg.action}`);
          break;
      }
      reloadPage();
    }
  };

  ws.onclose = async () => {
    reloadPage();
  };

  window.onresize = () => {
    fitAddon.fit();
  };
}

main();
