import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  base: '/fxjson/',
  title: "FxJSON",
  description: "FxJSON — 高性能Go JSON解析库",
  lang: 'zh-CN',
  ignoreDeadLinks: true,
  head: [
    ['meta', { name: 'keywords', content: 'Go, JSON, 解析, 高性能, 零分配, FxJSON' }],
    ['meta', { name: 'author', content: 'iCloudZa' }],
    // ['link', { rel: 'icon', href: '' }]
  ],
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    // logo: '',
    
    nav: [
      { text: '首页', link: '/' },
      { text: '指南', link: '/guide/quick-start' },
      { text: 'API', link: '/api/' },
      { text: '示例', link: '/examples/' },
      { text: '性能', link: '/performance/' },
      { 
        text: '相关链接',
        items: [
          { text: 'GitHub', link: 'https://github.com/icloudza/fxjson' },
          { text: 'Go Packages', link: 'https://pkg.go.dev/github.com/icloudza/fxjson' },
          { text: '问题反馈', link: 'https://github.com/icloudza/fxjson/issues' }
        ]
      }
    ],

    sidebar: {
      '/guide/': [
        {
          text: '开始使用',
          items: [
            { text: '快速开始', link: '/guide/quick-start' },
            { text: '安装配置', link: '/guide/installation' },
            { text: '基础概念', link: '/guide/concepts' }
          ]
        },
        {
          text: '核心功能',
          items: [
            { text: '序列化与反序列化', link: '/guide/serialization' },
            { text: '数据验证', link: '/guide/validation' },
            { text: '查询和聚合', link: '/guide/query-aggregation' }
          ]
        },
        {
          text: '高级特性',
          items: [
            { text: '批量操作', link: '/guide/batch-operations' },
            { text: '工具函数', link: '/guide/utility-functions' }
          ]
        }
      ],
      '/api/': [
        {
          text: '核心类型',
          items: [
            { text: 'Node', link: '/api/#node' },
            { text: 'NodeType', link: '/api/#nodetype' }
          ]
        },
        {
          text: '包级函数',
          items: [
            { text: '解析函数', link: '/api/#解析函数' },
            { text: '序列化函数', link: '/api/#序列化函数' },
            { text: '解码函数', link: '/api/#解码函数' }
          ]
        },
        {
          text: 'Node 方法',
          collapsed: false,
          items: [
            { text: '访问方法', link: '/api/#访问方法' },
            { text: '类型转换', link: '/api/#类型转换方法' },
            { text: '便捷方法', link: '/api/#便捷方法带默认值' },
            { text: '类型检查', link: '/api/#类型检查方法' },
            { text: '遍历方法', link: '/api/#遍历方法' },
            { text: '数据转换', link: '/api/#数据转换方法' },
            { text: '数据验证', link: '/api/#数据验证方法' },
            { text: '类型信息', link: '/api/#类型信息方法' },
            { text: '原始数据', link: '/api/#原始数据方法' },
            { text: 'JSON序列化', link: '/api/#json序列化方法' },
            { text: '字符串操作', link: '/api/#字符串操作方法' },
            { text: '数组操作', link: '/api/#数组操作方法' },
            { text: '对象操作', link: '/api/#对象操作方法' },
            { text: '批量获取', link: '/api/#批量获取方法' },
            { text: '查找和过滤', link: '/api/#查找和过滤方法' },
            { text: '统计和分析', link: '/api/#统计和分析方法' },
            { text: '比较和状态检查', link: '/api/#比较和状态检查方法' },
            { text: '数字操作', link: '/api/#数字操作方法' },
            { text: '高级查询', link: '/api/#高级查询方法' },
            { text: '其他方法', link: '/api/#其他方法' }
          ]
        },
        {
          text: '配置选项',
          items: [
            { text: 'ParseOptions', link: '/api/#parseoptions' },
            { text: 'SerializeOptions', link: '/api/#serializeoptions' }
          ]
        },
        {
          text: '示例',
          items: [
            { text: '完整使用示例', link: '/api/#完整使用示例' },
            { text: '性能相关', link: '/api/#性能相关' },
            { text: '使用建议', link: '/api/#使用建议' }
          ]
        }
      ],
      '/examples/': [
        {
          text: '示例导航',
          items: [
            { text: '快速开始', link: '/examples/' }
          ]
        },
        
      ],
      '/performance/': [
        {
          text: '性能文档',
          items: [
            { text: '性能对比', link: '/performance/' }
          ]
        }
      ]
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/icloudza/fxjson' }
    ],

    footer: {
      message: '基于 MIT 许可证发布',
      copyright: 'Copyright © 2024 iCloudZa'
    },

    search: {
      provider: 'local',
      options: {
        translations: {
          button: {
            buttonText: '搜索文档',
            buttonAriaLabel: '搜索文档'
          },
          modal: {
            noResultsText: '无法找到相关结果',
            resetButtonTitle: '清除查询条件',
            footer: {
              selectText: '选择',
              navigateText: '切换'
            }
          }
        }
      }
    },

    editLink: {
      pattern: 'https://github.com/icloudza/fxjson-docs/edit/main/:path',
      text: '在 GitHub 上编辑此页'
    },

    lastUpdated: {
      text: '最后更新于',
      formatOptions: {
        dateStyle: 'short',
        timeStyle: 'medium'
      }
    },

    docFooter: {
      prev: '上一页',
      next: '下一页'
    },

    outline: {
      label: '页面导航'
    },

    returnToTopLabel: '回到顶部',
    sidebarMenuLabel: '菜单',
    darkModeSwitchLabel: '主题',
    lightModeSwitchTitle: '切换到浅色模式',
    darkModeSwitchTitle: '切换到深色模式'
  }
})
