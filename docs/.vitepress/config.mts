import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: 'Sunbeam',
  base: '/sunbeam/',
  head: [
    ['link', { rel: 'icon', href: '/sunbeam/favicon.png' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:image', content: 'https://pomdtr.github.io/sunbeam/screenshot.png' }],
    ['meta', { property: 'og:description', content: 'Wrap your tools in keyboard-friendly TUIs' }]
  ],
  cleanUrls: true,
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    outline: [2, 3],
    nav: [
      { text: 'Home', link: '/' },
      {
        text: 'Docs',
        link: '/introduction'
      },
      {
        text: 'Extension Catalog',
        link: '/catalog'
      }
    ],
    search: {
      provider: 'local'
    },
    sidebar: [
      {
        text: 'Introduction',
        link: '/introduction',
      },
      {
        text: 'User Guide',
        items: [
          {
            text: "Installation",
            link: "/user-guide/installation",
          },
          {
            text: "Quick Start",
            link: "/user-guide/quickstart",
          },
          {
            text: "Integrations",
            link: "/user-guide/integrations",
          }
        ]
      },
      {
        text: 'Developer Guide',
        items: [
          {
            text: "Guidelines",
            link: "/developer-guide/guidelines"
          },
          {
            text: "Examples",
            items: [
              {
                text: "DevDocs (Shell)",
                link: "/developer-guide/examples/devdocs"
              },
              {
                text: "Hackernews (Typescript)",
                link: "/developer-guide/examples/hackernews"
              },
              {
                text: "File Browser (Python)",
                link: "/developer-guide/examples/file-browser"
              },
              {
                text: "Google Search (Shell)",
                link: "/developer-guide/examples/google-search"
              }
            ]
          },

          {
            text: "Publishing",
            link: "/developer-guide/publishing"
          }
        ]
      },
      {
        text: 'Reference',
        items: [
          {
            text: "Configuration",
            link: "/reference/config"
          },
          {
            text: "Schemas",
            collapsed: true,
            items: [
              {
                text: "Manifest",
                link: "/reference/schemas/manifest"
              },
              {
                text: "Payload",
                link: "/reference/schemas/payload"
              },
              {
                text: "List",
                link: "/reference/schemas/list"
              },
              {
                text: "Detail",
                link: "/reference/schemas/detail"
              },
              {
                text: "Action",
                link: "/reference/schemas/action"
              }
            ]
          },
          {
            text: "CLI",
            link: "/reference/cli"
          }
        ]
      },
      {
        text: 'Alternatives',
        link: '/alternatives'
      }

    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/pomdtr/sunbeam' },
    ]
  }
})
