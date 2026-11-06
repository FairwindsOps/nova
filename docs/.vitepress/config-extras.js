// To see all options:
// https://vuepress.vuejs.org/config/
// https://vuepress.vuejs.org/theme/default-theme-config.html
module.exports = {
  title: "Nova Documentation",
  description: "Documentation for Fairwinds' Nova",
  themeConfig: {
    docsRepo: "FairwindsOps/Nova",
    sidebar: [
      {
        title: "Nova",
        path: "/",
        sidebarDepth: 0,
      },
      {
        title: "Installation",
        path: "/installation",
      },
      {
        title: "Quickstart",
        path: "/quickstart",
      },
      {
        title: "Usage",
        path: "/usage",
      },
      {
        title: "Desired Versions",
        path: "/desired-versions",
      },
      {
        title: "Contributing",
        children: [
          {
            title: "Guide",
            path: "contributing/guide"
          },
          {
            title: "Code of Conduct",
            path: "contributing/code-of-conduct"
          }
        ]
      }
    ]
  },
}
