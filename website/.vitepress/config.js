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
        text: "Reference",
        items: [
          { text: "Page", link: "/developer-guide/reference/page.md" },
          {
            text: "Actions",
            link: "/developer-guide/reference/actions.md",
          },
          {
            text: "Inputs",
            link: "/developer-guide/reference/inputs.md",
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
  algolia: {
    appId: "OGGDU8PMQA",
    apiKey: "fd31a6a190c9dd3907611922cb46759c",
    indexName: "sunbeam",
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
