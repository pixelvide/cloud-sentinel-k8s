import { defineConfig } from "vitepress";

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "Cloud Sentinel K8s",
  description: "A modern, intuitive Kubernetes dashboard",

  base: "/cloud-sentinel-k8s/",
  sitemap: {
    hostname: "https://pixelvide.github.io/cloud-sentinel-k8s/",
    lastmodDateOnly: false,
  },

  markdown: {
    image: {
      lazyLoading: true,
    },
  },

  lastUpdated: true,
  locales: {
    root: {
      label: "English",
      lang: "en",
    },

  },

  head: [
    ["link", { rel: "icon", href: "/logo.svg" }],

  ],

  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    logo: "/logo.svg",
    search: {
      provider: "local",
    },
    langMenuLabel: "Language",
    editLink: {
      pattern: "https://github.com/pixelvide/cloud-sentinel-k8s/tree/main/docs/:path",
      text: "Edit this page on GitHub",
    },

    nav: [
      { text: "Home", link: "/" },
      { text: "Guide", link: "/guide/" },
      { text: "Configuration", link: "/config/" },
      { text: "FAQ", link: "/faq" },
    ],

    sidebar: {
      "/": [
        {
          text: "Introduction",
          items: [
            { text: "What is Cloud Sentinel K8s?", link: "/guide/" },
            { text: "Getting Started", link: "/guide/installation" },
          ],
        },
        {
          text: "Configuration",
          items: [
            { text: "User Management", link: "/config/user-management" },
            { text: "OAuth Setup", link: "/config/oauth-setup" },
            { text: "RBAC Configuration", link: "/config/rbac-config" },
            { text: "Prometheus Setup", link: "/config/prometheus-setup" },
            { text: "Managed K8s Auth", link: "/config/managed-k8s-auth" },
            { text: "Environment Variables", link: "/config/env" },
            { text: "Chart Values", link: "/config/chart-values" },
          ],
        },
        {
          text: "Usage",
          items: [
            { text: "Global Search", link: "/guide/global-search" },
            { text: "Resource Management", link: "/guide/resource-management" },
            { text: "Helm Management", link: "/guide/helm" },
            { text: "Related Resources", link: "/guide/related-resources" },
            { text: "Logs", link: "/guide/logs" },
            { text: "Monitor", link: "/guide/monitoring" },
            { text: "Web Terminal", link: "/guide/web-terminal" },
            { text: "Resource History", link: "/guide/resource-history" },
            { text: "Custom Sidebar", link: "/guide/custom-sidebar" },
            { text: "Kube Proxy", link: "/guide/kube-proxy" },
          ],
        },
        {
          text: "FAQ",
          link: "/faq",
        },
      ],

    },

    socialLinks: [{ icon: "github", link: "https://github.com/pixelvide/cloud-sentinel-k8s" }],

    footer: {
      message: "Released under the Apache License.",
      copyright: "Copyright © 2026-present Cloud Sentinel K8s Contributors",
    },
  },
});
