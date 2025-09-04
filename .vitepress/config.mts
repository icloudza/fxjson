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
            { text: '5分钟快速上手', link: '/guide/quick-start' },
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
          text: '快速索引',
          items: [
            { text: '核心解析函数', link: '/api/#核心解析函数' },
            { text: '基础数据访问', link: '/api/#基础数据访问' },
            { text: '类型转换方法', link: '/api/#类型转换方法' },
            { text: '数组操作', link: '/api/#数组操作' },
            { text: '对象操作', link: '/api/#对象操作' },
            { text: '高级功能', link: '/api/#高级功能' }
          ]
        },
        {
          text: '解析和访问',
          items: [
            { text: 'FromString / FromBytes', link: '/api/#fromstring' },
            { text: 'Get / GetPath', link: '/api/#get' },
            { text: 'Index / Exists', link: '/api/#index' }
          ]
        },
        {
          text: '数据转换',
          items: [
            { text: '安全转换 (推荐)', link: '/api/#安全转换方法推荐' },
            { text: '严格转换', link: '/api/#严格转换方法' },
            { text: '切片转换', link: '/api/#toslice系列方法' }
          ]
        },
        {
          text: '遍历和检查',
          items: [
            { text: '数组遍历', link: '/api/#arrayforeach' },
            { text: '对象遍历', link: '/api/#foreach' },
            { text: '类型检查', link: '/api/#类型检查方法' },
            { text: '深度遍历', link: '/api/#walk' }
          ]
        },
        {
          text: '验证和序列化',
          items: [
            { text: '数据验证', link: '/api/#数据验证' },
            { text: '结构体操作', link: '/api/#结构体操作' },
            { text: '配置选项', link: '/api/#配置选项' }
          ]
        },
        {
          text: '使用指南',
          items: [
            { text: '性能特性', link: '/api/#性能特性' },
            { text: '错误处理', link: '/api/#错误处理指南' },
            { text: '最佳实践', link: '/api/#最佳实践' }
          ]
        }
      ],
      '/examples/': [
        {
          text: '基础示例',
          items: [
            { text: '用户信息解析', link: '/examples/#示例-1-用户信息解析' },
            { text: '嵌套数据访问', link: '/examples/#示例-2-嵌套数据访问' }
          ]
        },
        {
          text: '数组和对象',
          items: [
            { text: '商品列表处理', link: '/examples/#示例-3-商品列表处理' },
            { text: '成绩统计分析', link: '/examples/#示例-4-成绩统计分析' },
            { text: '部门员工统计', link: '/examples/#示例-5-部门员工统计' }
          ]
        },
        {
          text: 'API 处理',
          items: [
            { text: '分页数据处理', link: '/examples/#示例-6-分页数据处理' },
            { text: '错误响应处理', link: '/examples/#示例-7-错误响应处理' },
            { text: '配置文件解析', link: '/examples/#示例-8-应用配置管理' }
          ]
        },
        {
          text: '高级应用',
          items: [
            { text: '数据验证', link: '/examples/#示例-9-用户注册验证' },
            { text: '性能优化', link: '/examples/#示例-10-大数据处理优化' },
            { text: '微服务配置', link: '/examples/#示例-12-微服务配置中心' }
          ]
        }
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
