// @ts-check

// @ts-ignore
const { globalShortcut, app, Tray, Menu, shell } = require("electron");
const { toggleWindows, hideWindows, showWindows } = require("./window");
const { getCenterOnCurrentScreen } = require("./screen");
const path = require("path");
const os = require("os");

let unload = () => { }

function onApp(app) {
  if (!app.requestSingleInstanceLock()) {
    app.quit()
    return
  }

  app.on('second-instance', () => {
    // Someone tried to run a second instance, we should focus our window.
    toggleWindows(app)
  })

  // Prevent the app from quitting when all windows are closed
  app.removeAllListeners('window-all-closed')
  app.on('window-all-closed', (event) => {
    // empty callback to prevent the default behavior
  })

  // Hide the dock icon
  if (process.platform == "darwin") {
    app.dock.hide();
  }

  // Create tray icon
  const iconPath = process.platform == "darwin" ? "../assets/trayiconTemplate.png" : "../assets/trayicon.png";
  const tray = new Tray(path.join(__dirname, iconPath));
  const contextMenu = Menu.buildFromTemplate([
    {
      label: 'Show Sunbeam',
      click: () => {
        showWindows(app);
      },
    },
    { type: "separator" },
    {
      label: 'Edit Sunbeam Config',
      click: () => {
        shell.openPath(path.join(os.homedir(), '.config', 'sunbeam', 'sunbeam.json'))
      },
    },
    {
      label: 'Edit Hyper Config',
      click: () => {
        shell.openPath(path.join(os.homedir(), '.hyper.js'))
      },
    },
    { type: 'separator' },
    {
      label: 'Browse Documentation',
      click: () => {
        shell.openExternal('https://sunbeam.deno.dev/docs');
      },
    },
    {
      label: 'Open Github Repository',
      click: () => {
        shell.openExternal('https://github.com/pomdtr/sunbeam');
      }
    },
    { type: 'separator' },
    {
      label: 'Quit',
      click: () => {
        app.quit();
      },
    },
  ]);
  tray.setToolTip('Sunbeam');
  tray.setContextMenu(contextMenu);

  // Hide windows when the app looses focus
  const onBlur = () => {
    hideWindows(app);
  }
  app.on("browser-window-blur", onBlur);

  unload = () => {
    if (tray) tray.destroy();
    globalShortcut.unregisterAll();
    app.removeListener("browser-window-blur", onBlur);
  };
};

function onWindow(win) {
  win.on("close", () => {
    if (process.platform == "darwin") {
      app.hide();
    }
  });
}


function onUnload() {
  unload();
}

// Hide window controls on macOS
function decorateBrowserOptions(defaults) {
  const bounds = getCenterOnCurrentScreen(defaults.width, defaults.height);
  return Object.assign({}, defaults, {
    ...bounds,
    titleBarStyle: '',
    transparent: true,
    frame: false,
    alwaysOnTop: true,
    type: "panel",
    skipTaskbar: true,
    movable: false,
    fullscreenable: false,
    minimizable: false,
    maximizable: false,
    resizable: false
  });
};


function decorateConfig(config) {
  globalShortcut.unregisterAll();

  if (config.hyperSunbeam && config.hyperSunbeam.hotkey) {
    const hotkey = config.hyperSunbeam.hotkey;
    globalShortcut.register(hotkey, () => toggleWindows(app));
  }

  const css = `
    .header_header {
      top: 0;
      right: 0;
      left: 0;
      visibility: hidden;
    }
    .tabs_borderShim {
      display: none;
    }
    .tabs_title {
      display: none;
    }
    .tabs_nav {
      height: auto;
    }
    .tabs_list {
      margin-left: 0;
    }
    .tab_tab:first-of-type {
      border-left-width: 0;
      padding-left: 1px;
    }

    .terms_terms {
      margin-top: 0;
    }
  `
  return Object.assign({}, config, {
    css: `
      ${config.css || ''}
      ${css}
    `
  });
}

// Adding Keymaps
function decorateKeymaps(keymaps) {
  return Object.assign({}, keymaps, {
    "tab:new": "",
    "window:new": "",
    "editor:clearBuffer": "",
  });
}

module.exports = {
  onApp,
  onWindow,
  onUnload,
  decorateBrowserOptions,
  decorateKeymaps,
  decorateConfig,
};
