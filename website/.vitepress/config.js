import path from "path";
import fs from "fs";
import url from "url";

const __dirname = url.fileURLToPath(new URL(".", import.meta.url));

const cmdDir = path.join(__dirname, "..", "cmd");
const cmdItems = fs.readdirSync(cmdDir).map((filename) => {
  const { name } = path.parse(filename);
  return {
    text: name.replaceAll("_", " "),
    link: `/cmd/${name}`,
  };
});

/**
 * @type {import('vitepress').DefaultTheme.Config}
 */
const themeConfig = {
  nav: [{ text: "Docs", link: "/guide/" }],
  sidebar: [
    {
      text: "User Guide",
      items: [
        { text: "Introduction", link: "/guide/" },
        { text: "Installation", link: "/guide/installation" },
        {
          text: "Managing Extensions",
          link: "/user-guide/managing-extensions",
        },
      ],
    },

    {
      text: "Command Line Usage",
      collapsed: true,
      collapsible: true,
      items: cmdItems,
    },
  ],
  socialLinks: [{ icon: "github", link: "https://github.com/pomdtr/sunbeam" }],
  editLink: {
    pattern: "https://github.com/pomdtr/sunbeam/edit/main/website/:path",
    text: "Edit this page on GitHub",
  },
};

/**
 * @type {import('vitepress').UserConfig}
 */
const config = {
  title: "Sunbeam",
  description:
    "Generate complex UIs from simple scripts written in any language.",
  base: "/sunbeam/",
  cleanUrls: "without-subfolders",
  themeConfig,
};

export default config;
