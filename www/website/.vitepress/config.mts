import { defineConfig } from "vitepress";

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "Sunbeam",
  head: [
    ["link", { rel: "icon", href: "/favicon.png" }],
    ["meta", { property: "og:type", content: "website" }],
    [
      "meta",
      {
        property: "og:image",
        content: "https://pomdtr.github.io/sunbeam/screenshot.png",
      },
    ],
    [
      "meta",
      {
        property: "og:description",
        content: "Wrap your tools in keyboard-friendly TUIs",
      },
    ],
  ],
  appearance: "force-dark",
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    outline: [2, 3],
    nav: [
      { text: "Home", link: "/" },
      {
        text: "Docs",
        link: "/docs/",
      },
      {
        text: "Extension Catalog",
        link: "/catalog/",
      },
    ],
    search: {
      provider: "local",
    },
    sidebar: [
      {
        text: "Introduction",
        link: "/docs/",
      },
      {
        text: "User Guide",
        items: [
          {
            text: "Installation",
            link: "/docs/user-guide/installation",
          },
          {
            text: "Quick Start",
            link: "/docs/user-guide/quickstart",
          },
          {
            text: "Integrations",
            link: "/docs/user-guide/integrations",
          },
        ],
      },
      {
        text: "Developer Guide",
        items: [
          {
            text: "Guidelines",
            link: "/docs/developer-guide/guidelines",
          },
          {
            text: "Examples",
            items: [
              {
                text: "DevDocs (Shell)",
                link: "/docs/developer-guide/examples/devdocs",
              },
              {
                text: "Hackernews (Typescript)",
                link: "/docs/developer-guide/examples/hackernews",
              },
              {
                text: "File Browser (Python)",
                link: "/docs/developer-guide/examples/file-browser",
              },
              {
                text: "Google Search (Shell)",
                link: "/docs/developer-guide/examples/google-search",
              },
            ],
          },
          {
            text: "Publishing",
            link: "/docs/developer-guide/publishing",
          },
          {
            text: "Tips",
            link: "/docs/developer-guide/tips",
          },
        ],
      },
      {
        text: "Reference",
        items: [
          {
            text: "Configuration",
            link: "/docs/reference/config",
          },
          {
            text: "Schemas",
            collapsed: true,
            items: [
              {
                text: "Manifest",
                link: "/docs/reference/schemas/manifest",
              },
              {
                text: "Payload",
                link: "/docs/reference/schemas/payload",
              },
              {
                text: "List",
                link: "/docs/reference/schemas/list",
              },
              {
                text: "Detail",
                link: "/docs/reference/schemas/detail",
              },
              {
                text: "Action",
                link: "/docs/reference/schemas/action",
              },
            ],
          },
          {
            text: "CLI",
            link: "/docs/reference/cli",
          },
        ],
      },
      {
        text: "Alternatives",
        link: "/docs/alternatives",
      },
    ],
    socialLinks: [
      { icon: "github", link: "https://github.com/pomdtr/sunbeam" },
    ],
  },
});
