import { spawn } from "child_process";
import {
  app,
  BrowserWindow,
  globalShortcut,
  screen,
  ipcMain,
  shell,
  clipboard,
} from "electron";
import path from "path";
import axios from "axios";
import url from "url";
import minimist from "minimist";

const args = minimist(process.argv.slice(2));

function createWindow() {
  const bounds = getCenterOnCurrentScreen();
  const win = new BrowserWindow({
    title: "Sunbeam",
    width: 750,
    height: 475,
    frame: false,
    x: bounds.x,
    y: bounds.y,
    alwaysOnTop: true,
    skipTaskbar: true,
    minimizable: false,
    transparent: true,
    maximizable: false,
    fullscreenable: false,
    movable: false,
    autoHideMenuBar: true,
    webPreferences: {
      preload: path.join(__dirname, "preload.js"),
      backgroundThrottling: false,
      spellcheck: false,
    },
    resizable: false,
    type: "panel",
    show: true,
    hasShadow: true,
  });
  win.setMenu(null);
  win.loadURL("http://localhost:8080");
  ipcMain.handle("hideWindow", () => win.hide());
  ipcMain.handle("showWindow", () => win.show());

  ipcMain.handle("copyToClipboard", (_: unknown, text: string) => {
    clipboard.writeText(text);
  });
  ipcMain.handle("openInBrowser", (_: unknown, url: string) => {
    shell.openExternal(url);
  });

  win.webContents.setWindowOpenHandler(({ url }) => {
    shell.openExternal(url);
    return { action: "deny" };
  });

  win.on("blur", () => {
    win.hide();
  });

  return win;
}

const getCenterOnCurrentScreen = () => {
  const cursor = screen.getCursorScreenPoint();
  // Get display with cursor
  const distScreen = screen.getDisplayNearestPoint({
    x: cursor.x,
    y: cursor.y,
  });

  const { width: screenWidth, height: screenHeight } = distScreen.workAreaSize;
  const width = 750;
  const height = 475;
  const x = distScreen.workArea.x + Math.floor(screenWidth / 2 - width / 2); // * distScreen.scaleFactor
  const y = distScreen.workArea.y + Math.floor(screenHeight / 2 - height / 2);

  return {
    width,
    height,
    x,
    y,
  };
};

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

app.on("window-all-closed", () => {
  // pass
});
app.setAsDefaultProtocolClient("sunbeam");
if (process.platform === "darwin") {
  app.dock.hide();
}

app.whenReady().then(async () => {
  const sunbeam = app.isPackaged
    ? path.join(process.resourcesPath, "sunbeam")
    : "sunbeam";
  let [host, port]: [string, number] = ["localhost", 8080];
  if (args.host && args.port) {
    host = args.host;
    port = args.port;
  } else {
    console.log(`Starting sunbeam at http://${host}:${port}`);
    const shell = process.env.SHELL;
    const command = `${sunbeam} serve --host ${host} --port ${port}`;
    const server = spawn(shell, ["-c", command], {
      env: {
        TERM: "xterm-256color",
        PATH: `${process.resourcesPath}:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin`,
      },
    });

    app.on("before-quit", () => {
      server.kill();
    });
  }
  let ready = false;
  while (!ready) {
    await sleep(500);
    try {
      const res = await axios.get(`http://${host}:${port}/ready`);
      if (res.status === 200) {
        ready = true;
      }
    } catch (e) {
      console.log("Sunbeam not ready yet...");
    }
  }

  const win = createWindow();
  win.webContents.on("dom-ready", () => {
    globalShortcut.register("CommandOrControl+;", async () => {
      if (win.isVisible()) {
        win.hide();
      } else {
        const bounds = getCenterOnCurrentScreen();
        if (JSON.stringify(bounds) !== JSON.stringify(win.getBounds())) {
          win.setBounds(bounds);
          await sleep(50);
        }
        win.show();
      }
    });
  });

  app.on("open-url", (_: unknown, sunbeamUrl: string) => {
    const parsedUrl = url.parse(sunbeamUrl);
    switch (parsedUrl.host) {
      case "run":
        win.loadURL(`http://${host}:${port}/run${parsedUrl.path}`);
    }
  });
});
