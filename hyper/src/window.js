function isFocused(w) {
  return w === this.getLastFocusedWindow();
}

function hideWindows(app) {
  const visibleWindows = [...app.getWindows()].filter(w => w.isVisible());

  if (!visibleWindows.length) {
    return false;
  }

  visibleWindows.sort(isFocused.bind(app)).forEach(w => {
    if (w.isFullScreen()) {
      return;
    }

    if (typeof w.hide === "function") {
      w.hide();
    } else {
      w.minimize();
    }
  });

  if (typeof app.hide === "function") {
    app.hide();
  }
}

function showWindows(app) {
  const windows = [...app.getWindows()].sort(isFocused.bind(app));

  windows.length === 0 ? app.createWindow() : windows.forEach(w => w.show());

  if (typeof app.show === "function") {
    app.show();
  }
}

function toggleWindows(app) {
  const focusedWindows = [...app.getWindows()].filter(
    w => w.isFocused() && w.isVisible()
  );

  focusedWindows.length > 0 ? hideWindows(app) : showWindows(app);
}

module.exports = {
  toggleWindows,
  hideWindows,
  showWindows
};
