import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: 'Sunbeam',
  base: '/sunbeam/',
  head: [
    ['link', { rel: 'icon', href: '/assets/favicon.png' }]
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
                link: "/developer-guide/shell"
              },
              {
                text: "Hackernews (Typescript)",
                link: "/developer-guide/typescript"
              },
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
