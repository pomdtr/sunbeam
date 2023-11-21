const { screen } = require("electron");

function getCenterOnCurrentScreen(width, height) {
  const cursor = screen.getCursorScreenPoint();
  // Get display with cursor
  const distScreen = screen.getDisplayNearestPoint({
    x: cursor.x,
    y: cursor.y
  });

  const { width: screenWidth, height: screenHeight } = distScreen.workAreaSize;
  const x = distScreen.workArea.x + Math.floor(screenWidth / 2 - width / 2); // * distScreen.scaleFactor
  const y = distScreen.workArea.y + Math.floor(screenHeight / 3 - height / 2);

  return {
    width,
    height,
    x,
    y
  };
}

module.exports = { getCenterOnCurrentScreen };
