const path = require("path");

const GOROOT = process.env.GOROOT || `${process.env.HOME}/go`;

module.exports = {
  packagerConfig: {
    extraResource: [path.join(GOROOT, "bin", "sunbeam")],
    icon: path.resolve(__dirname, "assets/icon.icns"),
    name: "Sunbeam",
    protocols: [
      {
        name: "Sunbeam",
        schemes: ["sunbeam"],
      },
    ],
  },
  rebuildConfig: {},
  makers: [
    {
      name: "@electron-forge/maker-squirrel",
      config: {},
    },
    {
      name: "@electron-forge/maker-zip",
      platforms: ["darwin"],
    },
    {
      name: "@electron-forge/maker-deb",
      config: {
        mimeType: ["x-scheme-handler/sunbeam"],
      },
    },
    {
      name: "@electron-forge/maker-rpm",
      config: {
        mimeType: ["x-scheme-handler/sunbeam"],
      },
    },
  ],
};
