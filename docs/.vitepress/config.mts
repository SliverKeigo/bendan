import { defineConfig } from 'vitepress'

export default defineConfig({
  lang: 'zh-CN',
  title: 'Bendan',
  description: '基于 NapCatQQ 与 OneBot v11 的 QQ 机器人',
  base: '/bendan/',
  cleanUrls: true,
  head: [
    ['meta', { name: 'theme-color', content: '#0b5d5e' }],
    ['link', { rel: 'icon', href: '/mark.svg', type: 'image/svg+xml' }],
  ],
  themeConfig: {
    logo: '/mark.svg',
    nav: [
      { text: '使用指南', link: '/guide/getting-started' },
      { text: '功能一览', link: '/guide/features' },
      { text: '配置部署', link: '/guide/configuration' },
      { text: 'GitHub', link: 'https://github.com/SliverKeigo/bendan' },
    ],
    sidebar: {
      '/guide/': [
        {
          text: '开始',
          items: [
            { text: '快速开始', link: '/guide/getting-started' },
            { text: '功能一览', link: '/guide/features' },
          ],
        },
        {
          text: '运维',
          items: [
            { text: '配置与部署', link: '/guide/configuration' },
            { text: '动作词表', link: '/guide/actions' },
            { text: '管理员命令', link: '/guide/administration' },
            { text: '状态查询', link: '/guide/status' },
          ],
        },
        {
          text: '参考',
          items: [
            { text: '运行与排查', link: '/guide/operations' },
          ],
        },
      ],
    },
    outline: { level: [2, 3], label: '本页内容' },
    search: { provider: 'local' },
    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Bendan Bot',
    },
  },
})
