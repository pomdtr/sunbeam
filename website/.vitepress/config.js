import path from "path";
import fs from "fs";
import url from "url";

const __dirname = url.fileURLToPath(new URL(".", import.meta.url));

const cmdDir = path.join(__dirname, "..", "src", "cmd");
const cmdItems = fs.readdirSync(cmdDir).map((filename) => {
  const { name } = path.parse(filename);
  return {
    text: name.replaceAll("_", " "),
    link: `/cmd/${name}`,
  };
});

/**
 * @type {import('vitepress').DefaultTheme.Sidebar}
 */
const sidebar = [
  {
    items: [
      { text: "Introduction", link: "/user-guide/" },
      { text: "Installation", link: "/user-guide/installation" },
    ],
  },
  {
    text: "User Guide",
    items: [
      {
        text: "Configuration",
        link: "/user-guide/configuration",
      },
      {
        text: "Managing Extensions",
        link: "/user-guide/managing-extensions",
      },
    ],
  },
  {
    text: "Developer Guide",
    items: [
      {
        text: "Extension Manifest",
        link: "/developer-guide/extension-manifest",
      },
      {
        text: "Pages",
        items: [
          { text: "List", link: "/developer-guide/pages/list" },
          {
            text: "Detail",
            link: "/developer-guide/pages/detail",
          },
          {
            text: "Form",
            link: "/developer-guide/pages/form",
          },
        ],
      },
      {
        text: "Examples",
        items: [
          {
            text: "File Browser",
            link: "/developer-guide/examples/file-browser",
          },
        ],
      },
    ],
  },
  {
    text: "Command Line Usage",
    collapsed: true,
    collapsible: true,
    items: cmdItems,
  },
];

/**
 * @type {import('vitepress').DefaultTheme.Config}
 */
const themeConfig = {
  nav: [{ text: "Docs", link: "/user-guide/" }],
  logo: "/logo.svg",
  outline: [2, 3],
  sidebar,
  socialLinks: [{ icon: "github", link: "https://github.com/pomdtr/sunbeam" }],
  editLink: {
    pattern: "https://github.com/pomdtr/sunbeam/edit/main/website/src/:path",
    text: "Edit this page on GitHub",
  },
};

/**
 * @type {import('vitepress').UserConfig}
 */
const config = {
  title: "Sunbeam",
  head: [["link", { rel: "icon", href: "/sunbeam/logo.svg" }]],
  description:
    "Generate complex UIs from simple scripts written in any language.",
  base: "/sunbeam/",
  cleanUrls: "without-subfolders",
  srcDir: "./src",
  themeConfig,
};

export default config;
