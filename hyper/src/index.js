// @ts-check

// @ts-ignore
const { globalShortcut, dialog, app, Tray, Menu, shell } = require("electron");
const { toggleWindows, hideWindows, showWindows } = require("./window");
const { getCenterOnCurrentScreen } = require("./screen");
const path = require("path");
const os = require("os");

let unload = () => { }

function onApp(app) {
  const { hotkey } = Object.assign(
    { hotkey: "Ctrl+;" },
    app.config.getConfig().sunbeam
  );

  app.dock.hide();

  const tray = new Tray(path.join(__dirname, "../assets/trayiconTemplate.png"));
  const contextMenu = Menu.buildFromTemplate([
    {
      label: 'Show Sunbeam',
      click: () => {
        showWindows(app);
      },
      accelerator: hotkey,
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

  const onActivate = () => {
    showWindows(app);
  }
  app.on("activate", onActivate);

  const onBlur = () => {
    hideWindows(app);
  }
  app.on("browser-window-blur", onBlur);

  if (!hotkey) return;
  globalShortcut.unregister(hotkey);
  if (!globalShortcut.register(hotkey, () => toggleWindows(app))) {
    dialog.showMessageBox({
      message: `Could not register hotkey (${hotkey})`,
      buttons: ["Ok"]
    });
  }

  unload = () => {
    app.dock.show();
    app.removeListener("activate", onActivate);
    app.removeListener("browser-window-blur", onBlur);
    globalShortcut.unregister(hotkey);
  };
};

function onWindow(win) {
  win.on("close", () => {
    app.hide()
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
  const macosCSS = `
    .header_header {
      top: 0;
      right: 0;
      left: 0;
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
  `
  const defaultCSS = `
    .header_windowHeader {
      display: none;
    }
    .tabs_nav {
      top: 0;
    }
    .tabs_list {
      padding-left: 0;
    }
    .tabs_list:before {
      display: none;
    }
    .tab_first {
      border-left-width: 0;
    }
    .terms_terms {
      margin-top: 0;
    }
    .terms_termsShifted {
      margin-top: 34px;
    }
    .tab_tab:after {
      display: none;
    }
  `

  return Object.assign({}, config, {
    css: `
      ${config.css || ''}
      ${process.platform === 'darwin' ? macosCSS : defaultCSS}
    `
  });
}

// Removes the redundant space on mac if there is only one tab
function getTabsProps(parentProps, props) {
  if (process.platform === 'darwin') {
    var classTermsList = document.getElementsByClassName('terms_terms')
    if (classTermsList.length > 0) {
      var classTerms = classTermsList[0]
      var header = document.getElementsByClassName('header_header')[0]
      if (props.tabs.length <= 1) {
        // @ts-ignore
        classTerms.style.marginTop = 0
        // @ts-ignore
        header.style.visibility = 'hidden'
      } else {
        // @ts-ignore
        classTerms.style.marginTop = ''
        // @ts-ignore
        header.style.visibility = ''
      }
    }
  }
  return Object.assign({}, parentProps, props)
}

module.exports = {
  onApp,
  onWindow,
  onUnload,
  decorateBrowserOptions,
  getTabsProps,
  decorateConfig,
};
